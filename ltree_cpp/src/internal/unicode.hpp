#pragma once

#include <cstddef> // char32_t

namespace unicode_internal {

/**
 * This checks if the category is in any of the following:
 *
 *      Letters: Lu, Ll, Lt, Lm, Lo
 *      Numbers: Nd, Nl, No
 *
 * This does not check whether the codepoint is "valid" in any sense. For
 * example, non-assigned characters map to "unknown" and hence `false`.
 *
 * Note:
 *  The lookup depends on the categories exported by `scripts/unicode_ranges`.
 *
 * Returns:
 *  false
 *      codepoint invalid or not in the supported list
 *  true
 *      otherwise
 */
[[nodiscard]] auto UnicodeIsAlphanumeric(char32_t codepoint) noexcept -> bool;

} // namespace unicode_internal
