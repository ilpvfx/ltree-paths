# Default build configurations for the project.

set(CMAKE_POSITION_INDEPENDENT_CODE ON)

# By default, build for release.
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE "Release")
endif()

# By default, build using C++17.
if(NOT CMAKE_CXX_STANDARD) 
    set(CMAKE_CXX_STANDARD 17)
endif()

# Disable building as shared lib for now, could enable this later 
# if necessary.
if(BUILD_SHARED_LIBS)
    message(FATAL_ERROR "Building shared libs not supported")
endif()

# Check if CTest should be enabled.
if(BUILD_TESTING)
    enable_testing()

    # Be very verbose on test failure.
    list(APPEND CMAKE_CTEST_ARGUMENTS "--output-on-failure")
endif()
