#include <algorithm>
#include <cassert>
#include <fstream>
#include <tuple>
#include <vector>

#include <catch2/catch_test_macros.hpp>
#include <nlohmann/json.hpp>

#include "internal/config.hpp"
#include "ltree/parse.hpp"
#include "ltree/path.hpp"

using json = nlohmann::json;

namespace {

using LTreeLabels = std::vector<std::string>;

struct ParsePassCase {
    std::string name;
    std::string input;
    LTreeLabels expected;
};
NLOHMANN_DEFINE_TYPE_NON_INTRUSIVE(ParsePassCase, name, input, expected);

struct ParseFailCase {
    std::string name;
    std::string input;
    std::string message;
};
NLOHMANN_DEFINE_TYPE_NON_INTRUSIVE(ParseFailCase, name, input, message);

struct StringConvPassCase {
    std::string name;
    LTreeLabels input;
    std::string expected;
};
NLOHMANN_DEFINE_TYPE_NON_INTRUSIVE(StringConvPassCase, name, input, expected);

struct JoinPassCase {
    std::string name;
    std::tuple<LTreeLabels, LTreeLabels> input;
    std::string expected;
};
NLOHMANN_DEFINE_TYPE_NON_INTRUSIVE(JoinPassCase, name, input, expected);

} // namespace

auto LoadCommonTestSuite() -> json {
    static char const *COMMON_TEST_SUITE_PATH_RELATIVE =
        "common/test_cases.json";

    std::ifstream f(COMMON_TEST_SUITE_PATH_RELATIVE);

    return json::parse(f);
}

auto JoinLabels(std::vector<std::string> const &labels) -> std::string {
    bool first = true;
    std::string joined{};
    for (auto label : labels) {
        if (!first) {
            joined.push_back(ltree::Token::DELIMITER_CHAR);
        } else {
            first = false;
        }
        joined.append(label);
    }
    return joined;
}

TEST_CASE("paths_MakePath_common_cases") {
    auto common = LoadCommonTestSuite();

    auto parse_text_cases = common.at("parse_text_cases");

    auto pass_cases =
        parse_text_cases.at("pass").template get<std::vector<ParsePassCase>>();
    for (auto pass_case : pass_cases) {
        CAPTURE(pass_case.name, pass_case.input);

        auto result = ltree::MakePath(pass_case.input);
        CHECK(Ok(result));

        auto actual = result.path.value().ToString();
        auto expected = JoinLabels(pass_case.expected);

        CHECK(expected == actual);
    }

    auto fail_cases =
        parse_text_cases.at("fail").template get<std::vector<ParseFailCase>>();
    for (auto fail_case : fail_cases) {
        CAPTURE(fail_case.name, fail_case.input);

        auto result = ltree::MakePath(fail_case.input);
        CHECK(!Ok(result));

        auto actual_message = ltree::ErrorMessage(result);

        CHECK(fail_case.message == actual_message);
    }
}

TEST_CASE("paths_Path_to_string_common_cases") {
    auto common = LoadCommonTestSuite();

    auto string_conv_cases = common.at("string_conv_cases");

    auto pass_cases = string_conv_cases.at("pass")
                          .template get<std::vector<StringConvPassCase>>();
    for (auto pass_case : pass_cases) {
        CAPTURE(pass_case.name, pass_case.input);

        auto result = ltree::MakePath(JoinLabels(pass_case.input));
        REQUIRE(Ok(result));

        auto actual = result.path.value().ToString();
        auto expected = pass_case.expected;

        CHECK(expected == actual);
    }

    // We don't expect error cases.
    CHECK(string_conv_cases.at("fail").empty());
}

