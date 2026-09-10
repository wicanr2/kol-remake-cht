#!/usr/bin/env python3
"""戰場的磚塊 atlas（RE-079）。

`sub_1E3A8:0x1E4E5` 把 `KOLPICS` 的 payload 8（帶任務時是 12）讀進一個 32 KB 的
遠端緩衝區，戰場就是用那一批磚塊畫出來的。版面與城鎮磚塊相同（RE-059）：
一塊磚 16×16，四個平面各 32 byte，plane-major；payload 解出來 10,241 byte，
前 10,240 是 80 塊磚，最後一個 byte 是尾料。

輸出是 10 欄 × 8 列的 atlas PNG，與城鎮磚圖同樣的排法。

  python3 tools/extract_battle_tiles.py --out assets/original-derived-png/ega/BATTLE
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from inventory_image_archives import EXPECTED_ZIP_SHA256  # noqa: E402
from probe_extract_comb_png import encode_rgb_png  # noqa: E402
from probe_extract_town_map_png import encode_gray_png  # noqa: E402
from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-png/ega")
TILE_SIDE, TILE_BYTES, PLANE_BYTES = 16, 128, 32
ATLAS_COLUMNS = 10
PAYLOADS = {8: "battle-tiles-08.png", 12: "battle-tiles-12.png"}
# 任務地點的戰場：`TOWN0` 的 payload 0x3D + i，RLE 壓過，解出來是 64×64 的格子陣列。
# RLE 是 `sub_2B3D1:0x2B41A`：最高位元沒設就是一格，0xFF 結束，
# 其餘是「(值 & 0x7F) 重複下一個 byte 次」。
#
# 解完之後原版還做一次轉置（`sub_2B3D1:0x2B436` 的 `[bx+si]`，si 每次加 64），
# 才放進執行期的 4,096 byte 緩衝區——繪製時讀的是 `(Y << 6) + X`
# （`sub_223C0:0x22540`），兩次相反的擺法抵銷，所以**解出來的位元組就是
# `(X << 6) + Y`**：第一個索引是參戰者的 X，也就是畫面上的列。
# 輸出的 PNG 直接照這個順序擺，一列一個 X。
MAP_PAYLOAD_BASE, MAP_COUNT, MAP_SIDE = 0x3D, 25, 64
# 第 24 號不是任務：競技場 sub_2AAA8:0x2B0C9 → sub_2B3D1(0x18)，SPEC-025。


def rle_decode(data: bytes) -> bytes:
    out = bytearray()
    i = 0
    while i < len(data):
        value = data[i]
        i += 1
        if not value & 0x80:
            out.append(value)
            continue
        if value == 0xFF:
            break
        if i >= len(data):
            raise SystemExit("RLE 的重複次數缺了一個 byte")
        count = data[i]
        i += 1
        out.extend([value & 0x7F] * count)
    if len(out) != MAP_SIDE * MAP_SIDE:
        raise SystemExit(f"戰場格子陣列解出 {len(out)} byte，預期 {MAP_SIDE * MAP_SIDE}")
    return bytes(out)


def tile_pixels(seg: bytes, base: int) -> bytes:
    planes = [seg[base + p * PLANE_BYTES:base + (p + 1) * PLANE_BYTES] for p in range(4)]
    out = bytearray(TILE_SIDE * TILE_SIDE)
    for n in range(TILE_SIDE * TILE_SIDE):
        bit = 0x80 >> (n & 7)
        out[n] = sum((1 << p) if planes[p][n >> 3] & bit else 0 for p in range(4))
    return bytes(out)


def atlas(seg: bytes) -> tuple[bytes, int, int, int]:
    count = len(seg) // TILE_BYTES
    rows = (count + ATLAS_COLUMNS - 1) // ATLAS_COLUMNS
    width, height = ATLAS_COLUMNS * TILE_SIDE, rows * TILE_SIDE
    px = bytearray(width * height)
    for i in range(count):
        tile = tile_pixels(seg, i * TILE_BYTES)
        ox, oy = (i % ATLAS_COLUMNS) * TILE_SIDE, (i // ATLAS_COLUMNS) * TILE_SIDE
        for y in range(TILE_SIDE):
            start = (oy + y) * width + ox
            px[start:start + TILE_SIDE] = tile[y * TILE_SIDE:(y + 1) * TILE_SIDE]
    return bytes(px), width, height, count


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--out", required=True)
    args = p.parse_args()

    out = pathlib.Path(args.out)
    if ALLOWED_OUTPUT_PARENT not in pathlib.PurePosixPath(out.as_posix()).parents:
        raise SystemExit(f"輸出必須位於 {ALLOWED_OUTPUT_PARENT}/ 底下")
    zip_path = pathlib.Path(args.zip)
    digest = hashlib.sha256(zip_path.read_bytes()).hexdigest()
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")

    with zipfile.ZipFile(zip_path) as zf:
        _, parts = payloads(read_volumes(zf, "KOLPICS"), "KOLPICS")
        _, parts_town = payloads(read_volumes(zf, "TOWN0"), "TOWN0")

    out.mkdir(parents=True, exist_ok=True)
    images = []
    for payload, name in PAYLOADS.items():
        seg = lzw_decode(parts[payload])
        pixels, width, height, count = atlas(seg)
        png = encode_rgb_png(pixels, width, height)
        (out / name).write_bytes(png)
        images.append({
            "file": name, "payload_index": payload,
            "payload_bytes": len(seg), "tile_count": count,
            "tile_side": TILE_SIDE, "atlas_columns": ATLAS_COLUMNS,
            "width": width, "height": height,
            "payload_sha256": hashlib.sha256(seg).hexdigest(),
            "png_sha256": hashlib.sha256(png).hexdigest(),
        })

    maps = []
    if len(parts_town) < MAP_PAYLOAD_BASE + MAP_COUNT:
        raise SystemExit("TOWN0 缺少競技場地圖 payload 85")
    for i in range(MAP_COUNT):
        raw = lzw_decode(parts_town[MAP_PAYLOAD_BASE + i])
        cells = rle_decode(raw)
        if any(tile >= 80 for tile in cells):
            raise SystemExit(f"戰場地圖 {i} 的磚塊編號超出實際 atlas 的 80 格")
        png = encode_gray_png(cells, MAP_SIDE, MAP_SIDE)
        name = f"battle-map-{i:02d}.png"
        (out / name).write_bytes(png)
        maps.append({"file": name, "site": i, "payload": MAP_PAYLOAD_BASE + i,
                     "kind": "arena" if i == 24 else "quest",
                     "source_archive": "TOWN0",
                     "payload_sha256": hashlib.sha256(raw).hexdigest(),
                     "compressed_bytes": len(raw), "side": MAP_SIDE,
                     "cell_index": "(X << 6) + Y",
                     "png_sha256": hashlib.sha256(png).hexdigest()})

    manifest = {
        "schema": "kol-battle-tiles-v1",
        "tool": "tools/extract_battle_tiles.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source_archive": "Knights-of-legend/KOLPICS",
        "evidence": "RE-079；磚塊版面同 RE-059",
        "encoding": "一塊磚 16×16，四個平面各 32 byte，plane-major；像素值是 EGA 索引",
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "images": images,
        "maps": maps,
        "map_encoding": "TOWN0 payload 0x3D + i，RLE 壓過；解出來是 64×64 的格子陣列，"
                        "像素值即磚塊編號，索引 (X << 6) + Y",
    }
    (out / "manifest.json").write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"images": len(images), "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
