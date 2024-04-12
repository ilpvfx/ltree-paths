# Discovery of the dependencies of this project.

function(check_target my_target)
    if(NOT TARGET ${my_target})
        message(FATAL_ERROR " ${my_target} CMake target required")
    endif()
endfunction() # check_target

if (BUILD_TESTING)
    include(FetchContent)

    include(FetchCatch2)
    FetchCatch2(VERSION 3.5.3)

    include(FetchNlohmannJson)
    FetchNlohmannJson(VERSION 3.11.3)
endif()