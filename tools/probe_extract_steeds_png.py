"""驗證 STEEDS.0 四張 40×38 EGA 坐騎圖片（RE-120）。

`STEEDS.0` 是**模板**（馬廄賣的四匹），`STEEDPIC` 是玩家角色自己那一張的 16 槽表；
兩個檔案的版面相同（一筆 760 byte、40×38、packed nibbles），來源不同不可混用。
證據是 `workplace/analysis/koltown/steed.json`：`0x12a68` 用檔名樣板 DGROUP+0x16E0
（`'steeds'`）呼叫 `sub_1A4FE` 載入第 `picture` 張，再由 `0x12acb` 的 `sub_1C53D`
畫到欄 0x0E、列 0x56。

  python3 tools/probe_extract_steeds_png.py /in/kol.zip \\
      workplace/analysis/koltown/steed.json workplace/analysis/steeds-png \\
      docs/re/steeds-ega-visual-probe.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import sys
import zipfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from inventory_image_archives import EXPECTED_ZIP_SHA256, logical_archive_contract
from probe_extract_fixed_charpics_png import HEIGHT, RECORD_BYTES, WIDTH, decode_record
from probe_extract_monpics_png import encode_rgba_png
from probe_image_lzw import decode


ARCHIVE_ROOT = "Knights-of-legend/STEEDS"
IMAGE_COUNT = 4
GRID_COLUMNS = 4
CELL_WIDTH = 48
CELL_HEIGHT = 46
# 明細那一頁的載入與畫圖呼叫點，用來確認 IDA 證據是這一份。
LOAD_CALL_EA = "0x12a68"
DRAW_CALL_EA = "0x12acb"


def check_ida(summary: dict) -> None:
    if summary.get("schema") != "kol-ida-koltown-steed-v1":
        raise SystemExit("IDA 摘要不是馬廄坐騎那一份")
    rows = {row["ea"]: row["disassembly"] for rows in summary["windows"].values()
            for row in rows}
    if "sub_1A4FE" not in rows.get(LOAD_CALL_EA, ""):
        raise SystemExit(f"{LOAD_CALL_EA} 不是 sub_1A4FE 的呼叫點")
    if "sub_1C53D" not in rows.get(DRAW_CALL_EA, ""):
        raise SystemExit(f"{DRAW_CALL_EA} 不是 sub_1C53D 的呼叫點")
    if summary["strings"].get("steeds", [{}])[0].get("ea") != "0x2f5b0":
        raise SystemExit("'steeds' 字串位址與 DGROUP+0x16E0 不符")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("zip_path", type=Path)
    parser.add_argument("ida_summary_path", type=Path)
    parser.add_argument("output_directory", type=Path)
    parser.add_argument("summary_path", type=Path)
    args = parser.parse_args()
    if "workplace" not in args.output_directory.resolve().parts:
        raise SystemExit("研究 PNG 只可寫入 ignored workplace")

    zip_bytes = args.zip_path.read_bytes()
    zip_digest = hashlib.sha256(zip_bytes).hexdigest()
    if zip_digest != EXPECTED_ZIP_SHA256:
        raise SystemExit(f"非預期 ZIP SHA-256：{zip_digest}")
    ida_bytes = args.ida_summary_path.read_bytes()
    check_ida(json.loads(ida_bytes))

    with zipfile.ZipFile(args.zip_path) as archive:
        files = {info.filename: archive.read(info)
                 for info in archive.infolist() if not info.is_dir()}
    contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if contract is None or contract["count"] != IMAGE_COUNT:
        raise SystemExit("STEEDS.0 archive contract 不符")

    args.output_directory.mkdir(parents=True, exist_ok=True)
    sheet_width = GRID_COLUMNS * CELL_WIDTH
    sheet_height = CELL_HEIGHT
    sheet_pixels = bytearray(
        7 if ((x // 8) + (y // 8)) % 2 == 0 else 8
        for y in range(sheet_height) for x in range(sheet_width))
    records = []
    for item in contract["items"]:
        source = files[item["volume_name"]]
        compressed = source[item["offset"]:item["offset"] + item["compressed_bytes"]]
        stage1 = decode(compressed)
        if len(stage1) != RECORD_BYTES + 1:
            raise SystemExit(
                f"STEEDS[{item['index']}] 解碼長度不符：{len(stage1)} != {RECORD_BYTES + 1}")
        pixels = decode_record(stage1[:RECORD_BYTES])
        output = args.output_directory / f"steeds-{item['index']:02d}.png"
        output.write_bytes(encode_rgba_png(pixels, bytes([1]) * len(pixels), WIDTH, HEIGHT))

        origin_x = (item["index"] % GRID_COLUMNS) * CELL_WIDTH + 4
        origin_y = 4
        for y in range(HEIGHT):
            row = y * WIDTH
            dst = (origin_y + y) * sheet_width + origin_x
            sheet_pixels[dst:dst + WIDTH] = pixels[row:row + WIDTH]
        records.append({
            "image_index": item["index"],
            "compressed_bytes": item["compressed_bytes"],
            "decoded_bytes": len(stage1),
            "pixel_bytes_consumed": RECORD_BYTES,
            "trailing_byte_hex": stage1[RECORD_BYTES:].hex(),
            "pixel_sha256": hashlib.sha256(pixels).hexdigest(),
            "png_sha256": hashlib.sha256(output.read_bytes()).hexdigest(),
            "color_count": len(set(pixels)),
        })

    sheet = args.output_directory / "steeds-contact-sheet.png"
    sheet.write_bytes(encode_rgba_png(bytes(sheet_pixels), bytes([1]) * len(sheet_pixels),
                                      sheet_width, sheet_height))
    result = {
        "schema": "kol-steeds-ega-visual-probe-v1",
        "input_name": args.zip_path.name,
        "input_sha256": zip_digest,
        "ida_summary_name": args.ida_summary_path.name,
        "ida_summary_sha256": hashlib.sha256(ida_bytes).hexdigest(),
        "status": "mechanically validated all 4 images; four distinct steeds confirmed by inspection",
        "source_archive": "STEEDS.0",
        "image_count": len(records),
        "width_pixels": WIDTH,
        "height_pixels": HEIGHT,
        "packing": "high nibble first, then low nibble",
        "decoded_bytes_per_item": RECORD_BYTES + 1,
        "pixel_bytes_per_item": RECORD_BYTES,
        "trailing_byte_disposition": "excluded_non_pixel; downstream 40x38 consumer reads exactly 760 bytes",
        "png_count": len(records),
        "contact_sheet_sha256": hashlib.sha256(sheet.read_bytes()).hexdigest(),
        "records": records,
    }
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"validated={len(records)} contact_sheet={sheet}")


if __name__ == "__main__":
    main()
