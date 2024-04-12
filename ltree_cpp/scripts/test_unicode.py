from pytest import raises
import unicode_ranges
from unicode_ranges import UnicodeRange


def test_parse_line():
    # An amalgamation of some real input
    test_data = """
    # General_Category=Math_Symbol

    002B          ; Sm #       PLUS SIGN
    003C..003E    ; Sm #   [3] LESS-THAN SIGN..GREATER-THAN SIGN
    1D540..1D544  ; Lu #   [5] MATHEMATICAL DOUBLE-STRUCK CAPITAL I..MATHEMATICAL DOUBLE-STRUCK CAPITAL M

    # Ignore this.. plzzzzz
    # 002B          ; Sm #       PLUS SIGN
    
    1D546         ; Lu #       MATHEMATICAL DOUBLE-STRUCK CAPITAL O
    1D54A..1D550  ; Lu #   [7] MATHEMATICAL DOUBLE-STRUCK CAPITAL S..MATHEMATICAL DOUBLE-STRUCK CAPITAL Y
    """

    expected = [
        UnicodeRange(range(0x002B, 0x002B + 1), "Sm"),
        UnicodeRange(range(0x003C, 0x003E + 1), "Sm"),
        UnicodeRange(range(0x1D540, 0x1D544 + 1), "Lu"),
        UnicodeRange(range(0x1D546, 0x1D546 + 1), "Lu"),
        UnicodeRange(range(0x1D54A, 0x1D550 + 1), "Lu"),
    ]

    parsed = [unicode_ranges.parse_line(line) for line in test_data.splitlines()]
    actual = [r for r in parsed if r is not None]

    assert expected == actual


def test_parse_range():
    from unicode_ranges import parse_range

    with raises(ValueError):
        parse_range("")
    
    with raises(ValueError):
        parse_range(" nothexadecimal ")

    with raises(ValueError):
        parse_range(" 0.. ")

    with raises(ValueError):
        parse_range(" 2..1")

    assert range(0, 2) == parse_range(" 0 .. 1")
    assert range(1, 3) == parse_range(" 01..02")
    assert range(1, 2) == parse_range(" 01")

    with raises(ValueError):
        parse_range(" -01")