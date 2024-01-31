use std::{
    fs::{self},
    iter::repeat,
    result::Result,
};

use paste::paste;
use serde::Deserialize;

use super::*;

//
// A simpler type that is easier to serialize & deserialize than the LTreePath itself.
//
#[derive(Deserialize, Debug, PartialEq, Eq, Clone)]
#[serde(transparent)]
struct LTreeLabels {
    labels: Vec<String>,
}

impl TryInto<LTreePath> for LTreeLabels {
    type Error = LTreePathParseError;

    fn try_into(self) -> Result<LTreePath, Self::Error> {
        let s = self.labels.join(".");
        s.try_into()
    }
}

const COMMON_TEST_SUITE_PATH_RELATIVE: &'static str = env!("COMMON_TEST_SUITE_PATH_RELATIVE");

macro_rules! make_case_structs {
    ($name:ident; $input:ident : $input_type:ty, $expected:ident : $expected_type:ty) => {
        paste! {
            #[derive(Deserialize, Debug)]
            struct [<$name PassCase>] {
                name: String,
                $input: $input_type,
                $expected: $expected_type,
            }

            #[allow(dead_code)]
            #[derive(Deserialize, Debug)]
            struct [<$name FailCase>] {
                name: String,
                $input: $input_type,
                message: String,
            }

            #[derive(Deserialize, Debug)]
            struct [<$name Cases>] {
                pass: Vec<[<$name PassCase>]>,
                fail: Vec<[<$name FailCase>]>,
            }
        }
    };
}

make_case_structs! {ParseText; input: String, expected: LTreeLabels}
make_case_structs! {StringConv; input: LTreeLabels, expected: String}
make_case_structs! {Join; input: (LTreeLabels, LTreeLabels), expected: String}

#[derive(Deserialize, Debug)]
struct CommonTestSuite {
    parse_text_cases: ParseTextCases,
    string_conv_cases: StringConvCases,
    join_cases: JoinCases,
}

macro_rules! print_pass_case {
    ($pass_case:ident) => {
        println!(
            "SHOULD PASS: {}\n\tinput = {:?}\n\texpected = {:?}",
            $pass_case.name, $pass_case.input, $pass_case.expected
        );
    };
}

macro_rules! print_fail_case {
    ($fail_case:ident) => {
        println!(
            "SHOULD FAIL: {}\n\tmessage = {}",
            $fail_case.name, $fail_case.message
        );
    };
}

fn run_parse_tests(cases: &ParseTextCases) {
    for pass_case in &cases.pass {
        print_pass_case!(pass_case);
        let path = LTreePath::from_str(&pass_case.input).expect("a valid LTree path");

        let simple_path = LTreeLabels {
            labels: path.labels().map(ToOwned::to_owned).collect(),
        };
        assert_eq!(pass_case.expected, simple_path);
    }

    for fail_case in &cases.fail {
        print_fail_case!(fail_case);

        let err = LTreePath::from_str(&fail_case.input).expect_err("this path should fail");
        assert_eq!(fail_case.message, err.to_string());
    }
}

fn run_string_conv_tests(cases: &StringConvCases) {
    for pass_case in &cases.pass {
        print_pass_case!(pass_case);
        let path: LTreePath = pass_case
            .input
            .clone()
            .try_into()
            .expect("a valid LTree path");

        assert_eq!(pass_case.expected, path.as_str());
    }

    for _ in &cases.fail {
        assert!(true, "There are no known failure cases.")
    }
}

fn run_join_tests(cases: &JoinCases) {
    for pass_case in &cases.pass {
        print_pass_case!(pass_case);
        let (a_labels, b_labels) = pass_case.input.clone();
        let a: LTreePath = a_labels.try_into().expect("a valid LTree path");
        let b: LTreePath = b_labels.try_into().expect("a valid LTree path");

        assert_eq!(pass_case.expected, a.join(&b).unwrap().as_str());
    }

    for _ in &cases.fail {
        assert!(true, "There are no known failure cases.")
    }
}

#[test]
fn test_common_suite() {
    println!(
        "Using {} as the path to the test data.",
        COMMON_TEST_SUITE_PATH_RELATIVE
    );

    let test_data_str = fs::read_to_string(COMMON_TEST_SUITE_PATH_RELATIVE)
        .expect("the test data should be located in the working directory");

    let suite: CommonTestSuite = serde_json::from_str(&test_data_str)
        .expect("the JSON data should be structured appropriately");

    run_parse_tests(&suite.parse_text_cases);
    run_string_conv_tests(&suite.string_conv_cases);
    run_join_tests(&suite.join_cases);
}

