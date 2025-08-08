cmake_minimum_required(VERSION 3.16)

project(NetTest VERSION 0.1 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_STANDARD 17)

find_package(Qt6 REQUIRED COMPONENTS Core Quick)

qt_standard_project_setup(REQUIRES 6.8)

qt_add_executable(appNetTest
    src/main.cpp
)

qt_add_qml_module(appNetTest
    URI NetTest
    VERSION 1.0
    QML_FILES
        page/Main.qml page/DNSQueryPage.qml
    SOURCES
        src/dns/DnsQuery.h src/dns/DnsQuery.cpp
        src/dns/DnsQueryTask.h src/dns/DnsQueryTask.cpp
)

include_directories(src)
include_directories(${CMAKE_SOURCE_DIR})

include_directories(src/dns)

set_target_properties(appNetTest PROPERTIES
    QT_ANDROID_EXTRA_LIBS
        ${CMAKE_SOURCE_DIR}/lib/libandroidnetcore.so
)

# Qt for iOS sets MACOSX_BUNDLE_GUI_IDENTIFIER automatically since Qt 6.1.
# If you are developing for iOS or macOS you should consider setting an
# explicit, fixed bundle identifier manually though.
set_target_properties(appNetTest PROPERTIES
#    MACOSX_BUNDLE_GUI_IDENTIFIER com.example.appNetTest
    MACOSX_BUNDLE_BUNDLE_VERSION ${PROJECT_VERSION}
    MACOSX_BUNDLE_SHORT_VERSION_STRING ${PROJECT_VERSION_MAJOR}.${PROJECT_VERSION_MINOR}
    MACOSX_BUNDLE TRUE
    WIN32_EXECUTABLE TRUE
)

target_link_libraries(appNetTest
    PRIVATE Qt6::Quick
)
target_link_libraries(appNetTest PRIVATE Qt6::Core)
target_link_libraries(appNetTest PRIVATE Qt6::Core)

if(NOT ANDROID)
    add_custom_command(TARGET appNetTest POST_BUILD
        COMMAND ${CMAKE_COMMAND} -E copy
            ${CMAKE_SOURCE_DIR}/lib/netcore.$<$<PLATFORM_ID:Windows>:dll>$<$<PLATFORM_ID:Linux>:so>$<$<PLATFORM_ID:Darwin>:dylib>
            $<TARGET_FILE_DIR:appNetTest>
        COMMENT "Copying Go dynamic library..."
    )
endif()

include(GNUInstallDirs)
install(TARGETS appNetTest
    BUNDLE DESTINATION .
    LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
)
