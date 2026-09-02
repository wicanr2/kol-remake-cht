#!/usr/bin/env python3
"""探針：城鎮建築 payload 的前 128 byte（文字格之前那一段）。

`TOWNn.0` 的建築 payload 從位移 128 起是 16 byte 一格的 5-bit 文字（RE-062／064）。
前面那 128 byte 目前只知道兩件事：位移 0 是建築名、位移 29 是人物名，都是
「長度前綴＋末字元設高位元」。本探針把整段原樣倒出來，找其餘欄位。

  tools/run_python_tool.sh tools/probe_building_header.py --town 1 --door 1
"""

from __future__ import annotations

import argparse
import hashlib
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from inventory_image_archives import EXPECTED_ZIP_SHA256  # noqa: E402
from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402
from probe_town_doors import text_payload_index  # noqa: E402

TEXT_BASE = 128


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--town", type=int, required=True)
    p.add_argument("--door", type=int, required=True)
    args = p.parse_args()
    zip_path = pathlib.Path(args.zip)
    if hashlib.sha256(zip_path.read_bytes()).hexdigest() != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")
    with zipfile.ZipFile(zip_path) as zf:
        root = f"TOWN{args.town}"
        _, parts = payloads(read_volumes(zf, root), root)
        body = lzw_decode(parts[text_payload_index(args.door)])
    print(f"payload 長度 {len(body)}")
    for base in range(0, TEXT_BASE, 16):
        row = body[base:base + 16]
        text = "".join(chr(b & 0x7F) if 32 <= (b & 0x7F) < 127 else "." for b in row)
        print(f"{base:04x}  {row.hex(' ')}  {text}")
    print("--- 文字格之後（若有）")
    slots = len(body) // 16
    print(f"共 {slots} 格；最後 4 格：")
    for base in range((slots - 4) * 16, slots * 16, 16):
        row = body[base:base + 16]
        text = "".join(chr(b & 0x7F) if 32 <= (b & 0x7F) < 127 else "." for b in row)
        print(f"{base:04x}  {row.hex(' ')}  {text}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
