"""將 COMBPICS 前五張之一輸出為本機測試 PNG；不得作 production extractor。"""

from __future__ import annotations

import argparse
import hashlib
import struct
import zlib
import zipfile
from pathlib import Path

from inventory_image_archives import EXPECTED_ZIP_SHA256
from probe_image_lzw import decode


WIDTH = 296
HEIGHT = 144
PALETTE = (
    (0x00, 0x00, 0x00),
    (0x00, 0x00, 0xAA),
    (0x00, 0xAA, 0x00),
    (0x00, 0xAA, 0xAA),
    (0xAA, 0x00, 0x00),
    (0xAA, 0x00, 0xAA),
    (0xAA, 0x55, 0x00),
    (0xAA, 0xAA, 0xAA),
    (0x55, 0x55, 0x55),
    (0x55, 0x55, 0xFF),
    (0x55, 0xFF, 0x55),
    (0x55, 0xFF, 0xFF),
    (0xFF, 0x55, 0x55),
    (0xFF, 0x55, 0xFF),
    (0xFF, 0xFF, 0x55),
    (0xFF, 0xFF, 0xFF),
)


def decode_nibbles(data: bytes) -> bytes:
    if len(data) < 3:
        raise ValueError("二階資料過短")
    data_offset = int.from_bytes(data[:2], "big")
    if not 2 <= data_offset <= len(data):
        raise ValueError(f"二階 data offset 越界：{data_offset}")
    controls = data[2:data_offset]
    packed = data[data_offset:]
    nibble_position = 0
    output = bytearray()

    def next_nibble() -> int:
        nonlocal nibble_position
        if nibble_position >= len(packed) * 2:
            raise ValueError("二階 nibble stream 越界")
        value = packed[nibble_position // 2]
        result = value >> 4 if nibble_position % 2 == 0 else value & 0x0F
        nibble_position += 1
        return result

    for control in controls:
        if control < 0x80:
            output.extend([next_nibble()] * (control + 1))
        else:
            output.extend(next_nibble() for _ in range(control - 0x7F))
    return bytes(output)


def chunk(kind: bytes, payload: bytes) -> bytes:
    return (
        struct.pack(">I", len(payload))
        + kind
        + payload
        + struct.pack(">I", zlib.crc32(kind + payload) & 0xFFFFFFFF)
    )


def encode_rgb_png(indices: bytes, width: int = WIDTH, height: int = HEIGHT) -> bytes:
    rows = []
    for y in range(height):
        row = indices[y * width : (y + 1) * width]
        rows.append(b"\0" + bytes(channel for index in row for channel in PALETTE[index]))
    header = struct.pack(">IIBBBBB", width, height, 8, 2, 0, 0, 0)
    return b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", header) + chunk(
        b"IDAT", zlib.compress(b"".join(rows), 9)
    ) + chunk(b"IEND", b"")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("zip_path", type=Path)
    parser.add_argument("index", type=int, choices=range(5))
    parser.add_argument("output_path", type=Path)
    args = parser.parse_args()
    if "workplace" not in args.output_path.resolve().parts:
        raise SystemExit("測試 PNG 只可寫入 ignored workplace")

    zip_bytes = args.zip_path.read_bytes()
    digest = hashlib.sha256(zip_bytes).hexdigest()
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit(f"非預期 ZIP SHA-256：{digest}")
    with zipfile.ZipFile(args.zip_path) as archive:
        source = archive.read("Knights-of-legend/COMBPICS.0")
    count = source[0]
    offsets = struct.unpack_from(f"<{count}H", source, 20)
    lengths = [right - left for left, right in zip(offsets, offsets[1:])]
    lengths.append(struct.unpack_from("<H", source, 10)[0])
    offset = offsets[args.index]
    first_stage = decode(source[offset : offset + lengths[args.index]])
    indices = decode_nibbles(first_stage)
    if len(indices) != WIDTH * HEIGHT:
        raise SystemExit(f"非預期像素數：{len(indices)}")
    if not 0 <= min(indices) <= max(indices) <= 15:
        raise SystemExit("色彩索引超出 4-bit 範圍")
    args.output_path.parent.mkdir(parents=True, exist_ok=True)
    args.output_path.write_bytes(encode_rgb_png(indices, WIDTH, HEIGHT))
    print(
        f"index={args.index} size={WIDTH}x{HEIGHT} colors={len(set(indices))} "
        f"png_sha256={hashlib.sha256(args.output_path.read_bytes()).hexdigest()}"
    )


if __name__ == "__main__":
    main()
