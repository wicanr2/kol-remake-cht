#!/usr/bin/env python3
"""新角色的預設盾牌（SPEC-036）：BOOKDATA 第 4 項 → PNG。

KOLBOOK 開機時 `sub_13E3F("bookdata", 4)` 把這一項 LZW 解到 DS:0x91A；建角 `sub_11820`
在 `0x1263D` 抄 760 byte 到 SHLDPICS 那一槽，所以每一名新角色的盾牌都是這一張，
之後只能在 EDIT SHIELD 改。解出來 761 byte，前 760 是 40 × 38 的 4 位元像素
（一列 20 byte，高 nibble 在左），最後多一個 0x00。

同時核對原版 SHLDPICS 的 16 槽：非空槽是不是都等於這一張（匯入的原版角色沿用這個
預設的前提）。不相等就失敗。

  tools/run_python_tool.sh tools/extract_default_shield.py \
      --out assets/original-derived-png/bookdata
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from extract_ega_assets import encode_rgba_png, verify_png_bytes  # noqa: E402
from inventory_image_archives import EXPECTED_ZIP_SHA256  # noqa: E402
from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-png")
WIDTH, HEIGHT, BYTES, SLOTS = 40, 38, 760, 16
ITEM = 4
SHLDPICS = "Knights-of-legend/SHLDPICS"


def nibbles(raw: bytes) -> bytes:
    out = bytearray()
    for b in raw:
        out += bytes((b >> 4, b & 0x0F))
    return bytes(out)


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--out", required=True)
    args = p.parse_args()
    out = pathlib.Path(args.out)
    if ALLOWED_OUTPUT_PARENT not in pathlib.PurePosixPath(out.as_posix()).parents:
        raise SystemExit(f"輸出必須位於 {ALLOWED_OUTPUT_PARENT}/ 底下")
    zip_bytes = pathlib.Path(args.zip).read_bytes()
    if hashlib.sha256(zip_bytes).hexdigest() != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")

    with zipfile.ZipFile(args.zip) as zf:
        _, parts = payloads(read_volumes(zf, "BOOKDATA"), "BOOKDATA")
        body = lzw_decode(parts[ITEM])
        slots = zf.read(SHLDPICS)
    if len(body) != BYTES + 1 or body[BYTES] != 0:
        raise SystemExit(f"BOOKDATA 第 {ITEM} 項解出 {len(body)} byte，預期 761 且末 byte 為 0")
    shield = body[:BYTES]
    if len(slots) != BYTES * SLOTS:
        raise SystemExit(f"SHLDPICS 長度 {len(slots)}，預期 {BYTES * SLOTS}")
    mismatched = []
    empty = 0
    for i in range(SLOTS):
        rec = slots[i * BYTES:(i + 1) * BYTES]
        if not any(rec):
            empty += 1
        elif rec != shield:
            mismatched.append(i)
    if mismatched:
        raise SystemExit(f"SHLDPICS 槽 {mismatched} 與預設盾牌不同，匯入角色不能沿用預設")

    png = encode_rgba_png(nibbles(shield), b"\x01" * (WIDTH * HEIGHT), WIDTH, HEIGHT)
    verify_png_bytes(png, WIDTH, HEIGHT)
    out.mkdir(parents=True, exist_ok=True)
    (out / "default-shield.png").write_bytes(png)
    manifest = {
        "schema": "kol-default-shield-v1",
        "tool": "tools/extract_default_shield.py",
        "source_zip_sha256": EXPECTED_ZIP_SHA256,
        "source_entry": "Knights-of-legend/BOOKDATA.0",
        "source_item": ITEM,
        "source_item_sha256": hashlib.sha256(parts[ITEM]).hexdigest(),
        "decoded_sha256": hashlib.sha256(shield).hexdigest(),
        "evidence": "SPEC-036；KOLBOOK sub_13E3F(\"bookdata\", 4) → DS:0x91A，sub_11820:0x1263D 抄 0x2F8 byte",
        "palette": "BIOS mode 0Dh EGA 16 色，索引即原版 4-bit 值",
        "shldpics_check": {"entry": SHLDPICS, "slots": SLOTS, "empty": empty,
                           "equal_to_default": SLOTS - empty},
        "images": [{"file": "default-shield.png", "width": WIDTH, "height": HEIGHT,
                    "png_sha256": hashlib.sha256(png).hexdigest()}],
        "evidence_status": "已證實",
        "rights": "原作衍生 PNG；依 2026-09-01 決定只存 private GitHub，不得公開發布",
    }
    (out / "manifest.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n",
                                       encoding="utf-8")
    print(json.dumps({"out": str(out), "equal": SLOTS - empty, "empty": empty}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
