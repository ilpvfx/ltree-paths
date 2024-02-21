import sys

# Requires Python 3.11
import tomllib

import argparse
from argparse import ArgumentParser, Namespace

EXIT_FAILURE_CODE = -1


def exit_failure(msg):
    print(msg, file=sys.stderr)
    exit(EXIT_FAILURE_CODE)


def check_property(data, key: str, expected_value: str):
    """
    >>> check_property({"key": 123}, "key", "123")
    True
    >>> check_property({"key": 123}, "key", "abc")
    Traceback (most recent call last):
        ...
    KeyError: "value for key key did not match: 'abc' != '123'"
    >>> check_property({"key": {"subkey": "hello"}}, "key.subkey", "hello")
    True
    >>> check_property({"key": {"subkey": "hello"}}, "key.notfound", "hello")
    Traceback (most recent call last):
        ...
    KeyError: 'key.notfound did not match a property in the data'
    >>> check_property({"key": {"subkey": "hello"}}, "", "hello")
    Traceback (most recent call last):
        ...
    ValueError: empty key
    >>> check_property({"key": {"subkey": "hello"}}, "key.notfound.", "hello")
    Traceback (most recent call last):
        ...
    ValueError: empty a label after a delimiter
    """
    key = key.strip()
    if not key:
        raise ValueError("empty key")

    rem_subkeys = key.split(".")
    if "" in rem_subkeys:
        raise ValueError("empty a label after a delimiter")

    subtree = data
    while rem_subkeys:
        next_subkey = rem_subkeys[0]

        if next_subkey in subtree:
            value = subtree[next_subkey]

            if len(rem_subkeys) == 1:
                str_value = str(value)
                if expected_value != str_value:
                    raise KeyError(
                        f"value for key {key} did not match: {repr(expected_value)} != {repr(str_value)}"
                    )
                else:
                    return True
            else:
                subtree = value
                if not isinstance(subtree, dict):
                    raise ValueError("one of the objects in the tree was not a dict")
        else:
            raise KeyError(f"{key} did not match a property in the data")
        rem_subkeys = rem_subkeys[1:]


def run_tests():
    import doctest

    doctest.testmod()


def run_compare(filename, key, expected):
    toml = None
    try:
        with open(filename, "rb") as f:
            toml = tomllib.load(f)
    except BaseException as e:
        exit_failure(f"Failure to read TOML: {e}")

    try:
        check_property(toml, key, expected)
    except BaseException as e:
        exit_failure(f"Comparison failed: {e}")


def main():
    parser = ArgumentParser()

    subparsers = parser.add_subparsers()

    compare = subparsers.add_parser("compare")
    compare.add_argument("TOMLFILE", type=str)
    compare.add_argument("KEY", type=str)
    compare.add_argument("EXPECTED", type=str)
    compare.set_defaults(cmd="compare")

    test = subparsers.add_parser("test")
    test.set_defaults(cmd="test")

    options = parser.parse_args()

    if options.cmd == "test":
        run_tests()
    elif options.cmd == "compare":
        run_compare(options.TOMLFILE, options.KEY, options.EXPECTED)
    else:
        exit_failure("Internal error")


if __name__ == "__main__":
    main()
