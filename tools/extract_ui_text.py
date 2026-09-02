#!/usr/bin/env python3
"""三份 UI 文字庫的抽取器（RE-063 的版面，RE-067 起作為譯文來源）。

`sub_1A8FD(記錄號, 文字庫)` 依 `0x0D` 分隔取第 N 段，是全遊戲的訊息入口。
文字庫是 archive 的 payload，LZW 解碼之後才是明文：

  KOLDATA payload 1    城鎮與戰鬥的 UI 訊息
  BOOKDATA payload 1   書本 UI 與工作人員名單
  ENDTXT  payload 0    結局敘事

標記語法：`~Vxx`／`~Hxx` 是十六進位座標、`~Cxx` 是顏色、`@` 是換行、
`|X` 是代換記號。這裡原樣保留，翻譯時要照樣搬。

  python3 tools/extract_ui_text.py --out assets/original-derived-data/ui-text.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from inventory_image_archives import EXPECTED_ZIP_SHA256  # noqa: E402
from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-data")
SEPARATOR = 0x0D
BANKS = (("KOLDATA", 1, "城鎮與戰鬥的 UI 訊息"),
         ("BOOKDATA", 1, "書本 UI 與工作人員名單"),
         ("ENDTXT", 0, "結局敘事"))


def printable(raw: bytes) -> str:
    return "".join(chr(b) if 32 <= b < 127 else f"\\x{b:02x}" for b in raw)


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

    banks = []
    with zipfile.ZipFile(zip_path) as zf:
        for name, index, note in BANKS:
            _, parts = payloads(read_volumes(zf, name), name)
            body = lzw_decode(parts[index])
            records = body.split(bytes([SEPARATOR]))
            # 尾端若是空段就不是一筆記錄，只是最後一個分隔符。
            if records and records[-1] == b"":
                records.pop()
            banks.append({
                "bank": name, "payload": index, "note": note,
                "length": len(body),
                "sha256": hashlib.sha256(body).hexdigest(),
                "record_count": len(records),
                "records": [{"index": i, "text": printable(r)}
                            for i, r in enumerate(records)],
            })

    doc = {
        "schema": "kol-ui-text-data-v1",
        "tool": "tools/extract_ui_text.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "evidence": "RE-063；RE-067",
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "markup": {"~Vxx": "垂直座標（十六進位）", "~Hxx": "水平座標（十六進位）",
                   "~Cxx": "顏色", "@": "換行", "|X": "代換記號"},
        "separator": SEPARATOR,
        "banks": banks,
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    tmp = out.with_suffix(out.suffix + f".tmp-{os.getpid()}")
    tmp.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    os.replace(tmp, out)
    print(json.dumps({"banks": [(b["bank"], b["record_count"]) for b in banks],
                      "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