// A hedge for the future, we shouldn't use gigabytes of memory for a simple test.
macro_rules! is_path_label_count_test_reasonable {
    () => {
        assert!(
            LTreePath::MAX_PATH_LABEL_COUNT < 1 << 30,
            "This test might take a very long time and eat a lot of resources. Redesign it."
        );
    };
}

#[test]
fn test_try_from_string_too_many_labels() {
    is_path_label_count_test_reasonable!();

    let too_long_path = repeat("A")
        .take(LTreePath::MAX_PATH_LABEL_COUNT + 1)
        .collect::<Vec<&str>>()
        .join(LTreePath::DELIMITER_STR);

    assert_eq!(
        Err(LTreePathParseError::PathTooLong),
        LTreePath::from_str(&too_long_path)
    );
    assert_eq!(
        Err(LTreePathParseError::PathTooLong),
        LTreePath::try_from(too_long_path)
    );
}

#[test]
fn test_join_too_many_labels() {
    is_path_label_count_test_reasonable!();

    let almost_too_long_path: String = repeat("A")
        .take(LTreePath::MAX_PATH_LABEL_COUNT)
        .collect::<Vec<&str>>()
        .join(LTreePath::DELIMITER_STR);

    let maximal_path = LTreePath::try_from(almost_too_long_path)
        .expect("this is just under the maximum label count");

    assert_eq!(
        Err(LTreePathParseError::PathTooLong),
        maximal_path.join(&LTreePath::from_str("B").expect("a valid label"))
    );
}

#[test]
fn test_buf_empty_buf() {
    let buf = LTreePathBuf::new();

    let path = buf.to_ltree_path();
    assert!(
        path.is_empty(),
        "An empty buf object should yield a valid empty path"
    );
}

#[test]
fn test_buf_push_label() {
    let mut buf = LTreePathBuf::new();

    // Test valid labels
    assert_eq!(Ok(()), buf.push_label("a_b123".to_owned()));
    assert_eq!(Ok(()), buf.push_label("123_hello-yo".to_owned()));

    // Test invalid labels
    assert_eq!(
        Err(LTreePathParseError::InvalidCharacter(0)),
        buf.push_label("$abc".to_owned())
    );
    let label_too_long = String::from_iter(repeat('a').take(LTreePath::MAX_LABEL_USV_COUNT + 1));
    assert_eq!(
        Err(LTreePathParseError::LabelTooLong),
        buf.push_label(label_too_long)
    );

    let path = buf.to_ltree_path();
    let expected = LTreePath::from_str("a_b123.123_hello-yo").unwrap();
    assert_eq!(expected, path);

    // Empty corner cases.
    let mut buf = LTreePathBuf::new();
    buf.push_label("a".to_owned()).unwrap();
    buf.push_label("".to_owned()).unwrap();
    assert_eq!("a", buf.to_ltree_path().to_string());

    let mut buf = LTreePathBuf::new();
    buf.push_label("".to_owned()).unwrap();
    buf.push_label("a".to_owned()).unwrap();
    assert_eq!("a", buf.to_ltree_path().to_string());

    let mut buf = LTreePathBuf::new();
    buf.push_label("".to_owned()).unwrap();
    buf.push_label("".to_owned()).unwrap();
    assert_eq!("", buf.to_ltree_path().to_string());
}

#[test]
fn test_buf_push_label_too_many() {
    is_path_label_count_test_reasonable!();

    let mut buf = LTreePathBuf::new();

    for _ in 0..LTreePath::MAX_PATH_LABEL_COUNT {
        buf.push_label("a".to_owned())
            .expect("should not be too many");
    }

    assert_eq!(
        Err(LTreePathParseError::PathTooLong),
        buf.push_label("b".to_owned())
    );
}

#[test]
fn test_buf_pop() {
    let mut buf = LTreePathBuf::new();
    assert_eq!(None, buf.pop());

    buf.push_label("a".to_owned()).unwrap();
    assert_eq!(Some("a".to_owned()), buf.pop());
    assert_eq!(None, buf.pop());
}

#[test]
fn test_buf_from_path() {
    let path = LTreePath::from_str("a.long.string").unwrap();
    let buf = LTreePathBuf::from(path.clone());

    assert_eq!(path, buf.to_ltree_path());
}