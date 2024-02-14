package ltree_go

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type LTreePathSimple = []string

type PassCase[TInput interface{}, TExpected interface{}] struct {
	Name     string
	Input    TInput
	Expected TExpected
}

type FailCase[TInput interface{}] struct {
	Name    string
	Input   TInput
	Message string
}

type ParseTextCases struct {
	Pass []PassCase[string, LTreePathSimple]
	Fail []FailCase[string]
}

type StringConvCases struct {
	Pass []PassCase[LTreePathSimple, string]
	Fail []FailCase[string] // Note (siot): There are none at the time of writing.
}

type JoinCases struct {
	Pass []PassCase[[]LTreePathSimple, string]
	Fail []FailCase[string] // Note (siot): There are none at the time of writing.
}

type CommonTestCases struct {
	Parse_Text_Cases  *ParseTextCases
	String_Conv_Cases *StringConvCases
	Join_Cases        *JoinCases
}

const COMMON_TEST_SUITE_PATH_RELATIVE = "../common/test_cases.json"

func RunParseTextCases(t *testing.T, cases *ParseTextCases) {

	for _, v := range cases.Pass {
		t.Logf("Parse Pass Case %s input: <%s>", v.Name, v.Input)

		path, err := NewLTreePath(v.Input)
		require.Nil(t, err)

		// We have to filter out empty strings, as split on an empty string returns an empty string.
		actual := slices.DeleteFunc(
			strings.Split(path.String(), "."),
			func(s string) bool { return len(s) == 0 })

		assert.Equal(t, v.Expected, actual)
	}

	for _, v := range cases.Fail {
		t.Logf("Parse Fail Case: %s input: <%s>", v.Name, v.Input)

		_, err := NewLTreePath(v.Input)
		require.NotNil(t, err)
		assert.Equal(t, v.Message, err.Error())
	}
}

func RunStringConvCases(t *testing.T, cases *StringConvCases) {
	for _, v := range cases.Pass {
		t.Logf("String Conv Pass Case %s input: <%s>", v.Name, v.Input)

		path1, _ := NewLTreePath(strings.Join(v.Input, "."))

		actual := path1.String()

		assert.Equal(t, v.Expected, actual)
	}

	for _, v := range cases.Fail {
		t.Logf("String Conv Fail Case %s input: <%s>", v.Name, v.Input)
		t.Fatal("Did not expect failing string conversion cases")
	}
}

func RunJoinCases(t *testing.T, cases *JoinCases) {
	for _, v := range cases.Pass {
		t.Logf("Join Pass Case %s input: <%s>", v.Name, v.Input)

		path1, _ := NewLTreePath(strings.Join(v.Input[0], "."))
		path2, _ := NewLTreePath(strings.Join(v.Input[1], "."))

		actual, err := path1.Join(path2)
		require.Nil(t, err)

		assert.Equal(t, v.Expected, actual.String())
	}

	for _, v := range cases.Fail {
		t.Logf("Join Fail Case %s input: <%s>", v.Name, v.Input)
		t.Fatal("Did not expect failing join cases")
	}
}

func TestCommon(t *testing.T) {
	file, err := os.Open(COMMON_TEST_SUITE_PATH_RELATIVE)
	if err != nil {
		t.Fatalf("Couldn't open the test cases: %s", err)
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Couldn't read the test cases: %s", err)
	}

	var common CommonTestCases
	if err := json.Unmarshal(bytes, &common); err != nil {
		t.Fatalf("JSON Decoding Problem: %s", err)
	}
	t.Logf("Successfully read the common test cases from: %s", COMMON_TEST_SUITE_PATH_RELATIVE)

	RunParseTextCases(t, common.Parse_Text_Cases)
	RunStringConvCases(t, common.String_Conv_Cases)
	RunJoinCases(t, common.Join_Cases)
}

func TestPathIsEmpty(t *testing.T) {
	path, _ := NewLTreePath("")
	require.True(t, path.IsEmpty())

	path, _ = NewLTreePath("a.b.c")
	require.False(t, path.IsEmpty(), "Non empty path should not be reported empty")
}

// Make sure this function handles sequences with bad UTF-8 in them. In essence, we want it to split
// on "." regardless how horrible the data is between. We check for UTF-8 problems later.
func TestSliceLabelsFuncBadUtf(t *testing.T) {
	// \x8C has the top continuation bits set, but the surrounding characters are 7-bit so none are
	// continuating bytes.
	shadyString := "abc\x8Cdef.123"

	labels := []string{}
	sliceLabelsFunc(shadyString, func(s string) error {
		labels = append(labels, s)
		return nil
	})

	expected := []string{"abc\x8Cdef", "123"}
	assert.Equal(t, expected, labels)
}

