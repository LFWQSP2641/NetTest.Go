#!/usr/bin/env python3
"""
Cross-platform replacement for build_android_go.sh

Usage:
  python scripts/build_android_go.py [debug|release]

Builds Android c-shared Go libraries for ABIs:
  - arm64-v8a (GOARCH=arm64)
  - armeabi-v7a (GOARCH=arm)
  - x86 (GOARCH=386)
  - x86_64 (GOARCH=amd64)

Requires ANDROID NDK. Will use first defined of: ANDROID_NDK_HOME, NDK_ROOT, NDK_HOME
"""
from __future__ import annotations

import os
import shutil
import subprocess
import sys
import platform
from pathlib import Path


def host_tag() -> str:
    sysname = platform.system()
    if sysname == "Linux":
        return "linux-x86_64"
    if sysname == "Darwin":
        return "darwin-x86_64"
    # Windows and MSYS/CYGWIN
    return "windows-x86_64"


def find_ndk() -> Path:
    for key in ("ANDROID_NDK_HOME", "NDK_ROOT", "NDK_HOME"):
        val = os.environ.get(key)
        if val:
            p = Path(val)
            if p.exists():
                return p
    raise RuntimeError("NDK path not found. Set ANDROID_NDK_HOME, NDK_ROOT or NDK_HOME")


def run(cmd: list[str], cwd: Path, env: dict | None = None) -> None:
    print("+", " ".join(cmd))
    subprocess.run(cmd, cwd=str(cwd), env=env, check=True)


def main(argv: list[str]) -> int:
    mode = "release"
    if len(argv) >= 2 and argv[1].lower() == "debug":
        mode = "debug"

    ndk = find_ndk()
    htag = host_tag()

    root = Path(__file__).resolve().parent
    go_dir = root / "Go"
    build_dir = go_dir / "build"
    lib_dir = root / "lib"
    build_dir.mkdir(parents=True, exist_ok=True)
    lib_dir.mkdir(parents=True, exist_ok=True)

    # Android API level 35 toolchains (matching original script)
    api = "35"

    targets = [
        (
            "arm64",
            f"aarch64-linux-android{api}-clang",
            "libandroidnetcore_arm64-v8a.so",
        ),
        (
            "arm",
            f"armv7a-linux-androideabi{api}-clang",
            "libandroidnetcore_armeabi-v7a.so",
        ),
        ("386", f"i686-linux-android{api}-clang", "libandroidnetcore_x86.so"),
        ("amd64", f"x86_64-linux-android{api}-clang", "libandroidnetcore_x86_64.so"),
    ]

    tool_base = ndk / "toolchains" / "llvm" / "prebuilt" / htag / "bin"

    for goarch, cc_name, out_name in targets:
        env = os.environ.copy()
        env["GOOS"] = "android"
        env["CGO_ENABLED"] = "1"
        env["GOARCH"] = goarch
        env["CC"] = str(tool_base / cc_name)

        out_path = build_dir / out_name
        cmd = [
            "go",
            "build",
            "-buildmode=c-shared",
        ]
        if mode == "release":
            cmd += ["-ldflags", "-s -w"]
        cmd += ["-o", str(out_path), "."]

        try:
            run(cmd, cwd=go_dir, env=env)
        except FileNotFoundError:
            print(
                "Error: 'go' command not found. Please install Go and ensure it is in PATH."
            )
            return 1
        except subprocess.CalledProcessError as e:
            print(f"go build failed for {goarch} with exit code {e.returncode}")
            return e.returncode or 1

        # Copy to lib/
        shutil.copy2(out_path, lib_dir / out_name)
        print(f"Copied {out_path} -> {lib_dir / out_name}")

    return 0


if __name__ == "__main__":
    try:
        sys.exit(main(sys.argv))
    except RuntimeError as e:
        print(f"Error: {e}")
        sys.exit(1)
