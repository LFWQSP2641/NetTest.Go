#!/bin/bash

cd "go"
case "$(uname -s)" in
    MINGW*|CYGWIN*|MSYS*)
        go build -buildmode=c-shared -o ./build/netcore.dll .
        ;;
    *)
        go build -buildmode=c-shared -o ./build/netcore.so .
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
