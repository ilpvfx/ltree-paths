# LTree Paths
Work with LTree paths in multiple languages.

## A Python example
Joining two paths:
```python
>>> from pyltree import LTreePath
>>> path1 = LTreePath(["my", "company", "products"])
>>> path2 = LTreePath("icecream.flavour.mintchip")
>>> path1.join(path2)
LTreePath(r"my.company.products.icecream.flavour.mintchip")
```

Join will correctly handle empty paths:
```python
>>> path = LTreePath("this.is.a.path")
>>> path.join(LTreePath(""))
LTreePath(r"this.is.a.path")
```

We can easily iterate the labels:
```python
>>> list(LTreePath("this.is.a.path"))
['this', 'is', 'a', 'path']
```

Paths are always validated:
```python
>>> LTreePath("the.last.label.is.invalid $!#")
Traceback (most recent call last):
  File "<stdin>", line 1, in <module>
pyltree_core.PyParseError: Parse error: Invalid character at offset 25
```

## Go example
Joining two paths:
```go
path1, err := NewLTreePath("my.company.products")
path2, err := NewLTreePath("icecream.flavour.mintchip")

path3, err := path1.Join(path2)
println(path3)
// my.company.products.icecream.flavour.mintchip
```

Quickly convert to string:
```go
path1, err := NewLTreePath("chariots.of.fire")
println(path1.String())
// chariots.of.fire
```

Use a builder to efficiently add many labels:
```go
builder := NewLTreePathBuilder()
builder.Append("chariots")
builder.Append("of")
builder.Append("fire")

println(builder.Path())
// chariots.of.fire
```

## Subprojects
- [ltree_rust](./ltree_rust/) (ltree_rust): The Rust reference implementation.
- [ltree_pyo3](./ltree_pyo3) (pyltree): The Python package, built from a Pyo3 Rust source.
- [ltree_go](./ltree_go) (ltree_go): The Go package.

## Testing
The file `common/test_cases.json` contains a set of test cases which are designed to be consumed 
from multiple languages.

See each subdirectory for information about testing for that particular language.

## Implementation details

### Python
- Memory errors in Rust will yield a PyOxide panic exception.
- Converting a string into a path may fail if the string contains lone surrogate characters. Typically this would happen on corrupted input or a purposefully broken Unicode string.

### Go
- Converting a string into a path may fail if the string contains lone surrogate characters or bad UTF-8 bytes. Typically this would happen on corrupted input or a purposefully broken Unicode string.


## Postgres details

### Sources
- https://www.postgresql.org/docs/current/ltree.html
- Information about the binary format is found in PostgreSQL's `contrib/ltree/ltree_io.c`.