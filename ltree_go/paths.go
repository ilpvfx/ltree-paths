//
// LTree Path support for Go.
//
// This package largely follows the API of the Rust module, with Go-specific flavour.
//

package ltree_go

import (
	"fmt"
	"strings"
)

// A path may not contain more labels than this.
const MAX_PATH_LABEL_COUNT uint = 65535

// A label may not contain more Unicode Scalar Values than this.
const MAX_LABEL_USV_COUNT uint = 1000

// Separator used between labels.
const CharDelimiter rune = '.'

// Represents a ltree path, as used in the PostgreSQL `ltree` extension.
//
// *Note: This targets PostgreSQL 16+.*
//
// The length of a path is defined as the count of Unicode Scalar Values and delimiters in it.
// This is equivalent to the number of codepoints after string conversion.
//
// # Performance
// Creating a path from a string involves a parsing step; this operation is linear to the number of
// bytes in the string.
//
// Joining two paths is fairly cheap, and converting one to a string is almost a no-op. Similarly,
// it is virtually free to acquire a string slice.
//
// Due to the storage format, iteration over labels is fairly expensive. If there is a need for
// extensive manipulation, in the middle or beginning of a path, it is likely better to work with a
// sequence of labels. When all is said and done, they can be joined cheaply. This can be done
// through the LTreePathBuilder type.
//
// # Paths
// A path consists of ltree labels, separated by dots (`.` U+002E).
//
// A label is a non-zero sequence of Unicode Scalar Values. Only alphanumeric codepoints, minus
// characters (`-` U+002D) and underscores (`_` U+005F) are allowed in a label.
//
// In PostgreSQL 16, the exact meaning of alphanumeric is determined by the database locale, but
// this library accepts any value Unicode (and Rust) deems alphanumeric.
//
// # Grammar
// The exact grammar for a string path is given in EBNF:
// ```text
// Path            = [{Label, dot}, Label];
// Label           = element, {element};
//
// element         = alphanumeric | minus | underscore;
// alphanumeric    = ? any Unicode alphanumeric scalar value ?;
// minus           = ? U+002D ?;
// underscore      = ? U+005F ?;
// dot             = ? U+002E ?;
// ```
//
// *Note: paths can actually be empty.*
//
// # References
// - `contrib / ltree / ltree_io.c` in the PostgreSQL trunk.
// - https://www.postgresql.org/docs/16/ltree.html
//
// # Compatibility notes
// - PostgreSQL 15 does not allow the minus-hyphen in a label, while PostgreSQL 16 does.
// - PostgreSQL 15 has a maximum label size of 255 "characters".
type LTreePath struct {
	// Note: like the Rust version, it might be worth adding a precomputed label count here.
	// It would remove a pretty large constant in Join's time complexity.
	value string
}

// Parse the provided string and if valid, return a LTreePath object representing it.
//
// The input must be valid UTF-8, this is checked when parsing the path.
//
// Creating a path has `O(n)` time complexity, where `n` is the number of unicode scalar values in
// the input.
//
// Returns `(path, nil)` or `(empty path, error)`.
func NewLTreePath(s string) (*LTreePath, error) {
	if err := validate(s); err != nil {
		return &LTreePath{}, err
	}
	return &LTreePath{s}, nil
}

// Join two paths and return the result.
//
// Empty paths behave like an identity element; joining an empty path with a non-empty path
// returns the non-empty path.
//
// This operation has `O(n)` time complexity, where `n` is the total length of the final path.
func (p *LTreePath) Join(other *LTreePath) (*LTreePath, error) {
	path1LabelCount := p.labelCount()
	path2LabelCount := other.labelCount()

	if path1LabelCount+path2LabelCount > MAX_PATH_LABEL_COUNT {
		return &LTreePath{}, makeError(PathTooLong, 0)
	}

	result := strings.Builder{}
	result.WriteString(p.String())

	if !p.IsEmpty() && !other.IsEmpty() {
		result.WriteRune(CharDelimiter)
	}

	result.WriteString(other.String())

	return &LTreePath{result.String()}, nil
}

// Whether this represents the empty path.
//
// This operation has `O(1)` time complexity.
func (p *LTreePath) IsEmpty() bool {
	return len(p.value) == 0
}

// Convert to string.
//
// This operation has `O(1)` time complexity.
func (p *LTreePath) String() string {
	return p.value
}

// Iterate over the byte slices, each slice representing a label in the path.
//
// Iteration continues until the callback gives an error or the last label is reached.
func (p *LTreePath) LabelsFunc(callback func(s string) error) error {
	if p.IsEmpty() {
		return nil
	}

	return sliceLabelsFunc(p.value, callback)
}

func (p *LTreePath) labelCount() uint {
	var count uint = 0
	p.LabelsFunc(func(s string) error {
		count += 1
		return nil
	})
	return count
}

func (p *LTreePath) Format(f fmt.State, verb rune) {
	f.Write([]byte(p.value))
}

// Represents a path error. Typically these are parsing errors.
type LTreePathError struct {
	message string
	code    LTreePathErrorCode
}

func (e *LTreePathError) Error() string {
	return e.message
}

func (e *LTreePathError) Code() LTreePathErrorCode {
	return e.code
}
