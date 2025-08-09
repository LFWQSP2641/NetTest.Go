#!/bin/bash

cd "Go"
MODE="release"
if [ "$1" = "debug" ]; then
    MODE="debug"
fi
case "$(uname -s)" in
    Linux*)     HOST_TAG=linux-x86_64;;
    Darwin*)    HOST_TAG=darwin-x86_64;;
    *MINGW*|*CYGWIN*|*MSYS*) HOST_TAG=windows-x86_64;;
    *)          echo "Unsupported system"; exit 1;;
esac
if [ -n "${ANDROID_NDK_HOME}" ]; then
    NDK_PATH="${ANDROID_NDK_HOME}"
elif [ -n "${NDK_ROOT}" ]; then
    NDK_PATH="${NDK_ROOT}"
elif [ -n "${NDK_HOME}" ]; then
    NDK_PATH="${NDK_HOME}"
else
    echo "Error: NDK path not found. Please set ANDROID_NDK_HOME, NDK_ROOT or NDK_HOME"
    exit 1
fi
export GOOS=android
export CGO_ENABLED=1

export GOARCH=arm64
export CC="${NDK_PATH}/toolchains/llvm/prebuilt/${HOST_TAG}/bin/aarch64-linux-android35-clang"
if [ "$MODE" = "release" ]; then
    go build -buildmode=c-shared -ldflags "-s -w" -o ./build/libandroidnetcore_arm64-v8a.so .
else
    go build -buildmode=c-shared -o ./build/libandroidnetcore_arm64-v8a.so .
fi

export GOARCH=arm
export CC="${NDK_PATH}/toolchains/llvm/prebuilt/${HOST_TAG}/bin/armv7a-linux-androideabi35-clang"
if [ "$MODE" = "release" ]; then
    go build -buildmode=c-shared -ldflags "-s -w" -o ./build/libandroidnetcore_armeabi-v7a.so .
else
    go build -buildmode=c-shared -o ./build/libandroidnetcore_armeabi-v7a.so .
fi

export GOARCH=386
export CC="${NDK_PATH}/toolchains/llvm/prebuilt/${HOST_TAG}/bin/i686-linux-android35-clang"
if [ "$MODE" = "release" ]; then
    go build -buildmode=c-shared -ldflags "-s -w" -o ./build/libandroidnetcore_x86.so .
else
    go build -buildmode=c-shared -o ./build/libandroidnetcore_x86.so .
fi

export GOARCH=amd64
export CC="${NDK_PATH}/toolchains/llvm/prebuilt/${HOST_TAG}/bin/x86_64-linux-android35-clang"
if [ "$MODE" = "release" ]; then
    go build -buildmode=c-shared -ldflags "-s -w" -o ./build/libandroidnetcore_x86_64.so .
else
    go build -buildmode=c-shared -o ./build/libandroidnetcore_x86_64.so .
fi

# copy
cd ..
mkdir -p Qt/lib
cp Go/build/libandroidnetcore_arm64-v8a.so Qt/lib/
cp Go/build/libandroidnetcore_armeabi-v7a.so Qt/lib/
cp Go/build/libandroidnetcore_x86.so Qt/lib/
cp Go/build/libandroidnetcore_x86_64.so Qt/lib/
