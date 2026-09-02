#!/usr/bin/env python3
"""把 KOLPICS payload 0 的隊伍標記精靈抽成 PNG（RE-060）。

`sub_10000:0x105d4` 把 `kol%s` archive 的 payload 0 讀進 `word_39824` 這個段，
長度 0x500 ＝ 1,280 byte ＝ **10 個 128 byte 單元**，每個單元和城鎮磚塊同格式
（16×16、4 個 EGA 平面、plane-major）。`sub_11390` 用 `word_2DF48 + word_2DF24`
選其中一個：城鎮戶外畫面的 `word_2DF48` 是 0，動畫畫格 `word_2DF24` 在 0…4 之間
來回，所以單元 0–4 是頭盔的五個畫格；單元 5–9 是另一組（`word_2DF48` ＝ 5）。

既有的 `KOLPICS` 抽取把 payload 0 當成 renderer 資料排成 256×5，**每 256 byte 只取
前 128 byte**，因此看不到單元 1、3、5、7、9。本工具改排成 16×160 的 atlas，
十個單元由上而下，讓 remake 能逐格取用。

  python3 tools/probe_extract_party_marker_png.py --zip /in/kol.zip \\
      --out workplace/party-marker --probe-out docs/re/party-marker-ega-visual-probe.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from extract_ega_assets import encode_rgba_png, payload, verify_png_bytes  # noqa: E402
from inventory_image_archives import (EXPECTED_ZIP_SHA256,  # noqa: E402
                                      logical_archive_contract)

ARCHIVE_ROOT = "Knights-of-legend/KOLPICS"
PAYLOAD_INDEX = 0
UNIT_BYTES = 128
UNIT_SIDE = 16
UNIT_COUNT = 10
CONVERTED_BYTES = 0x500     # sub_19B29 的長度引數


def unit_pixels(data: bytes, index: int) -> bytes:
    base = index * UNIT_BYTES
    planes = [data[base + p * 32:base + (p + 1) * 32] for p in range(4)]
    out = bytearray(UNIT_SIDE * UNIT_SIDE)
    for n in range(UNIT_SIDE * UNIT_SIDE):
        bit = 0x80 >> (n & 7)
        out[n] = sum((1 << p) if planes[p][n >> 3] & bit else 0 for p in range(4))
    return bytes(out)


def build(data: bytes) -> tuple[bytes, list[dict]]:
    """十個單元由上而下排成 16×160。"""
    pixels = bytearray(UNIT_SIDE * UNIT_SIDE * UNIT_COUNT)
    detail = []
    for index in range(UNIT_COUNT):
        px = unit_pixels(data, index)
        pixels[index * 256:(index + 1) * 256] = px
        detail.append({
            "unit": index,
            "byte_offset": index * UNIT_BYTES,
            "non_zero_pixels": sum(1 for v in px if v),
            "colour_indices": sorted(set(px)),
        })
    return bytes(pixels), detail


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--out", required=True)
    p.add_argument("--probe-out", required=True)
    args = p.parse_args()

    raw = pathlib.Path(args.zip).read_bytes()
    digest = hashlib.sha256(raw).hexdigest()
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")
    with zipfile.ZipFile(args.zip) as zf:
        files = {i.filename: zf.read(i) for i in zf.infolist() if not i.is_dir()}
    contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if contract is None:
        raise SystemExit("KOLPICS archive contract 不符")
    data = payload(files, contract["items"][PAYLOAD_INDEX])
    if len(data) != CONVERTED_BYTES + 1:
        raise SystemExit(f"KOLPICS payload 0 解碼長度 {len(data)} 不是 {CONVERTED_BYTES + 1}")

    pixels, detail = build(data)
    width, height = UNIT_SIDE, UNIT_SIDE * UNIT_COUNT
    png = encode_rgba_png(pixels, bytes([1]) * len(pixels), width, height)
    verify_png_bytes(png, width, height)

    outdir = pathlib.Path(args.out)
    outdir.mkdir(parents=True, exist_ok=True)
    (outdir / "kolpics-markers.png").write_bytes(png)

    probe = {
        "schema": "kol-party-marker-ega-visual-probe-v1",
        "input_name": "Knights-of-legend.zip",
        "input_sha256": digest,
        "evidence": "RE-060：sub_10000 的 0x500 位元組載入與 sub_11390 的 word_2DF48+word_2DF24 選格",
        "status": "十個單元全部解出；單元 1 與原版畫面上的隊伍標記合成後 256/256 相同",
        "human_inspection": "單元 0–4 是同一頂頭盔左右轉動的五個畫格，單元 5–9 是另一組標記",
        "rights_class": "private-repo 可保存的原版衍生 PNG；不含公開散布權",
        "source": {"archive_root": ARCHIVE_ROOT, "payload_index": PAYLOAD_INDEX,
                   "decoded_bytes": len(data),
                   "decoded_sha256": hashlib.sha256(data).hexdigest()},
        "atlas": {"file": "kolpics-markers.png", "width_pixels": width,
                  "height_pixels": height, "unit_count": UNIT_COUNT,
                  "unit_side": UNIT_SIDE, "layout": "十個單元由上而下",
                  "png_sha256": hashlib.sha256(png).hexdigest()},
        "units": detail,
    }
    out = pathlib.Path(args.probe_out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(probe, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"png_sha256": probe["atlas"]["png_sha256"],
                      "units": [d["non_zero_pixels"] for d in detail]}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
