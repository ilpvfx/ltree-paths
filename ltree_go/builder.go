package ltree_go

import "strings"

// A builder type for LTree Paths. This can be a much more efficient way of building paths when
// there are many labels to be joined together.
//
// When two relatively large paths are to be joined, this is less effective.
type LTreePathBuilder struct {
	labels []string
}

// Creates a new builder. The builder is immediately ready for use and represents the empty path
// upon creation.
func NewLTreePathBuilder() *LTreePathBuilder {
	return &LTreePathBuilder{}
}

// Tries to append the given label. Fails if the string `s` is not a valid LTree Path label or if
// the resulting path would be too long.
//
// This has `O(k)` time complexity where `k` is the number of elements of the `s`.
func (b *LTreePathBuilder) Append(s string) error {
	if _, err := validateLabel(s, 0); err != nil {
		return err
	}
	return b.nonValidatingAppend(s)
}

// Tries to append the given path. Fails if the resulting path would be too long.
//
// This has `O(k)` time complexity where `k` is the number of elements of the path.
func (b *LTreePathBuilder) AppendPath(p *LTreePath) error {
	if p.IsEmpty() {
		return nil
	}

	return sliceLabelsFunc(p.value, func(s string) error {
		return b.nonValidatingAppend(s)
	})
}

// Tries to pop the last label. If there are no labels (an empty path), this returns the empty
// string.
//
// This has `O(1)` time complexity.
func (b *LTreePathBuilder) Pop() string {
	count := len(b.labels)
	if count > 0 {
		last := b.labels[count-1]
		b.labels = b.labels[:count-1]
		return last
	} else {
		return ""
	}
}

// Return a path representing the current state of the builder.
//
// This has `O(n)` time complexity where `n` is the number of elements of the final path.
func (b *LTreePathBuilder) Path() *LTreePath {
	// We know the labels are valid by this point, so we can bypass the parsing.
	return &LTreePath{strings.Join(b.labels, ".")}
}

func (b *LTreePathBuilder) nonValidatingAppend(s string) error {
	if len(b.labels)+1 > int(MAX_PATH_LABEL_COUNT) {
		return makeError(PathTooLong, 0)
	}

	b.labels = append(b.labels, s)

	return nil
}
