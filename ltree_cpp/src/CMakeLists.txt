cpp_library(ltree
    PUBLIC_HEADERS_INSTALL_PREFIX
        "ltree"
    PUBLIC_HEADERS
        "error.hpp"
        "parse.hpp"
        "path.hpp"
    CPPFILES
        "parse.cpp"
        "path.cpp"
        "internal/error_msg.cpp"
        "internal/unicode.cpp"
    LIBRARIES
    INCLUDE_DIRS
        "../external/"
    DEFINES
)

cpp_test(ltree_test
    CPPFILES
        "tests/test_interface_parsing.cpp"
        "tests/test_paths.cpp"
    LIBRARIES
        ltree
        Catch2::Catch2WithMain
        nlohmann_json::nlohmann_json
    INCLUDE_DIRS
        "${CMAKE_CURRENT_LIST_DIR}"
)

include(GNUInstallDirs)

set(VERSION_FILE_CONFIGURE_PATH "${PROJECT_BINARY_DIR}/include/ltree/version.hpp")
configure_file(
    "version.hpp.in" ${VERSION_FILE_CONFIGURE_PATH}
)
install(FILES
  ${VERSION_FILE_CONFIGURE_PATH}
  DESTINATION "${CMAKE_INSTALL_INCLUDEDIR}/ltree/"
)


if(BUILD_TESTING) 
    # NOTE (siot): unsure if this is meaningful. It's only helpful if we run from inside "build/src".
    file(COPY "${PROJECT_SOURCE_DIR}/../common/test_cases.json" DESTINATION "common")
endif()
