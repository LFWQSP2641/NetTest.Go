#!/bin/bash

cd "go"
MODE="release"
if [ "$1" = "debug" ]; then
    MODE="debug"
fi
case "$(uname -s)" in
    MINGW*|CYGWIN*|MSYS*)
        if [ "$MODE" = "release" ]; then
            go build -buildmode=c-shared -ldflags "-s -w" -o ./build/netcore.dll .
        else
            go build -buildmode=c-shared -o ./build/netcore.dll .
        fi
        ;;
    *)
        if [ "$MODE" = "release" ]; then
            go build -buildmode=c-shared -ldflags "-s -w" -o ./build/netcore.so .
        else
            go build -buildmode=c-shared -o ./build/netcore.so .
        fi
        ;;
esac

# copy
cd ..
case "$(uname -s)" in
    MINGW*|CYGWIN*|MSYS*)
        cp go/build/netcore.dll qt/lib/
        ;;
    *)
        cp go/build/netcore.so qt/lib/
        ;;
esac
