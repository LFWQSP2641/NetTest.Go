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
mkdir -p lib
case "$(uname -s)" in
    *MINGW*|*CYGWIN*|*MSYS*)
        cp Go/build/netcore.dll lib/
        ;;
    *)
        cp Go/build/netcore.so lib/
        ;;
esac
