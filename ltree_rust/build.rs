use std::path::Path;

const COMMON_TEST_SUITE_FILENAME: &'static str = "test_cases.json";

fn main() {
    // Keep the library in sync with this name.
    println!(
        "cargo:rustc-env=COMMON_TEST_SUITE_FILENAME={}",
        COMMON_TEST_SUITE_FILENAME
    );
    println!(
        "cargo:rustc-env=COMMON_TEST_SUITE_PATH_RELATIVE={}",
        Path::new("../common")
            .join(COMMON_TEST_SUITE_FILENAME)
            .to_str()
            .expect("this should be a simple ASCII path")
    );
}
