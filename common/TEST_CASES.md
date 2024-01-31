# Test Cases
These are common test cases, useful for all the different ports. 

The tests cover these essential operations:
- Parsing (text representation)
- String conversion
- Joining two validated paths

All strings are UTF-8.

## Omitted cases
There are some tests that are impractical to specify this way. They should be tested directly in the port's tests:
- The maximum number of labels is 65535. This can be easily tested with a for loop.

Some implementations may support the binary format. There are no common tests for it yet.


## Test Cases Format
*TODO (siot): JSON Schema perhaps?* 

### Root Format
The root object should be a JSON object with the following members:
- `parse_text_cases`
- `string_conversion_cases`
- `join`

### Pass and Fail Test Cases
Each set of cases contains a list of `pass` tests and a list of `fail` tests.

|  Field Name       |  Type      |  Comment                                                        |
|-------------------|------------|-----------------------------------------------------------------|
| name              | string     | Keep it descriptive and machine processing friendly.            |
| input             | *          | See below                                                       |

The type of `input` is described in the following table:

|  Operation        |  Type      |  Comment                                                        |
|-------------------|------------|-----------------------------------------------------------------|
| parse text        | string     |                                                                 |
| string conversion | LTreePath  |                                                                 |
| join              | [ LTreePath, LTreePath ]  | The paths to join                                |

### Pass Cases
Contains the additional fields:

|  Field Name       |  Type      |  Comment                                                        |
|-------------------|------------|-----------------------------------------------------------------|
| expected          | *          | The output should match this exactly.                           |

The type of `expected` is described in the following table:

|  Operation        |  Type      |  Comment                                                        |
|-------------------|------------|-----------------------------------------------------------------|
| parse text        | LTreePath  | See below                                                       |
| string conversion | string     |                                                                 |
| join              | string     |                                                                 |

### Fail Cases
Contains the additional fields:

|  Field Name       |  Type      |  Comment                                                        |
|-------------------|------------|-----------------------------------------------------------------|
| message           | string     | The error message should match this exactly.                    |

### LTreePath
This is a list of strings with an implied separator.

For example, `a.b.c` is represented as `["a", "b", "c"]`. An empty list represents an empty LTree Path.

Two objects `a` and `b` are equal when they contain the same number of strings and `a[i] == b[i]` 
for `i=0,...,len(a)-1`.