"""
This parses a DerivedGeneralCategory.txt and outputs a C header, with the ranges represented.

See 
    - http://www.unicode.org/reports/tr44/#Derived_Extracted
    - https://www.unicode.org/Public/UCD/latest/ucd/extracted/DerivedGeneralCategory.txt
"""

import argparse
from dataclasses import dataclass
from enum import StrEnum
from typing import Callable, TextIO


@dataclass
class UnicodeRange:
    codepoints: range = range(0, 0)
    property_value: str = ""

    @property
    def is_empty(self):
        return bool(self.codepoints)

    def __str__(self) -> str:
        a, b = self.codepoints.start, self.codepoints.stop - 1
        return f"{a:X}..{b:X} ; {self.property_value}"


def parse_line(line: str) -> UnicodeRange | None:
    """Parses a line in a Derived*.txt file.

    See https://www.unicode.org/reports/tr44/#Data_Fields for a description of the format. In short:

        - Comments start with `#`.
        - A line is separated into columns. Columns are separated by `;`.
        - The first column is a range or a single codepoint (implicitly a range of 1).
        - The second column is an associated property value. The property this relates to depends
          on the file.

    This will only ever check the first two columns. It ignores any other column.

    Errors:

        TypeError - line is not a string type
        ValueError - a parse error occurred

    """
    line = line.strip()
    comment_start = line.find("#")

    # This indicates that the whole line is a comment.
    if comment_start == 0:
        return None

    contains_comment = comment_start != -1

    noncomment_end = comment_start if contains_comment else len(line)

    noncomment = line[0:noncomment_end].strip()
    if not noncomment:
        return None

    columns = noncomment.split(";")

    if len(columns) < 2:
        raise ValueError(f"Expected at least two comma separated columns in '{line}'")

    range_str, property_value, *_ = columns

    return UnicodeRange(parse_range(range_str), property_value.strip())


def parse_range(s: str) -> range:
    """A range is either a single hexadecimal number, or two hexadecimal numbers, separated by `..`.

    Whitespace is significant.

    Grammar:

        ldots := ".."
        Range := hexadecimal, (ldots, hexadecimal)

    It is an error if end < start.

    Errors:

        ValueError - a parsing error occurred

    """
    s = s.strip()
    endpoints = s.split("..")
    if not endpoints:
        raise ValueError(f"The range string was empty")
    if len(endpoints) > 2:
        raise ValueError(f"The range string '{s}' contains more than one endpoint")

    def from_hex(x):
        y = int(x, base=16)
        if y < 0:
            raise ValueError(
                f"Negative value {x}. Negative range endpoints are not allowed."
            )
        return y

    start = from_hex(endpoints[0])
    end = from_hex(endpoints[1]) if len(endpoints) > 1 else start

    if end < start:
        raise ValueError(f"{end} < {start} in {s}")

    # Unicode ranges are inclusive, but Python ranges have an exclusive end.
    end += 1

    return range(start, end)


def parse_file(
    f: TextIO, property_value_filter: Callable[[str], bool]
) -> list[UnicodeRange]:
    results = []

    line_n = 0
    for line in f.readlines():
        rng = parse_line(line)

        if rng and property_value_filter(rng.property_value):
            results.append(rng)

        line_n += 1

    return results


class OutputFormat(StrEnum):
    HUMAN = "HUMAN"
    CPP_HEADER = "CPP"


class Order(StrEnum):
    NONE = "NONE"
    RANGE_START = "RANGE_START"


class CommaSeparated:
    def __init__(self, values: str) -> None:
        self.values = [s for s in (s.strip() for s in values.split(",")) if s]

    def __iter__(self):
        return iter(self.values)

    def __repr__(self) -> str:
        return ",".join(self.values)

    def __contains__(self, item) -> bool:
        return item in self.values


@dataclass
class OutputSettings:
    format: OutputFormat
    allowed_property_values: CommaSeparated
    filename: str
    varname: str
    order: Order


def print_human(ranges: list[UnicodeRange], settings: OutputSettings):
    print(f"# Property value whitelist: {settings.allowed_property_values}")
    print(f"# {len(ranges)} range(s) printed")
    print(f"// Range ends are not inclusive")
    
    for rng in ranges:
        print(rng)


def print_c_header(ranges: list[UnicodeRange], settings: OutputSettings):
    print(f"constexpr std::size_t {settings.varname}_COUNT = {len(ranges)};")

    print(f"// Generated from: {settings.filename}")
    print(f"// Property value whitelist: {settings.allowed_property_values}")
    print(f"// Range count: {len(ranges)}")
    print(f"// Range ends are not inclusive")
    print(f"constexpr unicode_range {settings.varname}[{len(ranges)}] = {{")
    for rng in ranges:
        # TODO: Technically we should escape strings, but this is very unlikely to be needed.
        print(
            f'    {{ /* {rng.property_value} */ 0x{rng.codepoints.start:X}, 0x{rng.codepoints.stop:X} }}, '
        )
    print(f"}};")


def print_ranges(ranges: list[UnicodeRange], settings: OutputSettings):
    fun = None
    if settings.format == OutputFormat.HUMAN:
        fun = print_human
    elif settings.format == OutputFormat.CPP_HEADER:
        fun = print_c_header
    else:
        raise ValueError("format should be a valid OutputFormat member")

    if settings.order == Order.RANGE_START:
        ranges.sort(key=lambda rng: rng.codepoints.start)
    elif settings.order != Order.NONE:
        raise ValueError("order should be a valid Order member")

    fun(ranges, settings)


def main():
    argparser = argparse.ArgumentParser()
    argparser.add_argument("FILE", type=argparse.FileType())
    argparser.add_argument(
        "--filter",
        help="Comma separated list of accepted property values",
        type=CommaSeparated,
        required=True,
    )
    argparser.add_argument(
        "--format",
        help="Output format",
        type=OutputFormat,
        default=OutputFormat.CPP_HEADER,
    )
    argparser.add_argument(
        "--varname",
        help="Variable name used in some outputs",
        type=str,
        default="UNICODE_RANGES",
    )
    argparser.add_argument(
        "--order",
        help="The ordering used when outputting the ranges",
        type=Order,
        default=Order.NONE,
    )

    args = argparser.parse_args()

    allowed_property_values = args.filter

    def filter_func(s: str) -> bool:
        return s in allowed_property_values

    with args.FILE as f:
        ranges = parse_file(f, property_value_filter=filter_func)

        settings = OutputSettings(
            args.format,
            allowed_property_values,
            filename=args.FILE.name,
            varname=args.varname,
            order=args.order,
        )
        print_ranges(ranges, settings)


if __name__ == "__main__":
    main()
