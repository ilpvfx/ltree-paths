# ===============
# Catch2
# ===============

function(FetchCatch2)
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
  
  FetchContent_Declare(catch2
    URL https://github.com/catchorg/Catch2/archive/refs/tags/v${args_VERSION}.tar.gz
  )
  FetchContent_GetProperties(catch2)
  if(NOT catch2_POPULATED)
    FetchContent_Populate(catch2)
  
      set(catch2_CACHE_ARGS
        "-DCMAKE_INSTALL_PREFIX:STRING=${catch2_BINARY_DIR}/install" 
        "-DCMAKE_POSITION_INDEPENDENT_CODE:BOOL=ON"
        "-DCMAKE_VERBOSE_MAKEFILE:BOOL=ON"
        "-DCATCH_ENABLE_WERROR=OFF"       
        "-DCATCH_BUILD_TESTING:BOOL=OFF"
        "-DCATCH_INSTALL_DOCS:BOOL=OFF"   
        "-DCATCH_INSTALL_HELPERS:BOOL=OFF"   
      )
      if(CMAKE_TOOLCHAIN_FILE)
        list(APPEND catch2_CACHE_ARGS 
          "-DCMAKE_TOOLCHAIN_FILE:FILEPATH=${CMAKE_TOOLCHAIN_FILE}"
        )
      else()
        list(APPEND catch2_CACHE_ARGS 
          "-DCMAKE_CXX_COMPILER:FILEPATH=${CMAKE_CXX_COMPILER}"
          "-DCMAKE_C_COMPILER:FILEPATH=${CMAKE_C_COMPILER}"
        )
      endif()    
  
      get_property(isMulti GLOBAL PROPERTY GENERATOR_IS_MULTI_CONFIG)
      if(NOT isMulti)
        list(APPEND catch2_CACHE_ARGS "-DCMAKE_BUILD_TYPE:STRING=Release")
      endif()
  
      if(CMAKE_GENERATOR_PLATFORM)
        list(APPEND catch2_GENERATOR_ARGS
          --build-generator-platform "${CMAKE_GENERATOR_PLATFORM}"
        )
      endif()
      if(CMAKE_GENERATOR_TOOLSET)
        list(APPEND catch2_GENERATOR_ARGS
          --build-generator-toolset "${CMAKE_GENERATOR_TOOLSET}"
        )
      endif()
  
      message(STATUS "Configuring and building Catch2-${args_VERSION} immediately")
      execute_process(
        COMMAND ${CMAKE_CTEST_COMMAND}
                --build-and-test  ${catch2_SOURCE_DIR} ${catch2_BINARY_DIR}
                --build-generator "Unix Makefiles" ${catch2_GENERATOR_ARGS}
                --build-target    install
                --build-noclean
                --build-options   ${catch2_CACHE_ARGS}
        WORKING_DIRECTORY ${catch2_SOURCE_DIR}
        OUTPUT_FILE       ${catch2_BINARY_DIR}/build_output.log
        ERROR_FILE        ${catch2_BINARY_DIR}/build_output.log
        RESULT_VARIABLE   result
      )
      if(result)
        file(READ ${fftw_BINARY_DIR}/build_output.log build_log)
        message(FATAL_ERROR "Result = ${result}\nFailed Catch2-${args_VERSION} build, see build log:\n"
          "${build_log}")
      endif()
      message(STATUS "Catch2-${args_VERSION} build complete")
  endif()
  
  # Confirm that we can find Catch2.
  find_package(Catch2 
    #QUIET 
    REQUIRED 
    CONFIG
    PATHS "${catch2_BINARY_DIR}/install"
    NO_DEFAULT_PATH 
  ) 
  if (NOT Catch2_FOUND) 
    message(FATAL_ERROR "Catch2 not found")
  endif()
  check_target(Catch2::Catch2)
endfunction()