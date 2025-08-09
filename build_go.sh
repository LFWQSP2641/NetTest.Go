#!/bin/bash

cd "Go"
MODE="release"
if [ "$1" = "debug" ]; then
    MODE="debug"
fi
echo $(uname -s)
case "$(uname -s)" in
    *MINGW*|*CYGWIN*|*MSYS*)
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
mkdir -p Qt/lib
case "$(uname -s)" in
    *MINGW*|*CYGWIN*|*MSYS*)
        cp Go/build/netcore.dll Qt/lib/
        ;;
    *)
        cp Go/build/netcore.so Qt/lib/
        ;;
esac
