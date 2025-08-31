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
import re


def host_tag() -> str:
    sysname = platform.system()
    if sysname == "Linux":
        return "linux-x86_64"
    if sysname == "Darwin":
        return "darwin-x86_64"
    # Windows and MSYS/CYGWIN
    return "windows-x86_64"


def find_ndk() -> Path:
    # Prefer common env vars exported by CI or local setups.
    for key in ("ANDROID_NDK_HOME", "ANDROID_NDK_ROOT", "NDK_ROOT", "NDK_HOME"):
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

    # Base on current working directory so it works when invoked as `python scripts/build_android_go.py`
    root = Path.cwd()
    go_dir = root / "Go"
    build_dir = go_dir / "build"
    lib_dir = root / "lib"
    build_dir.mkdir(parents=True, exist_ok=True)
    lib_dir.mkdir(parents=True, exist_ok=True)

    # Target Android API level (used for target triple or cflags). Override via ANDROID_API_LEVEL
    api = os.environ.get("ANDROID_API_LEVEL", "28")

    # Define toolchain triples and minimal wrapper API levels present in NDK r23+.
    # NDK provides wrapper binaries only for certain minimal API levels
    targets = [
        {
            "goarch": "arm64",
            "triple": "aarch64-linux-android",
            "min_api": 28,
            "out": "libandroidnetcore_arm64-v8a.so",
        },
        {
            "goarch": "arm",
            "triple": "armv7a-linux-androideabi",
            "min_api": 28,
            "out": "libandroidnetcore_armeabi-v7a.so",
        },
        {
            "goarch": "386",
            "triple": "i686-linux-android",
            "min_api": 28,
            "out": "libandroidnetcore_x86.so",
        },
        {
            "goarch": "amd64",
            "triple": "x86_64-linux-android",
            "min_api": 28,
            "out": "libandroidnetcore_x86_64.so",
        },
    ]

    tool_base = ndk / "toolchains" / "llvm" / "prebuilt" / htag / "bin"

    def find_best_wrapper(triple: str, desired_api: int, min_api: int) -> Path | None:
        """Pick the best available *-clang wrapper for a triple.
        Prefer the highest API <= desired_api; otherwise pick the closest >= min_api.
        Handle Windows extensions (.exe/.cmd) as well as no-extension on Unix.
        """
        avail: dict[int, Path] = {}
        # Match both no-extension and Windows extensions
        patterns = [f"{triple}*-clang", f"{triple}*-clang.exe", f"{triple}*-clang.cmd"]
        for pattern in patterns:
            for p in tool_base.glob(pattern):
                # Strip Windows extensions for matching
                name_no_ext = p.stem if p.suffix in (".exe", ".cmd") else p.name
                m = re.match(rf"^{re.escape(triple)}(\d+)-clang$", name_no_ext)
                if m:
                    avail[int(m.group(1))] = p
        if not avail:
            return None
        # Prefer the highest API <= desired_api
        le = [a for a in avail if a <= desired_api]
        if le:
            return avail[max(le)]
        # Otherwise pick the smallest >= min_api
        ge = [a for a in avail if a >= min_api]
        if ge:
            return avail[min(ge)]
        # Fallback: pick the smallest available
        return avail[min(avail.keys())]

    for t in targets:
        goarch = t["goarch"]
        triple = t["triple"]
        min_api = t["min_api"]
        out_name = t["out"]
        env = os.environ.copy()
        env["GOOS"] = "android"
        env["CGO_ENABLED"] = "1"
        env["GOARCH"] = goarch
        if goarch == "arm":
            # Ensure we target ARMv7 for armeabi-v7a
            env["GOARM"] = "7"

        # Prefer an available wrapper closest to desired API, otherwise try unsuffixed wrapper, else generic clang
        desired_api_int = int(api)
        wrapper = find_best_wrapper(triple, desired_api_int, min_api)
        if wrapper is not None:
            env["CC"] = str(wrapper)
            print(f"Using CC={env['CC']} (API from wrapper)")
        else:
            # Try unsuffixed wrappers with possible Windows extensions
            candidates = [
                tool_base / f"{triple}-clang",
                tool_base / f"{triple}-clang.exe",
                tool_base / f"{triple}-clang.cmd",
            ]
            found = next((c for c in candidates if c.exists()), None)
            if found is not None:
                env["CC"] = str(found)
                print(f"Using CC={env['CC']} (unsuffixed wrapper)")
            else:
                # Fallback to generic clang and inject target/sysroot
                generic = tool_base / (
                    "clang.exe" if platform.system() == "Windows" else "clang"
                )
                env["CC"] = str(generic)
                sysroot = ndk / "toolchains" / "llvm" / "prebuilt" / htag / "sysroot"
                cflags = f"--target={triple}{api} --sysroot={sysroot}"
                # 为 32/64 位添加 sysroot 库检索路径
                lib_triple_map = {
                    "arm": "arm-linux-androideabi",
                    "arm64": "aarch64-linux-android",
                    "386": "i686-linux-android",
                    "amd64": "x86_64-linux-android",
                }
                lib_triple = lib_triple_map[goarch]
                lib_api_dir = sysroot / "usr" / "lib" / lib_triple / str(api)
                usr_lib_dir = sysroot / "usr" / "lib"
                ldflags_extra = f"-L{lib_api_dir} -L{usr_lib_dir}"
                env["CGO_CFLAGS"] = f"{cflags} {env.get('CGO_CFLAGS','')}".strip()
                env["CGO_LDFLAGS"] = (
                    f"{cflags} {ldflags_extra} {env.get('CGO_LDFLAGS','')}".strip()
                )
                print(
                    f"Using CC={env['CC']} (generic) with CGO_CFLAGS='{env['CGO_CFLAGS']}' CGO_LDFLAGS='{env['CGO_LDFLAGS']}'"
                )

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
