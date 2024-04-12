#include "ltree/path.hpp"

#include <cassert>
#include <limits> // std::numeric_limits
#include <stdexcept>
#include <type_traits>

#include "internal/config.hpp"
#include "internal/error_msg.hpp"
#include "ltree/parse.hpp"

namespace ltree {

auto MakePath(std::string_view text) -> PathResult {
    const auto parse = Utf8ValidPath(text);

    PathResult result{};
    result.code = parse.code;
    result.error_column = parse.error_column;

    if (parse.code == Error::SUCCESS) {
        // May throw!
        result.path.emplace(Path(std::string{text}, parse.label_count));
    }

    return result;
}

auto Join(Path const &a, Path const &b) -> PathResult {
    PathResult result{};

    // Note (siot): Hypothetically we could overflow here. In practice
    // this would require the limits to be increased and a 32-bit system
    // (or less). It is not expected that we'll have strings large
    // enough (approx 2^63) on a 64-bit system, especially since the
    // number of addressable bytes is far less.
    // Maybe this comment will come back and haunt me when I am retired.
    #ifndef NDEBUG
    static constexpr std::size_t MAX_STRING_SIZE =
        std::numeric_limits<std::size_t>::max();
    #endif
    assert(a.text_.length() < MAX_STRING_SIZE - b.text_.length());

    assert(a.label_count_ <= config_internal::MAX_PATH_LABEL_COUNT);
    assert(b.label_count_ <= config_internal::MAX_PATH_LABEL_COUNT);

    if (a.label_count_ >
        config_internal::MAX_PATH_LABEL_COUNT - b.label_count_) {
        result.code = Error::PATH_TOO_LONG;
    } else {
        std::string new_path;

        // May throw!
        new_path.reserve(a.text_.length() + b.text_.length() + 1);
        new_path.append(a.text_);
        if (!a.Empty() && !b.Empty()) {
            new_path.push_back(Token::DELIMITER_CHAR);
        }
        new_path.append(b.text_);

        auto new_label_count = a.label_count_ + b.label_count_;

        result.path.emplace(Path(std::move(new_path), new_label_count));
    }

    return result;
}

// Private constructor, it assumes the text is already validated.
Path::Path(Path::string_type &&text, Path::size_type label_count)
    : text_{std::move(text)}, label_count_{label_count} {}

auto Path::ToString() const -> std::string { return text_; }

auto Path::Empty() const noexcept -> bool { return text_.empty(); }

auto Path::operator==(Path const &other) const noexcept -> bool {
    return text_ == other.text_;
}

auto Path::operator!=(Path const &other) const noexcept -> bool {
    return text_ != other.text_;
}

auto ErrorMessage(PathResult const &result) -> std::string {
    return error_msg_internal::GenerateMessage(result.code,
                                               result.error_column);
}

} // namespace ltree
