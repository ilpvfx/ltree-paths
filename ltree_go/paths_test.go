package ltree_go

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
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

		if path, err := NewLTreePath(v.Input); err == nil {
			// We have to filter out empty strings, as split on an empty string returns an empty string.
			actual := slices.DeleteFunc(
				strings.Split(path.String(), "."),
				func(s string) bool { return len(s) == 0 })

			if !slices.Equal(v.Expected, actual) {
				t.Errorf(
					"expected %v but got %v (beware, []string{} and []string{\"\"} prints the same)",
					v.Expected,
					actual)
			}
		} else {
			t.Errorf("Parsing failed: %v", err)
		}
	}

	for _, v := range cases.Fail {
		t.Logf("Parse Fail Case: %s input: <%s>", v.Name, v.Input)

		p, err := NewLTreePath(v.Input)
		if err == nil {
			t.Errorf("Expected an error! Got <%v> (err = <%v>)", p, err)
		} else if err.Error() != v.Message {
			t.Errorf("Should fail with [%v] not [%v]", v.Message, err)
		}
	}
}

func RunStringConvCases(t *testing.T, cases *StringConvCases) {
	for _, v := range cases.Pass {
		t.Logf("String Conv Pass Case %s input: <%s>", v.Name, v.Input)

		path1, _ := NewLTreePath(strings.Join(v.Input, "."))

		actual := path1.String()
		if v.Expected != actual {
			t.Errorf("expected <%v> not <%v>", v.Expected, actual)
		}
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

		if err != nil {
			t.Error(err)
		}

		if v.Expected != actual.String() {
			t.Errorf("expected <%v> not <%v>", v.Expected, actual)
		}
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
	if !path.IsEmpty() {
		t.Errorf("Expected empty path to be reported empty")
	}

	path, _ = NewLTreePath("a.b.c")
	if path.IsEmpty() {
		t.Errorf("Expected non-empty path to be reported non-empty")
	}
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
	if !slices.Equal(expected, labels) {
		t.Errorf("Expected <%v> not <%v>", expected, labels)
	}
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

	if err != nil {
		t.Error(nil)
	}

	expected := []string{"abcdef", "123", "ᛒᛒᛒᛒᛒ"}
	if !slices.Equal(expected, labels) {
		t.Errorf("Expected <%v> not <%v>", expected, labels)
	}
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
	if err != stopErr {
		t.Error("Error returned inside callback was not returned by LabelsByteFunc")
	}

	expected := []string{"a", "b"}
	if !slices.Equal(expected, labels) {
		t.Errorf("Expected <%v> not <%v>", expected, labels)
	}
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

	if _, err := NewLTreePath(pathPrototype.String()); err == nil {
		t.Error("Expected error but got none")
	}
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
	if err != nil {
		t.Error(err)
	}
	thinMintWafer, err := NewLTreePath("a")
	if err != nil {
		t.Error(err)
	}

	if _, err := almostTooLong.Join(thinMintWafer); err == nil {
		t.Error("Expected error but got none")
	}
}

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
