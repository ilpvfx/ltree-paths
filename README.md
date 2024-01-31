# LTree Paths
Work with LTree paths in multiple languages.

## Subprojects
- [ltree_rust](./ltree_rust/): The Rust reference implementation.
- [ltree_pyo3](./ltree_pyo3): The Python package, built from a Pyo3 Rust source.
- [ltree_go](./ltree_go): The Go package.

## Testing
The file `common/test_cases.json` contains a set of test cases which are designed to be consumed 
from multiple languages.

See each subdirectory for information about testing for that particular language.

## Postgres details

### Sources
- https://www.postgresql.org/docs/current/ltree.html
- Information about the binary format is found in PostgreSQL's `contrib/ltree/ltree_io.c`.