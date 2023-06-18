cmake_minimum_required(VERSION 3.10)

cmake_minimum_required(VERSION 3.14.7)

# Set default build type to Release if empty.
set(CMAKE_BUILD_TYPE "Release" CACHE STRING
	"Choose the type of build, options are: Debug Release RelWithDebInfo MinSizeRel.")

project(DanceSolver VERSION 0.0 LANGUAGES CXX C)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_C_STANDARD 23)
set(CMAKE_CXX_FLAGS_RELEASE "-O3")
set(CMAKE_C_FLAGS_RELEASE "-O3")

#set(CMAKE_VERBOSE_MAKEFILE ON)

list(APPEND CMAKE_MODULE_PATH "${CMAKE_CURRENT_SOURCE_DIR}/cmake")

if(CMAKE_BUILD_TYPE STREQUAL "Release")
	add_definitions(-DNDEBUG)
endif()

# pkg-config
find_package(PkgConfig REQUIRED)

pkg_check_modules(re2 REQUIRED IMPORTED_TARGET re2)
find_package(ortools CONFIG REQUIRED)

get_target_property(ORTOOLS_PATH ortools::ortools LOCATION)
message(STATUS "ortools library is at ${ORTOOLS_PATH}")

# Hide all symbols by default.
set(CMAKE_C_VISIBILITY_PRESET hidden)
set(CMAKE_CXX_VISIBILITY_PRESET hidden)
set(CMAKE_VISIBILITY_INLINES_HIDDEN 1)

include_directories(src/include)

# set common compiler flags
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -Wextra -pedantic -Werror -g")

# include headers from the build dir - this is where the generated header for
# the go package is
include_directories(${CMAKE_CURRENT_BINARY_DIR})

# build static library
add_library(dance_solver_static STATIC src/logger.cpp src/dance_solver.cpp src/structs.cpp)
set_target_properties(dance_solver_static PROPERTIES OUTPUT_NAME dance_solver)
target_link_libraries(dance_solver_static PRIVATE ortools::ortools)

# add shared library
add_library(dance_solver SHARED src/logger.cpp src/dance_solver.cpp src/structs.cpp)
set_target_properties(dance_solver PROPERTIES OUTPUT_NAME dance_solver)
target_link_libraries(dance_solver PRIVATE ortools::ortools)

enable_testing()

add_subdirectory(tests)
