#pragma once

#include <cstddef> // std::size_t

namespace config_internal {

/// A path may not contain more labels than this.
inline constexpr std::size_t MAX_PATH_LABEL_COUNT = 65535;

/// A label may not contain more Unicode Scalar Values than this.
inline constexpr std::size_t MAX_LABEL_USV_COUNT = 1000;

} // namespace config_internal
