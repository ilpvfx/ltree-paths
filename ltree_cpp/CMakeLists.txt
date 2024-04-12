cmake_minimum_required(VERSION 3.16)

# Supress warnings about DOWNLOAD_EXTRACT_TIMESTAMP in FetchContent
cmake_policy(SET CMP0135 NEW)

file(STRINGS ../VERSION proj_version)
project(ltree VERSION ${proj_version})

set(LTREE_VERSION_MAJOR ${PROJECT_VERSION_MAJOR})
set(LTREE_VERSION_MINOR ${PROJECT_VERSION_MINOR})
set(LTREE_VERSION_PATCH ${PROJECT_VERSION_PATCH})
set(LTREE_VERSION       ${PROJECT_VERSION})

list(APPEND CMAKE_MODULE_PATH 
    ${CMAKE_CURRENT_LIST_DIR}/cmake
    ${CMAKE_CURRENT_LIST_DIR}/cmake/fetch
)

# Exposes build options (with default values) for this project.
include(Options)

# Set project defaults.
include(Defaults)

# Import package dependencies.
include(Packages)

# Include CMake build utilities for USD libraries/plugins.
include(BuildTools) 

add_subdirectory(src)

# Set project defaults.
include(Export)