type DummyError struct{}

func (de *DummyError) Error() string {
	return "Dummy Error"
}

func TestLabelsFunc(t *testing.T) {
	labels := []string{}

	nonEmpty, _ := NewLTreePath("abcdef.123.ᛒᛒᛒᛒᛒ")

	err := nonEmpty.LabelsFunc(func(s string) error {
		labels = append(labels, s)
		return nil
	})

	require.Nil(t, err)

	expected := []string{"abcdef", "123", "ᛒᛒᛒᛒᛒ"}
	assert.Equal(t, expected, labels)
}

func TestLabelsFuncEdgeCases(t *testing.T) {
	empty, _ := NewLTreePath("")

	shouldNotBeCalled := false
	empty.LabelsFunc(func(s string) error {
		shouldNotBeCalled = true
		return nil
	})

	if shouldNotBeCalled {
		t.Error("Callback called on empty path")
	}

	labels := []string{}

	stopErr := &DummyError{}
	nonEmpty, _ := NewLTreePath("a.b.c")
	err := nonEmpty.LabelsFunc(func(s string) error {
		if s == "c" {
			return stopErr
		}
		labels = append(labels, s)
		return nil
	})
	assert.Equal(t, err, stopErr, "Error returned inside callback was not returned by LabelsByteFunc")

	expected := []string{"a", "b"}
	assert.Equal(t, expected, labels)
}

func TestNewLTreePathCountTooLarge(t *testing.T) {
	if MAX_PATH_LABEL_COUNT > 1<<30 {
		t.Fatal("This test might take a very long time and eat a lot of resources. Redesign it.")
	}

	label := "a"
	pathPrototype := strings.Builder{}

	pathPrototype.WriteString(label)
	for i := 0; i < int(MAX_PATH_LABEL_COUNT); i++ {
		pathPrototype.WriteRune(CharDelimiter)
		pathPrototype.WriteString(label)
	}

	_, err := NewLTreePath(pathPrototype.String())
	require.NotNil(t, err, "The path should be too long and generate an error")
}

func TestJoinCountTooLarge(t *testing.T) {
	if MAX_PATH_LABEL_COUNT > 1<<30 {
		t.Fatal("This test might take a very long time and eat a lot of resources. Redesign it.")
	}

	label := "a"
	pathPrototype := strings.Builder{}
	pathPrototype.WriteString(label)
	for i := 0; i < int(MAX_PATH_LABEL_COUNT-1); i++ {
		pathPrototype.WriteRune(CharDelimiter)
		pathPrototype.WriteString(label)
	}

	almostTooLong, err := NewLTreePath(pathPrototype.String())
	require.Nil(t, err)

	thinMintWafer, err := NewLTreePath("a")
	require.Nil(t, err)

	_, err = almostTooLong.Join(thinMintWafer)
	require.NotNil(t, err, "This join should give us a path that is too long")

}

// Note: Do not use testify inside the benchmarks, it seems to have monstrous overhead.
func BenchmarkCreate(b *testing.B) {
	cases := []string{
		"widget-company",
		"widget-company.products",
		"widget-company.products.year",
		"widget-company.products.year.2024",
		"widget-company.products.year.2024.ᛒluetoothSpeaker",
	}

	for k, s := range cases {
		b.Run(fmt.Sprintf("label%v", k), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := NewLTreePath(s)
				if err != nil {
					b.Fatal(s, err)
				}
			}
		})
	}
}

func BenchmarkJoin(b *testing.B) {
	cases := []string{
		"",
		"widget-company",
		"widget-company.products",
		"widget-company.products.year",
		"widget-company.products.year.2024",
		"widget-company.products.year.2024.ᛒluetoothSpeaker",
	}

	path2, err := NewLTreePath("specs_v2.output.wattage.exagerated")
	if err != nil {
		b.Fatal(err)
	}

	for k, s := range cases {
		b.Run(fmt.Sprintf("label%v", k), func(b *testing.B) {
			path1, err := NewLTreePath(s)
			if err != nil {
				b.Fatal(s, err)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				path1.Join(path2)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	path, _ := NewLTreePath("abc.def.123")

	expected := path.String()
	actual := fmt.Sprintf("%v", path)

	assert.Equal(t, expected, actual)
}
