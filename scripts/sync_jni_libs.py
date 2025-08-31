#!/usr/bin/env python3
"""
Copy Go-built Android .so files from repo lib/ into Flutter Android jniLibs structure.

Usage:
  python scripts/sync_jni_libs.py

It will map:
  lib/libandroidnetcore_arm64-v8a.so -> flutter/android/app/src/main/jniLibs/arm64-v8a/
  lib/libandroidnetcore_armeabi-v7a.so -> flutter/android/app/src/main/jniLibs/armeabi-v7a/
  lib/libandroidnetcore_x86.so -> flutter/android/app/src/main/jniLibs/x86/
  lib/libandroidnetcore_x86_64.so -> flutter/android/app/src/main/jniLibs/x86_64/

Also writes an alias libnetcore.so per ABI for simpler System.loadLibrary("netcore").
"""
from __future__ import annotations

import shutil
from pathlib import Path

# Resolve repository root from current working directory so script can be invoked from repo root
# as `python scripts/sync_jni_libs.py` irrespective of script file location.
ROOT = Path.cwd()
LIB = ROOT / "lib"
APP = ROOT / "flutter" / "android" / "app"

MAPPING = {
    "arm64-v8a": "libandroidnetcore_arm64-v8a.so",
    "armeabi-v7a": "libandroidnetcore_armeabi-v7a.so",
    "x86": "libandroidnetcore_x86.so",
    "x86_64": "libandroidnetcore_x86_64.so",
}


def sync() -> int:
    if not LIB.exists():
        print(f"lib directory not found: {LIB}")
        return 1
    for abi, name in MAPPING.items():
        src = LIB / name
        if not src.exists():
            print(f"skip {abi}: {name} not found")
            continue
        dst_dir = APP / "src" / "main" / "jniLibs" / abi
        dst_dir.mkdir(parents=True, exist_ok=True)
        # Only write alias libnetcore.so for this ABI (do not copy libandroidnetcore_*.so)
        alias = dst_dir / "libnetcore.so"
        print(f"alias {src} -> {alias}")
        shutil.copy2(src, alias)
    return 0


if __name__ == "__main__":
    raise SystemExit(sync())
