# ============================================================================
# Public-facing convenience functions & macros for building.
# ============================================================================

# To gain access to standard install directory variables such as CMAKE_INSTALL_LIBDIR.
include(GNUInstallDirs)

#
# Public entry point for building a C++ based shared library.
#
function(cpp_library TARGET_NAME)
    set(options)

    set(oneValueArgs
    )

    set(multiValueArgs
        PUBLIC_HEADERS_INSTALL_PREFIX
        PUBLIC_HEADERS
        CPPFILES
        LIBRARIES
        INCLUDE_DIRS
        DEFINES
    )

    cmake_parse_arguments(args
        "${options}"
        "${oneValueArgs}"
        "${multiValueArgs}"
        ${ARGN}
    )

    #
    # Resolve build variables.
    #

    # Determine public header install location.
    if (args_PUBLIC_HEADERS_INSTALL_PREFIX)
        set(PUBLIC_HEADERS_INSTALL_PREFIX ${args_PUBLIC_HEADERS_INSTALL_PREFIX})
    else()
        set(PUBLIC_HEADERS_INSTALL_PREFIX ${TARGET_NAME})
    endif()

    # Copy public headers into build tree.
    file(
        COPY ${args_PUBLIC_HEADERS}
        DESTINATION ${CMAKE_BINARY_DIR}/${CMAKE_INSTALL_INCLUDEDIR}/${PUBLIC_HEADERS_INSTALL_PREFIX}
    )

    # Install public headers.
    install(
        FILES ${args_PUBLIC_HEADERS}
        DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}/${PUBLIC_HEADERS_INSTALL_PREFIX}
    )

    # Add a new library target.
    add_library(${TARGET_NAME})

    # Add sources for building the target.
    target_sources(${TARGET_NAME}
        PRIVATE
            ${args_CPPFILES}
            ${args_PUBLIC_HEADERS}
    )

    # Apply common compiler properties, and include path properties.
    _cpp_target_properties(${TARGET_NAME}
        INCLUDE_DIRS
            ${args_INCLUDE_DIRS}
        DEFINES 
            ${args_DEFINES}
        LIBRARIES 
            ${args_LIBRARIES}
    )

    # Install the library and export it as a public target.
    install(
        TARGETS ${TARGET_NAME}
        EXPORT ${CMAKE_PROJECT_NAME}-targets
        LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
        RUNTIME DESTINATION ${CMAKE_INSTALL_LIBDIR}
    )

    # Update the target properties.
    set_target_properties(${TARGET_NAME}
        PROPERTIES
            PREFIX "${CMAKE_SHARED_LIBRARY_PREFIX}"
    )

endfunction() # cpp_library


#
# Adds a cpp test that is executed by CTest.
#
function(cpp_test TEST_TARGET)

    if (NOT BUILD_TESTING)
        return()
    endif()

    set(options)

    set(oneValueArgs
    )

    set(multiValueArgs
        CPPFILES
        LIBRARIES
        INCLUDE_DIRS
    )

    cmake_parse_arguments(args
        "${options}"
        "${oneValueArgs}"
        "${multiValueArgs}"
        ${ARGN}
    )

    # Define a new executable.
    add_executable(${TEST_TARGET}
        ${args_CPPFILES}
    )

    # Apply properties.
    _cpp_target_properties(${TEST_TARGET}
        INCLUDE_DIRS
            ${args_INCLUDE_DIRS}
        LIBRARIES
            ${args_LIBRARIES}
    )

    # Add the test target.
    add_test(
        NAME ${TEST_TARGET}
        COMMAND $<TARGET_FILE:${TEST_TARGET}>
    )
endfunction() # cpp_test


function(_cpp_target_properties TARGET_NAME)
    set(options)
    set(oneValueArgs)
    set(multiValueArgs
        INCLUDE_DIRS
        DEFINES
        LIBRARIES
    )

    cmake_parse_arguments(
        args
        "${options}"
        "${oneValueArgs}"
        "${multiValueArgs}"
        ${ARGN}
    )

    target_compile_definitions(${TARGET_NAME}
        PRIVATE
            ${args_DEFINES}
    )

    target_compile_options(${TARGET_NAME}
        PRIVATE
            -Wno-deprecated -Wall
    )  

    # Exported include paths for this target.
    target_include_directories(${TARGET_NAME}
        INTERFACE
            $<INSTALL_INTERFACE:${CMAKE_INSTALL_INCLUDEDIR}>
    )

    # Project includes for building against.
    target_include_directories(${TARGET_NAME}
        PRIVATE
            $<BUILD_INTERFACE:${CMAKE_BINARY_DIR}/${CMAKE_INSTALL_INCLUDEDIR}>
    )

    # Setup include path for binary dir.
    # We set external includes as SYSTEM so that their warnings are muted.
    target_include_directories(${TARGET_NAME}
        SYSTEM
        PRIVATE
            ${args_INCLUDE_DIRS}
    )

    # Link to libraries.
    target_link_libraries(${TARGET_NAME}
        PRIVATE
            ${args_LIBRARIES}
    )
endfunction() # _cpp_target_properties
