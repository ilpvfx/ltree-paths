use super::LTreePath;
use std::fmt::Display;

pub type SourceLocation = usize;

#[derive(PartialEq, Debug)]
pub enum LTreePathParseError {
    ExpectedLabel(SourceLocation),
    InvalidCharacter(SourceLocation),
    LabelTooLong,
    PathTooLong,
}

impl Display for LTreePathParseError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            LTreePathParseError::ExpectedLabel(loc) => {
                write!(f, "Parse error: Expected label at offset {}", loc)
            }
            LTreePathParseError::InvalidCharacter(loc) => {
                write!(f, "Parse error: Invalid character at offset {}", loc)
            }
            LTreePathParseError::LabelTooLong => write!(
                f,
                "Parse error: Label too long (max {} UTF-8 scalar values)",
                LTreePath::MAX_LABEL_USV_COUNT
            ),
            LTreePathParseError::PathTooLong => {
                write!(
                    f,
                    "Parse error: Path too long (max {} labels)",
                    LTreePath::MAX_PATH_LABEL_COUNT
                )
            }
        }
    }
}

impl std::error::Error for LTreePathParseError { }

pub type LTreeParseResult<T> = Result<T, LTreePathParseError>;

pub(crate) fn validate_label(s: &str, offset: usize) -> LTreeParseResult<usize> {
    use LTreePathParseError::*;

    const MINUS: char = '-';
    const UNDERSCORE: char = '_';

    let mut new_offset = offset;
    let mut label_usv_count = 0usize;

    for c in s.chars() {
        label_usv_count += 1;

        if label_usv_count > LTreePath::MAX_LABEL_USV_COUNT {
            return Err(LabelTooLong);
        } else if c.is_alphanumeric() || c == MINUS || c == UNDERSCORE {
            new_offset += 1;
        } else {
            return Err(InvalidCharacter(new_offset));
        }
    }

    Ok(new_offset)
}

pub fn validate(s: &str) -> LTreeParseResult<()> {
    use LTreePathParseError::*;

    // NOTE: interestingly, if we use a str pattern below, we get a performance degradation.
    // It seems the standard library has some overhead in this case.
    const DELIMITER: char = LTreePath::DELIMITER;

    if s.is_empty() {
        Ok(())
    } else {
        assert!(!s.is_empty());

        let mut usv_offset = 0usize;
        let mut label_count = 0usize;

        for label in s.split(DELIMITER) {
            // Account for the delimiters.
            if label_count > 0 {
                usv_offset += 1;
            }

            // This indicates that we have a situation like ".", "a..b", ".a", "a.".
            if label.is_empty() {
                return Err(ExpectedLabel(usv_offset));
            }

            label_count += 1;
            if label_count > LTreePath::MAX_PATH_LABEL_COUNT {
                return Err(PathTooLong);
            }

            usv_offset = validate_label(label, usv_offset)?;
        }

        Ok(())
    }
}
