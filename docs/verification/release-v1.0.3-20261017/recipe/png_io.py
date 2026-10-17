#!/usr/bin/env python3
"""純標準函式庫的 PNG 讀寫，供對拍與抽取工具共用。

只支援本專案實際使用的兩種格式：8-bit truecolor（type 2）與 8-bit truecolor+alpha
（type 6）。遇到其他格式一律失敗，不做靜默降級。
"""

from __future__ import annotations

import pathlib
import struct
import zlib


class Image:
    __slots__ = ("width", "height", "channels", "pixels")

    def __init__(self, width: int, height: int, channels: int, pixels: bytes):
        self.width = width
        self.height = height
        self.channels = channels
        self.pixels = pixels

    def rgb(self) -> bytes:
        """丟掉 alpha，取 RGB。"""
        if self.channels == 3:
            return self.pixels
        out = bytearray(self.width * self.height * 3)
        for i in range(self.width * self.height):
            out[i * 3:i * 3 + 3] = self.pixels[i * 4:i * 4 + 3]
        return bytes(out)

    def crop(self, x0: int, y0: int, w: int, h: int) -> "Image":
        if x0 < 0 or y0 < 0 or x0 + w > self.width or y0 + h > self.height:
            raise ValueError("裁切區超出影像範圍")
        c = self.channels
        rows = []
        for y in range(y0, y0 + h):
            base = (y * self.width + x0) * c
            rows.append(self.pixels[base:base + w * c])
        return Image(w, h, c, b"".join(rows))


def read_png(path: pathlib.Path | str) -> Image:
    data = pathlib.Path(path).read_bytes()
    if data[:8] != b"\x89PNG\r\n\x1a\n":
        raise ValueError(f"{path} 不是 PNG")
    pos, idat, width, height, channels = 8, b"", 0, 0, 0
    while pos < len(data):
        length = struct.unpack_from(">I", data, pos)[0]
        tag = data[pos + 4:pos + 8]
        body = data[pos + 8:pos + 8 + length]
        if tag == b"IHDR":
            width, height, depth, ctype = struct.unpack(">IIBB", body[:10])
            if depth != 8 or ctype not in (2, 6):
                raise ValueError(f"{path} 不是 8-bit truecolor(+alpha) PNG")
            channels = 3 if ctype == 2 else 4
        elif tag == b"IDAT":
            idat += body
        pos += 12 + length
    if not channels:
        raise ValueError(f"{path} 缺少 IHDR")
    raw = zlib.decompress(idat)
    stride = width * channels
    if len(raw) != height * (stride + 1):
        raise ValueError(f"{path} 解壓長度不符")
    prev = bytearray(stride)
    out = bytearray()
    i = 0
    for _ in range(height):
        ftype = raw[i]
        i += 1
        line = bytearray(raw[i:i + stride])
        i += stride
        if ftype == 1:
            for x in range(channels, stride):
                line[x] = (line[x] + line[x - channels]) & 0xFF
        elif ftype == 2:
            for x in range(stride):
                line[x] = (line[x] + prev[x]) & 0xFF
        elif ftype == 3:
            for x in range(stride):
                left = line[x - channels] if x >= channels else 0
                line[x] = (line[x] + ((left + prev[x]) >> 1)) & 0xFF
        elif ftype == 4:
            for x in range(stride):
                a = line[x - channels] if x >= channels else 0
                b = prev[x]
                c = prev[x - channels] if x >= channels else 0
                p = a + b - c
                pa, pb, pc = abs(p - a), abs(p - b), abs(p - c)
                pred = a if (pa <= pb and pa <= pc) else (b if pb <= pc else c)
                line[x] = (line[x] + pred) & 0xFF
        elif ftype != 0:
            raise ValueError(f"{path} 有未支援的 filter {ftype}")
        out += line
        prev = line
    return Image(width, height, channels, bytes(out))


def _chunk(tag: bytes, body: bytes) -> bytes:
    payload = tag + body
    return (struct.pack(">I", len(body)) + payload
            + struct.pack(">I", zlib.crc32(payload) & 0xFFFFFFFF))


def write_png_rgb(path: pathlib.Path | str, width: int, height: int, rgb: bytes) -> str:
    raw = b"".join(b"\x00" + rgb[y * width * 3:(y + 1) * width * 3] for y in range(height))
    png = (b"\x89PNG\r\n\x1a\n"
           + _chunk(b"IHDR", struct.pack(">IIBBBBB", width, height, 8, 2, 0, 0, 0))
           + _chunk(b"IDAT", zlib.compress(raw, 9))
           + _chunk(b"IEND", b""))
    pathlib.Path(path).write_bytes(png)
    import hashlib
    return hashlib.sha256(png).hexdigest()
