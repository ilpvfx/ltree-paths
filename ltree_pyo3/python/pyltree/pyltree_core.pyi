from typing import Iterable

class LTreePath:
    def __init__(self, path: str | Iterable[str]) -> None:
        """Create a new path from a string or a sequence of strings."""

    def join(self, b: LTreePath) -> "LTreePath":
        """Join this and another path.

        This intelligently handles empty paths, and should be preferred to custom string operations.

        When joining many labels, prefer the LTreePathBuilder type.
        """

    def __bool__(self) -> str:
        """Tells whether the path is empty or not."""

    def __iter__(self) -> LTreeLabelIterator:
        """Acquire an iterator over the labels."""

class LTreePathBuilder:
    def __init__(self, value: None | LTreePath | Iterable[str]) -> None:
        """Creates a new builder, optionally pre-populating it with labels from a source."""

    def append(self, s: str):
        """Appends a label to the builder. Throws a parse error if the label is bad."""

    def pop(self) -> str | None:
        """Pop the last label or return None. Useful for finding the parent of a path."""

    def to_path(self) -> LTreePath:
        """Creates a new `LTreePath` object from the internal state. The internal state is not reset."""

class LTreeLabelIterator(Iterable[str]):
    """Iterator over the labels of a string."""
