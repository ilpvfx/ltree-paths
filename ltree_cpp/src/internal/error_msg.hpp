#pragma once

#include <cstddef> // std::size_t
#include <string>  // std::string

#include "ltree/error.hpp"

namespace error_msg_internal {

[[nodiscard]] auto GenerateMessage(ltree::Error::ValueType code,
                                   std::size_t error_column) -> std::string;

} // namespace error_msg_internal
