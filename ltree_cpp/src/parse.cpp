#include "ltree/parse.hpp"

#include <utfcpp/utf8.h>

#include "internal/config.hpp"
#include "internal/error_msg.hpp"
#include "internal/unicode.hpp"

struct ParsingState {
    std::size_t bytes_accepted;
    std::size_t usvs_accepted;
    std::size_t labels_accepted;
    ltree::Error::ValueType status;
};

/**
 * Parse the input as a label until either the end is reached, or an error
 * occurs.
 *
 * Errors occur on
 * - start == end
 * - the first character is not a valid label character
 * - a character is neither a valid label character nor a path delimiter
 * - UTF-8 decode failed
 *
 * `state.status` is set to an appropriate error on failure.
 *
 * This does not update `labels_accepted`.
 *
 * Returns:
 *  an iterator pointing to the byte right after the last valid codepoint
 */
template <typename TIterator>
static auto NextLabel(TIterator start, TIterator end, ParsingState &state)
    -> TIterator {
    if (start == end) {
        state.status = ltree::Error::EXPECTED_LABEL;
        return end;
    }

    state.status = ltree::Error::SUCCESS;
    std::size_t current_label_usv_count = 0;

    TIterator it = start;
    try {
        while (it != end) {
            auto cp = utf8::peek_next(it, end);
            static_assert(std::is_same_v<decltype(cp), char32_t>);

            if (!ltree::ValidLabelCodepoint(cp) && cp != ltree::Token::DELIMITER) {
                state.status = ltree::Error::INVALID_CHARACTER;
                break;
            } else if (cp == ltree::Token::DELIMITER) {
                break;
            }

            auto previous_codepoint = it;
            utf8::next(it, end);
            state.bytes_accepted += std::distance(previous_codepoint, it);
            current_label_usv_count += 1;
            state.usvs_accepted += 1;

            if (current_label_usv_count >
                config_internal::MAX_LABEL_USV_COUNT) {
                state.status = ltree::Error::LABEL_TOO_LONG;
                break;
            }
        }
    } catch (utf8::exception &) {
        /*
           Note: this may hide a bug in our code! We hope the tests catch this.
           If we give incorrect start and end iterators to the utfcpp library,
           it might throw not_enough_room. This error is also thrown when we
           are given a UTF-8 continuation sequence that ends too soon.
        */
        state.status = ltree::Error::BAD_UNICODE;
    }

    // First character was a dud.
    if (state.status == ltree::Error::SUCCESS && it == start) {
        state.status = ltree::Error::EXPECTED_LABEL;
    }

    return it;
}

namespace ltree {

auto ValidLabelCodepoint(char32_t codepoint) noexcept -> bool {
    return unicode_internal::UnicodeIsAlphanumeric(codepoint) ||
           codepoint == Token::MINUS || codepoint == Token::UNDERSCORE;
}

auto Utf8ValidLabel(std::string_view text) -> ParseResult {
    auto const text_begin = std::cbegin(text);
    auto const text_end = std::cend(text);
    auto it = text_begin;

    ParsingState state{};
    NextLabel(it, text_end, state);

    ParseResult result{};
    result.code = state.status;
    result.error_column = state.bytes_accepted;
    result.label_count = 1;

    return result;
}

auto Utf8ValidPath(std::string_view text) -> ParseResult {
    auto const text_begin = std::cbegin(text);
    auto const text_end = std::cend(text);
    auto it = text_begin;

    if (it == text_end) {
        return ParseResult{};
    }

    ParsingState state{};
    bool expecting_label = true;

    while (it != text_end) {
        if (state.labels_accepted > config_internal::MAX_PATH_LABEL_COUNT) {
            state.status = Error::PATH_TOO_LONG;
            break;
        }

        if (expecting_label) {
            it = NextLabel(it, text_end, state);

            expecting_label = false;
        } else {
            // The delimiter is ASCII and as such fits into one UTF-8 code unit
            // (a byte). We don't need to decode it.
            if (state.status == Error::SUCCESS && *it == Token::DELIMITER) {
                expecting_label = true;
                state.bytes_accepted += 1;
                state.usvs_accepted += 1;
                state.labels_accepted += 1;
                it++;
            } else {
                break;
            }
        }
    }

    if (state.status == Error::SUCCESS) {
        // We ended on a delimiter, which is not allowed.
        if (expecting_label) {
            state.status = Error::EXPECTED_LABEL;
        } else {
            state.labels_accepted += 1;
        }

        if (state.labels_accepted > config_internal::MAX_PATH_LABEL_COUNT) {
            state.status = Error::PATH_TOO_LONG;
        }
    }

    ParseResult result{};
    result.code = state.status;
    result.error_column = state.bytes_accepted;
    result.label_count = state.labels_accepted;

    return result;
}

auto ErrorMessage(ParseResult const &result) -> std::string {
    return error_msg_internal::GenerateMessage(result.code,
                                               result.error_column);
}

} // namespace ltree