cmake_minimum_required(VERSION 3.16)

project(NetTest VERSION 0.6.4 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_STANDARD 17)

find_package(Qt6 REQUIRED COMPONENTS Core Quick QuickControls2 LinguistTools)

qt_standard_project_setup(
    REQUIRES 6.5
    I18N_TRANSLATED_LANGUAGES zh_CN
)

qt_add_translations(appNetTest
    RESOURCE_PREFIX "/"
    TS_FILE_DIR translations
)

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
    RESOURCES
        android/AndroidManifest.xml android/build.gradle
        android/res/values/libs.xml android/res/xml/qtprovider_paths.xml
        android/gradle/wrapper/gradle-wrapper.jar
        android/gradle/wrapper/gradle-wrapper.properties
        android/gradle.properties android/gradlew android/gradlew.bat
)

include_directories(src)
include_directories(${CMAKE_SOURCE_DIR})

include_directories(src/dns)

# Qt for iOS sets MACOSX_BUNDLE_GUI_IDENTIFIER automatically since Qt 6.1.
# If you are developing for iOS or macOS you should consider setting an
# explicit, fixed bundle identifier manually though.
set_target_properties(appNetTest PROPERTIES
#    MACOSX_BUNDLE_GUI_IDENTIFIER com.example.appNetTest
    MACOSX_BUNDLE_BUNDLE_VERSION ${PROJECT_VERSION}
    MACOSX_BUNDLE_SHORT_VERSION_STRING ${PROJECT_VERSION_MAJOR}.${PROJECT_VERSION_MINOR}
    MACOSX_BUNDLE TRUE
    WIN32_EXECUTABLE TRUE
    QT_ANDROID_VERSION_NAME ${CMAKE_PROJECT_VERSION}
    QT_ANDROID_VERSION_CODE 1
    QT_ANDROID_PACKAGE_NAME "com.LFWQSP2641.appNetTestGo"
    QT_ANDROID_EXTRA_LIBS ${CMAKE_SOURCE_DIR}/../lib/libandroidnetcore_${CMAKE_ANDROID_ARCH_ABI}.so
)

target_link_libraries(appNetTest
    PRIVATE Qt6::Core Qt6::Quick Qt6::QuickControls2
)

# Copy Go dynamic library for non-Android builds
if(NOT CMAKE_SYSTEM_NAME STREQUAL "Android")
    add_custom_command(TARGET appNetTest POST_BUILD
        COMMAND ${CMAKE_COMMAND} -E copy
            ${CMAKE_SOURCE_DIR}/../lib/netcore.$<$<PLATFORM_ID:Windows>:dll>$<$<PLATFORM_ID:Linux>:so>$<$<PLATFORM_ID:Darwin>:dylib>
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
