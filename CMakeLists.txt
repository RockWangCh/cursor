cmake_minimum_required(VERSION 3.18)
project(crp_router LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

add_library(crp_lib INTERFACE)
target_include_directories(crp_lib INTERFACE include)

add_executable(crp_main src/main.cpp)
target_link_libraries(crp_main PRIVATE crp_lib)