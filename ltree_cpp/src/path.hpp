#pragma once

#include <optional>
#include <string>
#include <string_view>

#include "ltree/error.hpp"

namespace ltree {

class Path;
struct PathResult;

/**
 * Parse the text and return a Path object.
 *
 * The text must be a valid LTree Path, and it must be valid UTF-8.
 *
 * UTF-8 validity is enforced: parsing will fail on the first decoding error.
 *
 * The syntax of an LTree Path is documented in detail on `ltree::Path`.
 *
 * If `text` is empty, the resulting Path object represents an empty path as
 * well.
 *
 * May throw a std::exception (notably on memory errors).
 *
 * Returns:
 *  The same errors as `ltree::valid_utf8_path`
 */
auto MakePath(std::string_view text) -> PathResult;

/**
 * Join two paths and return the result.
 *
 * Empty paths behave like an identity element; joining a empty path with a
 * non-empty path returns the non-empty path.
 * ```cpp
 * // TODO EXAMPLE
 * ```
 *
 * This operation has `O(n)` time complexity, where n is the total length of
 * the final path.
 *
 * May throw a std::exception (notably on memory errors).
 *
 * Returns:
 *  PATH_TOO_LONG
 *    the combined total string length exceeds `MAX_PATH_LABEL_COUNT`
 */
auto Join(Path const &a, Path const &b) -> PathResult;

/**
 * Represents a ltree path, as used in the PostgreSQL `ltree` extension.
 *
 * *Note: This targets PostgreSQL 16+.*
 *
 * # Usage
 * ```cpp
 * auto path_result = ltree::MakePath("a.short.path");
 * if (ltree::Ok(path_result)) {
 *  // ltree promises this should be valid
 *  auto path = *path_result.path;
 *  DoStuff(path);
 * } else {
 *  // ...
 * }
 * ```
 * As shown in the example, if a result is "Ok", the result's optional is valid.
 *
 * One can also do:
 * ```cpp
 * auto path_result = ltree::MakePath("a.short.path");
 * if (ltree::Ok(path_result)) {
 *  auto path = path_result.path.value();
 *  DoStuff(path);
 * } else {
 *  // ...
 * }
 * ```
 * In this case, an exception will be thrown if the optional `path` is not set.
 * This is a little bit safer as it guards against bugs in `ltree`. Furthermore,
 * one can avoid the exception by checking `path.has_value()`.
 *
 * The length of a path is defined as the count of Unicode Scalar Values and
 * delimiters in it. This is equivalent to the number of codepoints after string
 * conversion.
 *
 * # Performance
 * Creating a path from a string involves a parsing step; this operation is
 * linear to the number of bytes in the string.
 *
 * Joining two paths is fairly cheap, and converting one to a string is almost
 * a no-op. Similarly, it is virtually free to acquire a string slice.
 *
 * Due to the storage format, iteration over labels is fairly expensive. If
 * there is a need for extensive manipulation, in the middle or beginning of a
 * path, it is likely better to work with a sequence of labels. When all is
 * said and done, they can be joined cheaply.
 *
 * # Paths
 * A path consists of ltree labels, separated by dots (`.` U+002E).
 *
 * A label is a non-zero sequence of Unicode Scalar Values. Only alphanumeric
 * codepoints, minus characters (`-` U+002D) and underscores (`_` U+005F) are
 * allowed in a label.
 *
 * In PostgreSQL 16, the exact meaning of alphanumeric is determined by the
 * database locale, but this library accepts any value Unicode deems
 * alphanumeric.
 *
 * # Grammar
 * The exact grammar for a string path is given in EBNF:
 * ```text
 * Path            = [{Label, dot}, Label];
 * Label           = element, {element};
 *
 * element         = alphanumeric | minus | underscore;
 * alphanumeric    = ? any Unicode alphanumeric scalar value ?;
 * minus           = ? U+002D ?;
 * underscore      = ? U+005F ?;
 * dot             = ? U+002E ?;
 * ```
 *
 * *Note: paths can actually be empty.*
 *
 * # References
 * - `contrib / ltree / ltree_io.c` in the PostgreSQL trunk.
 * - https://www.postgresql.org/docs/16/ltree.html
 *
 * # Compatibility notes
 * - PostgreSQL 15 does not allow the minus-hyphen in a label, while PostgreSQL
 * 16 does.
 * - PostgreSQL 15 has a maximum label size of 255 "characters".
 */
class Path {
  public:
    Path(Path const &) = default;
    Path(Path &&) noexcept = default;
    Path &operator=(Path const &) = default;
    Path &operator=(Path &&) = default;

    /**
     * Whether this represents the empty path.
     *
     * This operation has `O(1)` time complexity.
     */
    [[nodiscard]] auto Empty() const noexcept -> bool;

    /**
     * Get the path as a string.
     *
     * This operation has `O(n)` time complexity.
     *
     * Exceptions:
     *  std::exception
     */
    auto ToString() const -> std::string;

    bool operator==(Path const &) const noexcept;
    bool operator!=(Path const &) const noexcept;

  private:
    using string_type = std::string;
    using size_type = std::size_t;

    friend auto MakePath(std::string_view text) -> PathResult;
    friend auto Join(Path const &a, Path const &b) -> PathResult;

    Path(string_type &&text, size_type label_count);

    string_type text_;
    size_type label_count_;
    // NOTE(siot): Potential future optimisation:
    // vector<size_type> label_offsets_
    // We could use this quickly to iterate through the labels.
};

struct PathResult {
    // Status of the operation.
    Error::ValueType code;
    // (if "ok") this holds the result of the operation
    std::optional<Path> path;
    // (if not "ok") The byte index (if any) of any invalid data.
    std::size_t error_column;
};

/**
 * The result is considered "ok" if `code == error::SUCCESS`.
 */
[[nodiscard]] constexpr auto Ok(PathResult const &result) noexcept -> bool {
    return result.code == Error::SUCCESS;
}

/**
 * Generate a descriptive error message for the given result.
 */
auto ErrorMessage(PathResult const &) -> std::string;

} // namespace ltree
