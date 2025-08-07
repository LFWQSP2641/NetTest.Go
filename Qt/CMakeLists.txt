cmake_minimum_required(VERSION 3.16)

project(NetTest VERSION 0.1.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(Qt6 REQUIRED COMPONENTS Core)

qt_standard_project_setup(REQUIRES 6.5)

set(SOURCES
    main.cpp
)

qt_add_executable(${PROJECT_NAME} ${SOURCES})

include_directories(${CMAKE_SOURCE_DIR})

target_link_libraries(${PROJECT_NAME} PRIVATE Qt6::Core)

install(TARGETS ${PROJECT_NAME} DESTINATION bin)

if(NOT ${DEVELOPMENT_MODE})
    set_target_properties(${PROJECT_NAME} PROPERTIES
        WIN32_EXECUTABLE ON
        MACOSX_BUNDLE ON
    )
    add_compile_definitions(QT_NO_DEBUG_OUTPUT)
endif()
