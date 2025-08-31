#!/usr/bin/env python3
"""
Cross-platform replacement for build_go.sh

Usage:
  python scripts/build_go.py [debug|release]

Behavior mirrors build_go.sh:
- Builds Go C-shared library into Go/build/netcore.(dll|so)
- Copies artifact into lib/ at repo root
"""
from __future__ import annotations

import os
import shutil
import subprocess
import sys
import platform
from pathlib import Path


def is_windows_like() -> bool:
    sysname = platform.system()
    if sysname in ("Windows",):
        return True
    up = sysname.upper()
    return any(s in up for s in ("MINGW", "CYGWIN", "MSYS"))


def run(cmd: list[str], cwd: Path, env: dict | None = None) -> None:
    print("+", " ".join(cmd))
    subprocess.run(cmd, cwd=str(cwd), env=env, check=True)


def main(argv: list[str]) -> int:
    # Determine mode
    mode = "release"
    if len(argv) >= 2 and argv[1].lower() == "debug":
        mode = "debug"

    root = Path(__file__).resolve().parent
    go_dir = root / "Go"
    build_dir = go_dir / "build"
    lib_dir = root / "lib"

    build_dir.mkdir(parents=True, exist_ok=True)
    lib_dir.mkdir(parents=True, exist_ok=True)

    # Choose output name based on platform
    if is_windows_like():
        out_name = "netcore.dll"
    else:
        out_name = "netcore.so"

    out_path = build_dir / out_name

    # Build args
    cmd = [
        "go",
        "build",
        "-buildmode=c-shared",
    ]
    if mode == "release":
        cmd += ["-ldflags", "-s -w"]
    cmd += ["-o", str(out_path), "."]

    try:
        run(cmd, cwd=go_dir)
    except FileNotFoundError:
        print(
            "Error: 'go' command not found. Please install Go and ensure it is in PATH."
        )
        return 1
    except subprocess.CalledProcessError as e:
        print(f"go build failed with exit code {e.returncode}")
        return e.returncode or 1

    # Copy to lib/
    dest = lib_dir / out_name
    shutil.copy2(out_path, dest)
    print(f"Copied {out_path} -> {dest}")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
