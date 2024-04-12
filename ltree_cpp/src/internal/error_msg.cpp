#include "error_msg.hpp"

#include <sstream>

#include "config.hpp"

namespace error_msg_internal {

auto GenerateMessage(ltree::Error::ValueType code, std::size_t error_column)
    -> std::string {
    // NOTE(siot): If we bump up to C++20 we could use std::format.
    std::ostringstream oss;
    switch (code) {
    case ltree::Error::SUCCESS:
        oss << "Success";
        break;
    case ltree::Error::EXPECTED_LABEL:
        oss << "Parse error: Expected label at offset " << error_column;
        break;
    case ltree::Error::INVALID_CHARACTER:
        oss << "Parse error: Invalid character at offset " << error_column;
        break;
    case ltree::Error::LABEL_TOO_LONG:
        oss << "Parse error: Label too long (max "
            << config_internal::MAX_LABEL_USV_COUNT << " UTF-8 scalar values)";
        break;
    case ltree::Error::PATH_TOO_LONG:
        oss << "Parse error: Path too long (max "
            << config_internal::MAX_PATH_LABEL_COUNT << " labels)";
        break;
    case ltree::Error::BAD_UNICODE:
        oss << "Parse error: UTF-8 decoding failed at offset " << error_column;
        break;
    default:
        oss << "Internal Error: unknown error code";
        break;
    }
    return oss.str();
}

} // namespace error_msg_internal
