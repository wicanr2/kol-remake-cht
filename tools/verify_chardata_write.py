#!/usr/bin/env python3
"""remake 的 CHARDATA writer 對原版實際寫入的逐 byte 覆蓋率。

原版建立角色時只改動槽位裡的一部分 bytes。這支工具拿原版寫入前後的 `CHARDATA`
算出實際被改動的位移，再和 remake writer 產出的記錄逐 byte 比對，逐項報告：

- 原版改動而 remake 也寫對的位移
- 原版改動但 remake 沒有寫（來源尚未反追的欄位）
- 原版改動而 remake 寫了不同值（真正的錯誤）
- remake 寫了但原版沒改的位移（多寫）

`CHARDATA` 含原版樣板 bytes，只作本機比對，不入庫；入庫的是這份收據裡的位移與統計。

  python3 tools/verify_chardata_write.py --before CHARDATA.before --after CHARDATA.after \\
      --record workplace/chardata-write.json --out docs/verification/chardata-write.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys

RECORD_BYTES = 768


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--before", required=True)
    p.add_argument("--after", required=True)
    p.add_argument("--record", required=True)
    p.add_argument("--out", required=True)
    args = p.parse_args()

    before = pathlib.Path(args.before).read_bytes()
    after = pathlib.Path(args.after).read_bytes()
    if len(before) != len(after):
        raise SystemExit("寫入前後的 CHARDATA 長度不同")
    doc = json.loads(pathlib.Path(args.record).read_text())
    if doc.get("schema") != "kol-chardata-write-v1":
        raise SystemExit("非預期的 record schema")
    slot = doc["slot"]
    remake = bytes.fromhex(doc["record_hex"])
    if len(remake) != RECORD_BYTES:
        raise SystemExit("remake 記錄長度不符")

    base = slot * RECORD_BYTES
    orig_before = before[base:base + RECORD_BYTES]
    orig_after = after[base:base + RECORD_BYTES]
    changed = [i for i in range(RECORD_BYTES) if orig_before[i] != orig_after[i]]
    written = set(doc["written_offsets"])

    matched, unwritten, wrong, extra = [], [], [], []
    for i in changed:
        if remake[i] == orig_after[i]:
            matched.append(i)
        elif i in written:
            wrong.append({"offset": i, "remake": remake[i], "original": orig_after[i]})
        else:
            unwritten.append(i)
    for i in sorted(written):
        if i not in changed and remake[i] != orig_after[i]:
            extra.append({"offset": i, "remake": remake[i], "original": orig_after[i]})

    receipt = {
        "schema": "kol-chardata-write-verification-v1",
        "tool": "tools/verify_chardata_write.py",
        "slot": slot,
        "character": doc["character"],
        "contract": doc["contract"],
        "before_sha256": hashlib.sha256(before).hexdigest(),
        "after_sha256": hashlib.sha256(after).hexdigest(),
        "criterion": "原版改動的每一個 byte，remake 都要寫出相同的值；來源尚未反追的欄位"
                     "列為未寫，寫錯與多寫都必須是零。",
        "original_changed_bytes": len(changed),
        "matched_bytes": len(matched),
        "unwritten_bytes": len(unwritten),
        "wrong_bytes": len(wrong),
        "extra_bytes": len(extra),
        "coverage": round(len(matched) / len(changed), 4) if changed else 0.0,
        "matched_offsets": [f"0x{i:02x}" for i in matched],
        "unwritten_offsets": [f"0x{i:02x}" for i in unwritten],
        "wrong": wrong,
        "extra": extra,
    }
    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({k: receipt[k] for k in
                      ("original_changed_bytes", "matched_bytes", "unwritten_bytes",
                       "wrong_bytes", "extra_bytes", "coverage")}, ensure_ascii=False))
    return 1 if wrong or extra else 0


if __name__ == "__main__":
    sys.exit(main())
