//
// LTree Path support for Go.
//
// This package largely follows the API of the Rust module, with Go-specific flavour.
//

package ltree_go

import "strings"

type LTreePath struct {
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
	return sliceLabelsFunc(p.value, callback)
}

func (p *LTreePath) labelCount() uint {
	var count uint = 0
	sliceLabelsFunc(p.value, func(_ string) error {
		count += 1
		return nil
	})
	return count
}

// Represents a path error. Typically these are parsing errors.
type LTreePathError struct {
	message string
	code    ParseErrorCode
}

func (e *LTreePathError) Error() string {
	return e.message
}

func (e *LTreePathError) Code() ParseErrorCode {
	return e.code
}

// A path may not contain more labels than this.
const MAX_PATH_LABEL_COUNT uint = 65535

// A label may not contain more Unicode Scalar Values than this.
const MAX_LABEL_USV_COUNT uint = 1000

// CharDelimiter used between labels.
const CharDelimiter rune = '.'
