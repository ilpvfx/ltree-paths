package ltree_go

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	builder := NewLTreePathBuilder()

	path := builder.Path()
	assert.True(t, path.IsEmpty(), "A new builder should return empty paths")
}

func TestBuilderAppend(t *testing.T) {
	builder := NewLTreePathBuilder()

	err := builder.Append("abc")
	if err != nil {
		t.Fatal(err)
	}

	err = builder.Append("very bad label $£$@£$")
	require.NotNil(t, err)

	builder.Append("123")

	path := builder.Path()
	// TODO (siot): Path comparisons?
	expected := "abc.123"
	assert.Equal(t, expected, path.String())
}

func TestBuilderAppendTooLong(t *testing.T) {
	if MAX_PATH_LABEL_COUNT > 1<<30 {
		t.Fatal("This test might take a very long time and eat a lot of resources. Redesign it.")
	}

	label := "a"
	pathBuilder := NewLTreePathBuilder()

	err := pathBuilder.Append(label)
	for i := 0; i < int(MAX_PATH_LABEL_COUNT-1); i++ {
		err = pathBuilder.Append(label)
	}

	require.Nil(t, err)

	err = pathBuilder.Append(label)
	require.NotNil(t, err, "This path should be too long and fail")
}

func TestBuilderAppendEmptyLabels(t *testing.T) {
	builder := NewLTreePathBuilder()

	err := builder.Append("")
	require.NotNil(t, err)
	assert.True(t, builder.Path().IsEmpty(), "Expected an empty path")

	builder.Append("abc")
}

func TestBuilderPop(t *testing.T) {
	builder := NewLTreePathBuilder()
	assert.Equal(t, builder.Pop(), "", "Popping on an empty builder should give the empty string")

	builder.Append("a")
	builder.Append("b")
	builder.Append("c")

	assert.Equal(t, builder.Pop(), "c")
	assert.Equal(t, builder.Pop(), "b")
	assert.Equal(t, builder.Pop(), "a")
	assert.Equal(t, builder.Pop(), "", "Popping on an empty builder should give the empty string")
}

func TestBuilderAppendPath(t *testing.T) {
	builder := NewLTreePathBuilder()

	empty, _ := NewLTreePath("")
	err := builder.AppendPath(empty)
	require.Nil(t, err)

	nonEmpty, _ := NewLTreePath("abc.123")
	err = builder.AppendPath(nonEmpty)
	require.Nil(t, err)

	assert.Equal(t, nonEmpty.String(), builder.Path().String())
}
