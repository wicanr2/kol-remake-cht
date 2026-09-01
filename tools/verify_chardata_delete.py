#!/usr/bin/env python3
"""刪除角色的逐 byte 對拍。

建立角色的屬性是擲骰出來的，兩邊不可能逐 byte 相同；刪除沒有這個問題——同一份
`CHARDATA`、刪同一個槽位，原版與 remake 應該產出**完全相同的檔案**。因此這裡的判準
是全檔相等，外加逐項列出原版實際改動的位移，確認 remake 沒有多寫也沒有漏寫。

  python3 tools/verify_chardata_delete.py --before 原版CHARDATA --original 原版執行後 \\
      --remake remake產出 --slot 1 --out 收據.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib

RECORD_BYTES = 768
SLOT_COUNT = 16


def diff(a: bytes, b: bytes) -> list[dict]:
    out = []
    for i, (x, y) in enumerate(zip(a, b)):
        if x != y:
            out.append({"offset": i, "slot": i // RECORD_BYTES,
                        "slot_offset": "0x%03x" % (i % RECORD_BYTES),
                        "before": "%02x" % x, "after": "%02x" % y})
    return out


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--before", required=True)
    p.add_argument("--original", required=True)
    p.add_argument("--remake", required=True)
    p.add_argument("--slot", type=int, required=True)
    p.add_argument("--out", required=True)
    args = p.parse_args()

    before = pathlib.Path(args.before).read_bytes()
    original = pathlib.Path(args.original).read_bytes()
    remake = pathlib.Path(args.remake).read_bytes()
    if not (len(before) == len(original) == len(remake) == RECORD_BYTES * SLOT_COUNT):
        raise SystemExit("三個檔案的長度必須都是 %d" % (RECORD_BYTES * SLOT_COUNT))

    original_diff = diff(before, original)
    remake_diff = diff(before, remake)
    identical = original == remake
    receipt = {
        "schema": "kol-chardata-delete-v1",
        "tool": "tools/verify_chardata_delete.py",
        "criterion": "同一份 CHARDATA 刪同一個槽位，原版與 remake 的輸出必須逐 byte 相同",
        "slot": args.slot,
        "sha256": {
            "before": hashlib.sha256(before).hexdigest(),
            "original_after": hashlib.sha256(original).hexdigest(),
            "remake_after": hashlib.sha256(remake).hexdigest(),
        },
        "original_changed_bytes": original_diff,
        "remake_changed_bytes": remake_diff,
        "identical": identical,
        "pass": identical and original_diff == remake_diff,
    }
    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"pass": receipt["pass"], "identical": identical,
                      "original_changed": len(original_diff),
                      "remake_changed": len(remake_diff)}))
    return 0 if receipt["pass"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
