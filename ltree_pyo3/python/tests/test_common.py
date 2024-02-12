from dataclasses import dataclass
import re

from pytest import raises, fixture

from pyltree import LTreePath, ParseError, LTreePathBuilder

COMMON_TEST_CASES_PATH = "../common/test_cases.json"


@fixture
def common_test_data():
    import json

    data = None
    with open(COMMON_TEST_CASES_PATH) as f:
        data = json.load(f)

    return data


@dataclass
class CommonTester:
    name: str
    operation: callable
    asserter: callable


def run_common_test(test_data, tester: CommonTester):
    cases = test_data[tester.name]

    pass_cases = cases["pass"]
    for case in pass_cases:
        input = case["input"]
        expected = case["expected"]

        actual = tester.operation(input)
        tester.asserter(expected, actual)

    for case in cases["fail"]:
        input = case["input"]
        message = case["message"]
        with raises(ParseError, match=re.escape(message)):
            tester.operation(input)


def test_parse_text_cases(common_test_data):
    def operation(input):
        return LTreePath(input)

    def asserter(expected, actual):
        assert [label for label in str(actual).split(".") if label != ""] == expected

    tester = CommonTester("parse_text_cases", operation, asserter)
    run_common_test(common_test_data, tester)


def test_string_conv_cases(common_test_data):
    def operation(input):
        a = LTreePath(".".join(input))
        return str(a)

    def asserter(expected, actual):
        assert actual == expected

    tester = CommonTester("string_conv_cases", operation, asserter)
    run_common_test(common_test_data, tester)


def test_join_cases(common_test_data):
    def operation(input):
        a = LTreePath(".".join(input[0]))
        b = LTreePath(".".join(input[1]))
        return a.join(b)

    def asserter(expected, actual):
        assert str(actual) == expected

    tester = CommonTester("join_cases", operation, asserter)
    run_common_test(common_test_data, tester)


def test_repr():
    path = LTreePath("x.y.z")
    assert eval(repr(path)) == path


def test_equality():
    a = LTreePath("x.y")
    b = LTreePath("z")
    c = LTreePath("x.y")

    assert a == a
    assert a != b
    assert b != a
    assert a == c

    # We shouldn't support other types atm.
    assert not (a == 123)
    assert not (a == ["x", "y"])


def test_path_from_iterable():
    a = LTreePath("x.y.z".split("."))

    with raises(TypeError):
        LTreePath(["a", "b", 123])


def test_iter_over_labels():
    a = LTreePath("x.y.z")

    assert ["x", "y", "z"] == list(a)

    b = LTreePath("")
    assert [] == list(b)


def test_empty_is_bool_false():
    a = LTreePath("x.y.z")
    assert bool(a)

    b = LTreePath("")
    assert not bool(b)


def test_builder_empty():
    builder = LTreePathBuilder()
    path = builder.to_path()

    assert not bool(path)
    assert str(path) == ""

    builder.append("a")
    assert LTreePath("a") == builder.to_path()

    with raises(ParseError):
        builder.append("a bad label?!?!")
    assert LTreePath("a") == builder.to_path()

    # An empty list should also work
    builder = LTreePathBuilder([])
    path = builder.to_path()

    assert not bool(path)
    assert str(path) == ""


def test_builder_from_labels():
    builder = LTreePathBuilder(["a", "b", "c"])
    path = builder.to_path()

    assert path == LTreePath("a.b.c")

    with raises(ParseError):
        builder = LTreePathBuilder(["a", "bad label is this yes?", "c"])


def test_builder_to_ltree_path_does_not_flush():
    builder = LTreePathBuilder(["a", "b", "c"])
    for _ in range(0, 2):
        assert LTreePath("a.b.c") == builder.to_path()


def test_builder_construct_with_bad_type():
    for val in [123, ["a", "b", 123], False]:
        print(val)
        with raises(TypeError):
            LTreePathBuilder(val)


def test_builder_from_ltree_path():
    path = LTreePath("a.b.c")
    builder = LTreePathBuilder(path)
    assert path == builder.to_path()


def test_builder_pop():
    builder = LTreePathBuilder("a.b.c".split("."))
    assert "c" == builder.pop()
    assert "b" == builder.pop()
    assert "a" == builder.pop()
    assert None == builder.pop()

def test_unicode_surrogates():
    surrogate1 = "\uD800"
    
    with raises(UnicodeEncodeError):
        LTreePath(["nextoneisbad", surrogate1])

    with raises(UnicodeEncodeError):
        LTreePathBuilder(["nextoneisbad", surrogate1])