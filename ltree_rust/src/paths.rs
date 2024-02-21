use std::str::FromStr;

mod parsing;
pub use parsing::{LTreeParseResult, LTreePathParseError};

use self::parsing::validate_label;

#[cfg(test)]
mod paths_tests;

///
/// Represents a ltree path, as used in the PostgreSQL `ltree` extension.
///
/// *Note: This targets PostgreSQL 16+.*
///
/// # Usage
/// To create a path, use either [LTreePath::try_from] or [LTreePath::from_str]. The string will be
/// validated, and barring any locale issues, should be accepted by PostgreSQL.
///
/// ```rust
/// use ltree_rust::LTreePath;
/// use std::str::FromStr;
/// LTreePath::from_str("what.is.love").expect("the path is correct");
/// LTreePath::from_str("this.is.a#bad.path").expect_err("the path is not correct");
/// ```
///
/// The length of a path is defined as the count of Unicode Scalar Values and delimiters in it.
/// This is equivalent to the number of codepoints after string conversion.
///
/// # Performance
/// Creating a path from a string involves a parsing step; this operation is linear to the number of
/// bytes in the string.
///
/// Joining two paths is fairly cheap, and converting one to a string is almost a no-op. Similarly,
/// it is virtually free to acquire a string slice.
///
/// Due to the storage format, iteration over labels is fairly expensive. If there is a need for
/// extensive manipulation, in the middle or beginning of a path, it is likely better to work with a
/// sequence of labels. When all is said and done, they can be joined cheaply. This can be done
/// through the [LTreePathBuf] type.
///
/// # Paths
/// A path consists of ltree labels, separated by dots (`.` U+002E).
///
/// A label is a non-zero sequence of Unicode Scalar Values. Only alphanumeric codepoints, minus
/// characters (`-` U+002D) and underscores (`_` U+005F) are allowed in a label.
///
/// In PostgreSQL 16, the exact meaning of alphanumeric is determined by the database locale, but
/// this library accepts any value Unicode (and Rust) deems alphanumeric.
///
/// # Grammar
/// The exact grammar for a string path is given in EBNF:
/// ```text
/// Path            = [{Label, dot}, Label];
/// Label           = element, {element};
///
/// element         = alphanumeric | minus | underscore;
/// alphanumeric    = ? any Unicode alphanumeric scalar value ?;
/// minus           = ? U+002D ?;
/// underscore      = ? U+005F ?;
/// dot             = ? U+002E ?;
/// ```
///
/// *Note: paths can actually be empty.*
///
/// # References
/// - `contrib / ltree / ltree_io.c` in the PostgreSQL trunk.
/// - https://www.postgresql.org/docs/16/ltree.html
///
/// # Compatibility notes
/// - PostgreSQL 15 does not allow the minus-hyphen in a label, while PostgreSQL 16 does.
/// - PostgreSQL 15 has a maximum label size of 255 "characters".
///
#[derive(Debug, Clone, PartialEq, Hash)]
#[cfg_attr(test, derive(serde::Deserialize))] // We use serde only in tests.
pub struct LTreePath {
    // Note for the future: A Box<str> here would save a tiny bit memory but would slow down string
    // extraction.
    // A likely scenario is for someone to pass an LTreePath around, then extract the string to
    // store or pass to a database library. A Box<str> would require us to allocate memory, while
    // the current scheme lets us move the string out of this type, with almost no cost.
    value: String,
}

impl LTreePath {
    ///
    /// A path may not contain more labels than this.
    ///
    pub const MAX_PATH_LABEL_COUNT: usize = 65535;
    ///
    /// A label may not contain more Unicode Scalar Values than this.
    ///
    pub const MAX_LABEL_USV_COUNT: usize = 1000;
    ///
    /// Delimiter used between labels.
    ///
    pub const DELIMITER: char = '.';
    ///
    /// Delimiter used between labels, as a string slice for convenience.
    ///
    pub const DELIMITER_STR: &'static str = ".";

