# ===============
# nlohmann::json
# ===============

function(FetchNlohmannJson)
  set(options)

  set(oneValueArgs
    VERSION
  )

  set(multiValueArgs
  )

  cmake_parse_arguments(args
      "${options}"
      "${oneValueArgs}"
      "${multiValueArgs}"
      ${ARGN}
  )

  FetchContent_Declare(nlohmann_json
    URL https://github.com/nlohmann/json/releases/download/v${args_VERSION}/json.tar.xz
  )

  FetchContent_GetProperties(nlohmann_json)
  if(NOT nlohmann_json_POPULATED)
    FetchContent_Populate(nlohmann_json)
  
    set(nlohmann_json_CACHE_ARGS
      "-DCMAKE_INSTALL_PREFIX:STRING=${nlohmann_json_BINARY_DIR}/install" 
      "-DCMAKE_POSITION_INDEPENDENT_CODE:BOOL=ON"
      "-DCMAKE_VERBOSE_MAKEFILE:BOOL=ON"         
      "-DJSON_BuildTests:BOOL=OFF"
    )
    if(CMAKE_TOOLCHAIN_FILE)
      list(APPEND nlohmann_json_CACHE_ARGS 
        "-DCMAKE_TOOLCHAIN_FILE:FILEPATH=${CMAKE_TOOLCHAIN_FILE}"
      )
    else()
      list(APPEND nlohmann_json_CACHE_ARGS 
        "-DCMAKE_CXX_COMPILER:FILEPATH=${CMAKE_CXX_COMPILER}"
        "-DCMAKE_C_COMPILER:FILEPATH=${CMAKE_C_COMPILER}"
      )
    endif()    
  
    get_property(isMulti GLOBAL PROPERTY GENERATOR_IS_MULTI_CONFIG)
    if(NOT isMulti)
      list(APPEND nlohmann_json_CACHE_ARGS "-DCMAKE_BUILD_TYPE:STRING=Release")
    endif()
  
    if(CMAKE_GENERATOR_PLATFORM)
      list(APPEND nlohmann_json_GENERATOR_ARGS
        --build-generator-platform "${CMAKE_GENERATOR_PLATFORM}"
      )
    endif()
    if(CMAKE_GENERATOR_TOOLSET)
      list(APPEND nlohmann_json_GENERATOR_ARGS
        --build-generator-toolset "${CMAKE_GENERATOR_TOOLSET}"
      )
    endif()
  
    message(STATUS "Configuring and building nlohmann::json-${args_VERSION} immediately")
    execute_process(
      COMMAND ${CMAKE_CTEST_COMMAND}
              --build-and-test  ${nlohmann_json_SOURCE_DIR} ${nlohmann_json_BINARY_DIR}
              --build-generator "Unix Makefiles" ${nlohmann_json_GENERATOR_ARGS}
              --build-target    install
              --build-noclean
              --build-options   ${nlohmann_json_CACHE_ARGS}
      WORKING_DIRECTORY ${nlohmann_json_SOURCE_DIR}
      OUTPUT_FILE       ${nlohmann_json_BINARY_DIR}/build_output.log
      ERROR_FILE        ${nlohmann_json_BINARY_DIR}/build_output.log
      RESULT_VARIABLE   result
    )
    if(result)
      file(READ ${nlohmann_json_BINARY_DIR}/build_output.log build_log)
      message(FATAL_ERROR "Result = ${result}\nFailed nlohmann::json-${args_VERSION} build, see build log:\n"
        "${build_log}")
    endif()
    message(STATUS "nlohmann::json-${args_VERSION} build complete")
  endif()
  
  # Confirm that we can find nlohmann::json.
  find_package(nlohmann_json 
    QUIET 
    REQUIRED 
    CONFIG
    PATHS "${nlohmann_json_BINARY_DIR}/install"
    NO_DEFAULT_PATH 
  )
  if (NOT nlohmann_json_FOUND) 
    message(FATAL_ERROR "nlohmann::json-${args_VERSION} not found")
  endif()
  check_target(nlohmann_json::nlohmann_json)
endfunction()