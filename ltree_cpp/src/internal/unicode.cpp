#include "unicode.hpp"

namespace unicode_internal {

struct unicode_range {
    char32_t start;
    // End is not inclusive.
    char32_t end;
};

#include "unicode_alphanumeric_ranges.incl"

auto UnicodeIsAlphanumeric(char32_t codepoint) noexcept -> bool {
    // NOTE(siot): We assume the list is sorted by range start, in ascending
    // order.
    for (std::size_t i = 0; i < ALPHANUMERIC_RANGES_COUNT; ++i) {
        if (ALPHANUMERIC_RANGES[i].start <= codepoint &&
            codepoint < ALPHANUMERIC_RANGES[i].end) {
            return true;
        }
    }
    return false;
}

} // namespace unicode_internal
