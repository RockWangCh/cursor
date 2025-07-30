cmake_minimum_required(VERSION 3.14)
project(CRPRouter)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 编译选项
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -O3 -march=native -mtune=native")
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -Wextra -Wpedantic")

# OpenMP支持
find_package(OpenMP)
if(OpenMP_CXX_FOUND)
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} ${OpenMP_CXX_FLAGS}")
endif()

# 线程库
find_package(Threads REQUIRED)

# 第三方库
include(FetchContent)

# cpp-httplib
FetchContent_Declare(
    httplib
    GIT_REPOSITORY https://github.com/yhirose/cpp-httplib.git
    GIT_TAG v0.14.1
)
FetchContent_MakeAvailable(httplib)

# nlohmann/json
FetchContent_Declare(
    json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.11.2
)
FetchContent_MakeAvailable(json)

# 源文件
set(SOURCES
    main.cpp
    crp_router.cpp
    road_network.cpp
    navinfo_adapter.cpp
    crp_preprocessor.cpp
)

# 可执行文件
add_executable(crp_router_service ${SOURCES})

# 链接库
target_link_libraries(crp_router_service
    PRIVATE
    httplib::httplib
    nlohmann_json::nlohmann_json
    Threads::Threads
)

if(OpenMP_CXX_FOUND)
    target_link_libraries(crp_router_service PRIVATE OpenMP::OpenMP_CXX)
endif()

# 包含目录
target_include_directories(crp_router_service
    PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}
)

# 测试
enable_testing()
add_subdirectory(tests)

# 安装
install(TARGETS crp_router_service
    RUNTIME DESTINATION bin
)

# Docker构建脚本
configure_file(
    ${CMAKE_CURRENT_SOURCE_DIR}/Dockerfile.in
    ${CMAKE_CURRENT_BINARY_DIR}/Dockerfile
    @ONLY
)