find_package(Catch2 REQUIRED)
include(CTest)
include(Catch)

# build tests with asan
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fsanitize=address -fno-omit-frame-pointer")
set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -fsanitize=address -fno-omit-frame-pointer")

add_executable(tests testlogger.c dance_solver_tests.cpp)
target_link_libraries(tests PRIVATE Catch2::Catch2WithMain dance_solver_static)
catch_discover_tests(tests)

find_package(Check REQUIRED)

include_directories(${CHECK_INCLUDE_DIRS} ${CMAKE_PROJECT_INCLUDE})
add_executable(tests_c_api testlogger.c dance_solver_c_api_tests.c)
target_link_libraries(tests_c_api ${CHECK_LIBRARIES} dance_solver_static)

add_test(tests_c_api tests_c_api)
