#!/usr/bin/env python3
"""比對正常玩家路徑執行前後的角色檔，逐槽列出實際被寫入的 offset。

輸入是固定原版 ZIP 內的 `CHARDATA`／`CHARPICS`／`SHLDPICS`，以及
`tools/dosbox_character_ui_oracle.py --save-character-files` 存下的執行後版本。
輸出只有槽號、offset 區間與 byte 值，不含整份原版記錄，可版控。

  python3 tools/diff_character_files.py --zip kol.zip --after 目錄 [--out 差異.json]
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys
import zipfile

STRIDES = {"CHARDATA": 768, "CHARPICS": 760, "SHLDPICS": 760}
SLOTS = 16


def load_original(zip_path: pathlib.Path) -> dict[str, bytes]:
    out = {}
    with zipfile.ZipFile(zip_path) as zf:
        for info in zf.infolist():
            name = info.filename.split("/")[-1]
            if name in STRIDES:
                out[name] = zf.read(info)
    missing = set(STRIDES) - set(out)
    if missing:
        raise SystemExit(f"ZIP 缺少角色檔：{sorted(missing)}")
    return out


def runs_of_differences(a: bytes, b: bytes) -> list[tuple[int, int]]:
    """把不同的 byte 位置壓成 (起, 迄含) 區間。"""
    spans: list[tuple[int, int]] = []
    start = None
    for i in range(max(len(a), len(b))):
        x = a[i] if i < len(a) else None
        y = b[i] if i < len(b) else None
        if x != y:
            if start is None:
                start = i
        elif start is not None:
            spans.append((start, i - 1))
            start = None
    if start is not None:
        spans.append((start, max(len(a), len(b)) - 1))
    return spans


def diff_file(name: str, before: bytes, after: bytes) -> dict:
    stride = STRIDES[name]
    slots = []
    for slot in range(SLOTS):
        lo, hi = slot * stride, (slot + 1) * stride
        b0, b1 = before[lo:hi], after[lo:hi]
        if b0 == b1:
            continue
        spans = runs_of_differences(b0, b1)
        slots.append({
            "slot": slot,
            "changed_bytes": sum(e - s + 1 for s, e in spans),
            "spans": [
                {
                    "offset": f"0x{s:03x}",
                    "length": e - s + 1,
                    "before": b0[s:e + 1].hex(),
                    "after": b1[s:e + 1].hex(),
                }
                for s, e in spans
            ],
        })
    return {
        "file": name,
        "stride": stride,
        "bytes": len(after),
        "sha256_before": hashlib.sha256(before).hexdigest(),
        "sha256_after": hashlib.sha256(after).hexdigest(),
        "changed_slots": [s["slot"] for s in slots],
        "slots": slots,
    }


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", required=True)
    p.add_argument("--after", required=True, help="含執行後角色檔的目錄")
    p.add_argument("--out", default=None)
    p.add_argument("--max-span-hex", type=int, default=64,
                   help="每個區間輸出的 hex byte 上限，超過只留長度")
    args = p.parse_args()

    before = load_original(pathlib.Path(args.zip))
    afterdir = pathlib.Path(args.after)
    result = {"tool": "tools/diff_character_files.py", "files": []}
    for name in STRIDES:
        after = (afterdir / name).read_bytes()
        entry = diff_file(name, before[name], after)
        for slot in entry["slots"]:
            for span in slot["spans"]:
                if span["length"] > args.max_span_hex:
                    span["before"] = f"<{span['length']} bytes>"
                    span["after"] = f"<{span['length']} bytes>"
        result["files"].append(entry)
    text = json.dumps(result, ensure_ascii=False, indent=2) + "\n"
    if args.out:
        pathlib.Path(args.out).write_text(text)
    print(text)
    return 0


if __name__ == "__main__":
    sys.exit(main())
