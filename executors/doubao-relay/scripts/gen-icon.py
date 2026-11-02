#!/usr/bin/env python3
"""生成 1024x1024 纯色(#3B82F6)占位 PNG,只用标准库。

用法: python3 scripts/gen-icon.py [输出路径]
默认输出 scripts/app-icon.png,之后执行:
    cargo tauri icon scripts/app-icon.png -o src-tauri/icons
"""
import os
import struct
import sys
import zlib

SIZE = 1024
COLOR = (0x3B, 0x82, 0xF6, 0xFF)


def chunk(tag: bytes, data: bytes) -> bytes:
    body = tag + data
    return struct.pack(">I", len(data)) + body + struct.pack(">I", zlib.crc32(body) & 0xFFFFFFFF)


def build_png() -> bytes:
    row = b"\x00" + bytes(COLOR) * SIZE  # filter type 0 + RGBA 像素
    raw = row * SIZE
    ihdr = struct.pack(">IIBBBBB", SIZE, SIZE, 8, 6, 0, 0, 0)  # 8bit, RGBA
    return (
        b"\x89PNG\r\n\x1a\n"
        + chunk(b"IHDR", ihdr)
        + chunk(b"IDAT", zlib.compress(raw, 9))
        + chunk(b"IEND", b"")
    )


def main() -> None:
    default = os.path.join(os.path.dirname(os.path.abspath(__file__)), "app-icon.png")
    out = sys.argv[1] if len(sys.argv) > 1 else default
    with open(out, "wb") as f:
        f.write(build_png())
    print(out)


if __name__ == "__main__":
    main()
