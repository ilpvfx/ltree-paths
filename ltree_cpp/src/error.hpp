#pragma once

namespace ltree {

namespace Error {
using ValueType = int;
inline constexpr ValueType SUCCESS = 0;
inline constexpr ValueType EXPECTED_LABEL = 1;
inline constexpr ValueType INVALID_CHARACTER = 2;
inline constexpr ValueType LABEL_TOO_LONG = 3;
inline constexpr ValueType PATH_TOO_LONG = 4;
// C++ strings can contain whatever bytes the user desires, and this might give
// us problems, for example:
// - lone surrogate halves
// - broken continuation sequences in UTF-8
inline constexpr ValueType BAD_UNICODE = 5;
} // namespace Error

} // namespace ltree
