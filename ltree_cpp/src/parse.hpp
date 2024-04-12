#pragma once

#include <cstddef> // std::size_t, char32_t
#include <string_view>

#include "ltree/error.hpp"

namespace ltree {

/**
 * See the module documentation for more information.
 */
[[nodiscard]] auto ValidLabelCodepoint(char32_t codepoint) noexcept -> bool;

namespace Token {
inline constexpr char32_t MINUS = U'\x002D';
inline constexpr char32_t UNDERSCORE = U'\x005F';
inline constexpr char32_t DELIMITER = U'\x002E';
inline constexpr char DELIMITER_CHAR = '.';
} // namespace Token

struct ParseResult {
    Error::ValueType code;
    // (if "ok") The number of labels found in the path.
    std::size_t label_count;
    // (if not "ok") The byte index (if any) of any invalid data.
    std::size_t error_column;
};

/**
 * Returns true if the given result is deemed to be successful; otherwise false.
 */
[[nodiscard]]
constexpr auto Ok(ParseResult const &result) noexcept -> bool { return result.code == Error::SUCCESS; }

/**
 * Generate a descriptive error message for the given result.
 */
auto ErrorMessage(ParseResult const &result) -> std::string;


/**
 * Decodes the text as UTF-8 and tries to ascertain whether this is a valid
 * LTree label.
 *
 * Returns:
 *  See `ParseResult`.
 */
[[nodiscard]] auto Utf8ValidLabel(std::string_view text) -> ParseResult;

/**
 * Decodes the text as UTF-8 and tries to ascertain whether this is a valid
 * LTree path.
 *
 * Note:
 *  The empty string is a valid path, but not a valid label.
 *
 * Returns:
 *  See `ParseResult`.
 */
[[nodiscard]] auto Utf8ValidPath(std::string_view text) -> ParseResult;

} // namespace ltree
