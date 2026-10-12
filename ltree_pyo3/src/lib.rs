use std::str::FromStr;

use pyo3::{
    basic::CompareOp,
    create_exception,
    exceptions::{PyNotImplementedError, PyStopIteration, PyTypeError},
    prelude::*,
    types::{PyIterator, PyString},
};

use ltree_rust::{
    paths::{LTreeParseResult, LTreePathBuf},
    *,
};

create_exception!(pyltree_core, PyParseError, pyo3::exceptions::PyValueError);

#[pyclass(name = "LTreePath", module = "pyltree", frozen)]
struct PyLTreePath(LTreePath);

// NOTE: documentation should be placed in the *.pyi file. Doc comments on pymethods here are
// overwritten using the typings file later, so it would be wasted work.

#[pymethods]
impl PyLTreePath {
    #[new]
    fn new(path: &Bound<'_, PyAny>, py: Python<'_>) -> PyResult<Self> {
        // FIXME: the error returned when a parse error occurs in the iteration of labels is pretty
        // confusing as it doesn't say which label was bad. This could be improved by at least
        // saying what character it was, not just its position.
        if let Ok(str_path) = path.cast::<PyString>() {
            // see buf_from_iterator why to_cow might fail.
            map_to_py_result(LTreePath::from_str(&str_path.to_cow()?))
        }
        // We must try the iterator last, as the preceeding Python types may be iterable.
        else if let Ok(iter) = path.try_iter() {
            map_to_py_result(Ok(buf_from_iterator(&iter, py)?.to_ltree_path()))
        } else {
            Err(PyTypeError::new_err(
                "Argument should be empty or an iterable of strings",
            ))
        }
    }

    fn join(&self, b: &Self) -> PyResult<Self> {
        let a_inner = &self.0;
        let b_inner = &b.0;

        map_to_py_result(a_inner.join(b_inner))
    }

    fn __repr__(&self) -> String {
        format!("LTreePath(r\"{}\")", self.0.as_str())
    }

    fn __str__(&self) -> &str {
        self.0.as_str()
    }

    fn __bool__(&self) -> bool {
        !self.0.is_empty()
    }

    fn __richcmp__(&self, other: &PyLTreePath, op: CompareOp) -> PyResult<bool> {
        match op {
            CompareOp::Eq => Ok(self.0.as_str() == other.0.as_str()),
            CompareOp::Ne => Ok(self.0.as_str() != other.0.as_str()),
            _ => Err(PyNotImplementedError::new_err("Not a supported operation")),
        }
    }

    fn __iter__(&self) -> PyLTreeLabelIterator {
        // FIXME: this is so-so. We are basically splitting and then duplicating every label.
        PyLTreeLabelIterator {
            iter: Box::new(
                self.0
                    .labels()
                    .map(ToOwned::to_owned)
                    .collect::<Vec<String>>()
                    .into_iter(),
            ),
        }
    }
}

#[pyclass(name = "LTreeLabelIterator", module = "pyltree")]
struct PyLTreeLabelIterator {
    iter: Box<dyn Iterator<Item = String> + Send + Sync>,
}

#[pymethods]
impl PyLTreeLabelIterator {
    fn __iter__(me: PyRef<Self>) -> PyRef<Self> {
        me
    }

    fn __next__(mut me: PyRefMut<Self>) -> Option<String> {
        me.iter.next()
    }
}

#[pyclass(name = "LTreePathBuilder", module = "pyltree")]
struct PyLTreePathBuilder {
    buf: LTreePathBuf,
}

impl PyLTreePathBuilder {
    fn new_empty() -> Self {
        PyLTreePathBuilder {
            buf: LTreePathBuf::new(),
        }
    }
}

#[pymethods]
impl PyLTreePathBuilder {
    #[new]
    #[pyo3(signature = (value=None))]
    fn new(value: Option<&Bound<'_, PyAny>>, py: Python<'_>) -> PyResult<PyLTreePathBuilder> {
        if let Some(value) = value {
            // We get better performance by using the conversion of a path into a buf directly,
            // instead of relying on iteration below. This is because the buf knows it doesn't have
            // to parse anything.
            if let Ok(path) = value.extract::<PyRef<PyLTreePath>>() {
                Ok(PyLTreePathBuilder {
                    buf: LTreePathBuf::from(path.0.clone()),
                })
            } else if let Ok(iter) = value.try_iter() {
                buf_from_iterator(&iter, py).map(|buf| PyLTreePathBuilder { buf })
            } else {
                Err(PyTypeError::new_err(
                    "Argument should be empty or an iterable of strings",
                ))
            }
        } else {
            Ok(Self::new_empty())
        }
    }

    fn append(&mut self, s: &str) -> PyResult<()> {
        self.buf.push_label(s.to_owned()).map_err(to_py_err)
    }

    fn pop(&mut self) -> Option<String> {
        self.buf.pop()
    }

    fn to_path(&self) -> PyLTreePath {
        let path = self.buf.to_ltree_path();

        PyLTreePath(path)
    }
}

#[pymodule]
fn pyltree_core(m: &Bound<'_, PyModule>) -> PyResult<()> {
    m.add_class::<PyLTreePath>()?;
    m.add_class::<PyLTreePathBuilder>()?;
    m.add("ParseError", m.py().get_type::<PyParseError>())?;

    Ok(())
}

fn buf_from_iterator(iter: &Bound<'_, PyIterator>, py: Python<'_>) -> PyResult<LTreePathBuf> {
    let mut buf = LTreePathBuf::new();

    for next in iter {
        match next {
            Ok(item) => {
                if let Ok(string_item) = item.cast::<PyString>() {
                    // If you are curious as to why to_cow can fail here: Python strings have varying
                    // internal representation depending on the content. It could thus be UCS-2 and
                    // we might have lone surrogate code units which is the one thing we can't
                    // convert into UTF-8.
                    let str_ref = string_item.to_cow()?;
                    buf.push_label(str_ref.into_owned()).map_err(to_py_err)?;
                } else {
                    return Err(PyTypeError::new_err("iterator returned non-string item"));
                }
            }
            Err(e) => {
                if e.is_instance_of::<PyStopIteration>(py) {
                    break;
                } else {
                    return Err(e);
                }
            }
        }
    }

    Ok(buf)
}

fn to_py_err(err: LTreePathParseError) -> PyErr {
    PyParseError::new_err(err.to_string())
}

fn map_to_py_result(result: LTreeParseResult<LTreePath>) -> PyResult<PyLTreePath> {
    result.map(|p| PyLTreePath(p)).map_err(to_py_err)
}