    /// Join two paths and return the result.
    ///
    /// Empty paths behave like an identity element; joining an empty path with a non-empty path
    /// returns the non-empty path.
    /// ```rust
    /// use ltree_rust::LTreePath;
    /// use std::str::FromStr;
    ///
    /// let a = LTreePath::from_str("").unwrap();
    /// let b = LTreePath::from_str("a.non-empty.path").unwrap();
    /// assert_eq!(b, a.join(&b).unwrap());
    /// ```
    ///
    /// This operation has `O(n)` time complexity, where n is the total length of the final path.
    ///
    /// **Panics** if the two strings have more than usize bytes in total.
    pub fn join(&self, other: &Self) -> LTreeParseResult<Self> {
        let capacity = self
            .len()
            .checked_add(other.len())
            .expect("joining two strings should not overflow an usize")
            .checked_add(Self::DELIMITER_STR.len())
            .expect("Joining two strings should not overflow an usize");

        let mut final_string = String::with_capacity(capacity);

        final_string.push_str(&self.value);
        if !self.is_empty() && !other.is_empty() {
            if self.labels().count() == Self::MAX_PATH_LABEL_COUNT
                || other.labels().count() == Self::MAX_PATH_LABEL_COUNT
            {
                return Err(LTreePathParseError::PathTooLong);
            }
            final_string.push(Self::DELIMITER);
        }
        final_string.push_str(&other.value);

        Ok(LTreePath {
            value: final_string,
        })
    }

    /// Get the path as a str. This operation is extremely cheap.
    ///
    /// This operation has `O(1)` time complexity.
    #[inline]
    pub fn as_str(&self) -> &str {
        &self.value
    }

    /// Whether this represents the empty path.
    ///
    /// This operation has `O(1)` time complexity.
    #[inline]
    pub fn is_empty(&self) -> bool {
        self.value.is_empty()
    }

    /// Create an empty path. This is convenient way to create an initial value for some operations.
    pub fn empty() -> Self {
        Self {
            value: String::new(),
        }
    }

    /// Iterate over the labels in the path.
    ///
    /// Iterating over all the labels has `O(n)` time complexity, where n is the length of the path.
    pub fn labels(&self) -> impl Iterator<Item = &str> {
        // "".split(".") will yield a single empty string as an exceptional case. We don't want
        // this. As LTreePath always validates the input on creation, we can assume this doesn't
        // hide problems caused by an invalid path (like "a..b" would).
        self.value.split(Self::DELIMITER).filter(|s| !s.is_empty())
    }

    // Note (siot): Don't expose this to outside code right now. Keep the API small.
    fn len(&self) -> usize {
        self.value.len()
    }
}

impl FromStr for LTreePath {
    type Err = LTreePathParseError;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        parsing::validate(s)?;

        Ok(LTreePath {
            value: s.to_owned(),
        })
    }
}

impl TryFrom<String> for LTreePath {
    type Error = LTreePathParseError;

    fn try_from(value: String) -> Result<Self, Self::Error> {
        parsing::validate(&value)?;

        Ok(LTreePath { value })
    }
}

impl Into<String> for LTreePath {
    fn into(self) -> String {
        self.value
    }
}

impl std::fmt::Display for LTreePath {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(&self.value)
    }
}

///
/// A builder for [LTreePath].
///
/// # Performance
/// Prefer [LTreePath::join] when two paths of non-trivial length are to be joined.
///
/// When many labels are to be added in sequence, prefer [LTreePathBuf], as it can be many times
/// quicker.
///
pub struct LTreePathBuf {
    labels: Vec<String>,
}

impl LTreePathBuf {
    pub fn new() -> Self {
        Self::with_capacity(0)
    }

    /// Gets a new builder with a given initial `capacity`. The capacity refers to the number of
    /// labels used in the intermediate stage.
    ///
    /// Letting `capacity = 0` signals a default capacity.
    pub fn with_capacity(capacity: usize) -> Self {
        Self {
            labels: Vec::with_capacity(capacity),
        }
    }

    pub fn push_label(&mut self, s: String) -> LTreeParseResult<()> {
        if s.is_empty() {
            Ok(())
        } else if self.labels.len() == LTreePath::MAX_PATH_LABEL_COUNT {
            Err(LTreePathParseError::PathTooLong)
        } else {
            validate_label(&s, 0)?;

            self.labels.push(s);

            Ok(())
        }
    }

    /// Pop the last label. For a path with more than one label, this gives the parent.
    pub fn pop(&mut self) -> Option<String> {
        self.labels.pop()
    }

    pub fn to_ltree_path(&self) -> LTreePath {
        let path_string = self.labels.join(LTreePath::DELIMITER_STR);

        LTreePath { value: path_string }
    }
}

impl From<LTreePath> for LTreePathBuf {
    /// Converts a path into a buf. This is much faster than splitting the string since the buf can
    /// be sure that the labels are valid.
    fn from(value: LTreePath) -> Self {
        LTreePathBuf {
            labels: value.labels().map(ToOwned::to_owned).collect(),
        }
    }
}
