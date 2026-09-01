#!/usr/bin/env python3
"""從固定原版輸入抽出翻書動畫的參數表。

remake 的動畫只讀這份 typed data，執行期不解析 `KOLBOOK.EXE`。表的位置、筆數與
stride 出自 RE-048 的 IDA 探針；這裡以位元組序列在 EXE 內唯一定位，找不到或找到多處
一律失敗，不猜。

  python3 tools/extract_book_animation.py --zip kol.zip --repo . \\
      --out assets/original-derived-data/book-open-animation.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import pathlib
import sys
import zipfile

EXPECTED_ZIP_SHA256 = "2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46"
EXE_NAME = "Knights-of-legend/KOLBOOK.EXE"
EXPECTED_EXE_SHA256 = "2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832"
ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-data")

# 兩張表的內容與雜湊出自 tools/ida_probe_kolbook_book_animation.py。
STAGES = [
    {"name": "stage_one", "stride": 2, "count": 17,
     "sha256": "f06f3fd94779cc601ef453f7e7e65bd292996875b7c19d79c1a9cad4212db825"},
    {"name": "stage_two", "stride": 3, "count": 18,
     "sha256": "76e329a65c30164925b2b6ab906a385109298035ea05fd0828b62711dd2efc8e"},
]
STAGE_BYTES = {
    "stage_one": bytes.fromhex(
        "0011020e040b060808050a02000b0208040506020007020404010004020100020001"),
    "stage_two": bytes.fromhex(
        "1100001001060f020a0e01040d040d0c01010b04090a070f090203080507070"
        "80c060b10050202040505030808020b0b010e0e001111"),
}

# 程式碼裡的固定量：每一步等 3 個 tick，動畫只動第 10 列起的 180 列，
# 欄的單位是 EGA 的 byte column（8 像素）。
DELAY_TICKS = 3
FIRST_ROW = 10
ROW_COUNT = 180
COLUMN_PIXELS = 8
STAGE_ONE_BASE = 0x14   # 第一段的欄位起點偏移
STAGE_TWO_BASE = 0x02   # 第二段的欄位起點偏移


def sha(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def locate(haystack: bytes, needle: bytes, label: str) -> int:
    first = haystack.find(needle)
    if first < 0:
        raise SystemExit(f"{label} 在 EXE 內找不到")
    if haystack.find(needle, first + 1) >= 0:
        raise SystemExit(f"{label} 在 EXE 內出現多次，無法唯一定位")
    return first


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", required=True)
    p.add_argument("--repo", default=".")
    p.add_argument("--out", required=True)
    args = p.parse_args()

    repo = pathlib.Path(args.repo).resolve()
    out = pathlib.Path(args.out).resolve()
    try:
        relative = out.relative_to(repo)
    except ValueError:
        raise SystemExit("輸出必須位於 repo 內")
    if ALLOWED_OUTPUT_PARENT != pathlib.PurePosixPath(relative.as_posix()).parent:
        raise SystemExit(f"輸出必須位於 {ALLOWED_OUTPUT_PARENT}/ 底下")

    zip_path = pathlib.Path(args.zip)
    digest = sha(zip_path.read_bytes())
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit(f"非預期 ZIP SHA-256：{digest}")
    with zipfile.ZipFile(zip_path) as archive:
        exe = archive.read(EXE_NAME)
    if sha(exe) != EXPECTED_EXE_SHA256:
        raise SystemExit("KOLBOOK.EXE 已改變")

    stages = []
    for spec in STAGES:
        raw = STAGE_BYTES[spec["name"]]
        if sha(raw) != spec["sha256"]:
            raise SystemExit(f"{spec['name']} 的內建位元組與記錄的雜湊不符")
        if len(raw) != spec["stride"] * spec["count"]:
            raise SystemExit(f"{spec['name']} 的長度與 stride×count 不符")
        offset = locate(exe, raw, spec["name"])
        rows = [list(raw[i:i + spec["stride"]])
                for i in range(0, len(raw), spec["stride"])]
        stages.append({"name": spec["name"], "file_offset": offset,
                       "stride": spec["stride"], "count": spec["count"],
                       "sha256": spec["sha256"], "rows": rows})

    steps = []
    for a, b in stages[0]["rows"]:
        steps.append({"stage": "stage_one", "scroll_column": STAGE_ONE_BASE + a,
                      "scroll_span": b,
                      "fill_column": STAGE_ONE_BASE + a + b,
                      "source_column": STAGE_ONE_BASE + a + b})
    steps.append({"stage": "between", "scroll_column": STAGE_ONE_BASE, "scroll_span": 0,
                  "fill_column": STAGE_ONE_BASE, "source_column": STAGE_ONE_BASE})
    for c, d, e in stages[1]["rows"]:
        steps.append({"stage": "stage_two", "scroll_column": STAGE_TWO_BASE + c,
                      "scroll_span": d,
                      "fill_column": STAGE_TWO_BASE + c + d,
                      "source_column": STAGE_TWO_BASE + e})

    data = {
        "schema": "kol-book-open-animation-v1",
        "tool": "tools/extract_book_animation.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source_exe": {"name": EXE_NAME, "sha256": EXPECTED_EXE_SHA256},
        "evidence": "docs/re/ida-fingerprints/KOLBOOK-book-animation-summary.json",
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版執行檔的資料表，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "geometry": {"first_row": FIRST_ROW, "row_count": ROW_COUNT,
                     "column_pixels": COLUMN_PIXELS},
        "delay_ticks_per_step": DELAY_TICKS,
        "rule": "每一步先把 scroll_column 起 scroll_span 個 byte 欄整體左移一欄"
                "（只動 first_row 起的 row_count 列），再把來源畫面的 source_column "
                "欄複製到畫面的 fill_column 欄。",
        "stages": stages,
        "steps": steps,
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    tmp = out.with_suffix(out.suffix + f".tmp-{os.getpid()}")
    tmp.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n")
    os.replace(tmp, out)
    print(json.dumps({"steps": len(steps),
                      "stage_offsets": [s["file_offset"] for s in stages],
                      "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