TEST_CASE("paths_Join_common_cases") {
    auto common = LoadCommonTestSuite();

    auto join_cases = common.at("join_cases");

    auto pass_cases =
        join_cases.at("pass").template get<std::vector<JoinPassCase>>();
    for (auto pass_case : pass_cases) {
        CAPTURE(pass_case.name, pass_case.input);

        auto [labels1, labels2] = pass_case.input;
        auto path_str1 = JoinLabels(labels1);
        auto path_str2 = JoinLabels(labels2);

        auto path1 = ltree::MakePath(path_str1).path.value();
        auto path2 = ltree::MakePath(path_str2).path.value();

        auto expected = pass_case.expected;

        auto join = ltree::Join(path1, path2);
        REQUIRE(Ok(join));

        CHECK(expected == join.path.value().ToString());
    }

    // We don't expect error cases.
    CHECK(join_cases.at("fail").empty());
}

TEST_CASE("paths_MakePath_good_path_sets_results_correctly") {
    char const *path_str = "this.path-should.be.valid123";
    auto result = ltree::MakePath(path_str);

    CHECK(result.code == ltree::Error::SUCCESS);
    REQUIRE(result.path.has_value());
    auto path = result.path.value();

    auto actual_path_str = path.ToString();
    CHECK(path_str == actual_path_str);
}

TEST_CASE("paths_MakePath_bad_path_sets_results_correctly") {
    char const *path_str = "this.path should.be!!!invalid123";
    auto result = ltree::MakePath(path_str);

    CHECK(result.code == ltree::Error::INVALID_CHARACTER);
    CHECK(result.error_column == 9);
    CHECK(!result.path.has_value());
}

auto GenerateRepeatingPath(std::size_t label_count) -> std::string {
    assert(label_count >= 1);

    char const label = 'a';

    std::string path_str{};
    path_str.reserve(label_count); // Approximation

    path_str.push_back(label);
    for (std::size_t i = 1; i < label_count; i++) {
        path_str.push_back(ltree::Token::DELIMITER_CHAR);
        path_str.push_back(label);
    }
    return path_str;
}

TEST_CASE("paths_MakePath_too_many_labels") {
    // This test might take a very long time and eat a lot of resources.
    // Redesign it if this assert fails.
    assert(config_internal::MAX_PATH_LABEL_COUNT < 1 << 30);

    auto path_str =
        GenerateRepeatingPath(config_internal::MAX_PATH_LABEL_COUNT);
    auto should_be_ok = ltree::MakePath(path_str);
    CHECK(Ok(should_be_ok));

    path_str = GenerateRepeatingPath(config_internal::MAX_PATH_LABEL_COUNT + 1);
    auto should_be_bad = ltree::MakePath(path_str);
    CHECK(should_be_bad.code == ltree::Error::PATH_TOO_LONG);
}

TEST_CASE("paths_Join_too_many_labels") {
    assert(config_internal::MAX_PATH_LABEL_COUNT < 1 << 30);

    auto path_str =
        GenerateRepeatingPath(config_internal::MAX_PATH_LABEL_COUNT);
    auto almost_too_long = ltree::MakePath(path_str).path.value();

    auto single_label = ltree::MakePath("b").path.value();

    auto should_be_bad = ltree::Join(almost_too_long, single_label);
    CHECK(should_be_bad.code == ltree::Error::PATH_TOO_LONG);
}

TEST_CASE("paths_Path_equality") {
    char const *path_str1 = "an_interesting.path123";
    char const *path_str2 = path_str1;

    auto path1 = ltree::MakePath(path_str1).path.value();
    auto path2 = ltree::MakePath(path_str2).path.value();

    CHECK(path1 == path1);
    CHECK(path2 == path1);
    CHECK(path1 == path2);

    char const *path_str3 = "a-different.path";
    auto path3 = ltree::MakePath(path_str3).path.value();

    CHECK(path1 != path3);
    CHECK(path3 != path1);

    auto empty1 = ltree::MakePath("").path.value();
    auto empty2 = ltree::MakePath("").path.value();

    REQUIRE(empty1.Empty());
    REQUIRE(empty2.Empty());
    CHECK(empty1 == empty2);
}