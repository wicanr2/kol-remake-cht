#!/usr/bin/env python3
"""把野外治療那一條的六段敘述抽出來（RE-112 續，RE-128 解出來源）。

怪物編號 0xFF 那一條不是任務怪，是**治療的遭遇**：低位階隊伍才擲得到，答應之後
把六名角色的傷勢與飢餓清掉。它的敘述不在怪物記錄裡，而在 DGROUP `0x2350`——
那一塊在 BSS，整份反組譯零個參照點。零不是答案：換一個過濾器掃 `sub_1A4FE` 的
每一個載入點，就看到 `sub_29004:0x2905E` 把 **`TOWN0` 的 payload 4** 載進
DGROUP `0x22D0`，而 `0x2350` 正是它的 `+0x80`。

版面是六段 × 四行 × 16 byte：**每行的第 0 個 byte 是要解幾個 5-bit 值**，其餘
15 個 byte 是那些值（MSB-first）。這是 5-bit 名稱的第二種用法——建築與商品那一種
（`packed_name`）是「高 nibble 是字數、字元從第 4 個 bit 起」，最多 15 個字；
這一種整個 byte 都是字數，一行最多 24 個。不照字數截斷就會在句尾拖出一串垃圾字母。

  tools/run_python_tool.sh tools/extract_wilderness_scenes.py --zip /in/kol.zip \\
      --out assets/original-derived-data/wilderness-scenes.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from extract_town_text import render, slot_values  # noqa: E402
from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402

ARCHIVE = "TOWN0"
# `sub_29004:0x2905E` 載的是 entry 4。
PAYLOAD = 4
# DGROUP `0x2350` − `0x22D0`（緩衝起點）。
SCENE_BASE = 0x80
SCENES, LINES, LINE_BYTES = 6, 4, 0x10
# 一行最多解幾個 5-bit 值（15 byte × 8 ÷ 5）。
MAX_CHARS = 24


def scene(body: bytes, index: int) -> list[str]:
    """一段四行。每行照自己的第 0 個 byte 截斷。"""
    out = []
    for row in range(LINES):
        off = SCENE_BASE + index * LINES * LINE_BYTES + row * LINE_BYTES
        count = body[off]
        if count > MAX_CHARS:
            raise SystemExit(f"第 {index} 段第 {row} 行的字數是 {count}，超過 {MAX_CHARS}")
        text, unknown = render(slot_values(body, off, count))
        if unknown:
            raise SystemExit(f"第 {index} 段第 {row} 行有未定的跳脫序列 {unknown}")
        out.append(text.rstrip())
    return out


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--zip", required=True)
    ap.add_argument("--out", required=True)
    args = ap.parse_args()

    import zipfile
    with zipfile.ZipFile(args.zip) as zf:
        volumes = read_volumes(zf, ARCHIVE)
        contract, parts = payloads(volumes, ARCHIVE)
        if len(parts) <= PAYLOAD:
            raise SystemExit(f"{ARCHIVE} 只有 {len(parts)} 個 payload")
        raw = parts[PAYLOAD]
        body = lzw_decode(raw)
    need = SCENE_BASE + SCENES * LINES * LINE_BYTES
    if len(body) < need:
        raise SystemExit(f"payload {PAYLOAD} 解開之後是 {len(body)} 個 byte，"
                         f"至少要 {need}")

    scenes = [scene(body, i) for i in range(SCENES)]
    # 每一段都要有字：全空代表位置錯了，而位置錯一樣讀得出「看起來像資料」的東西。
    for i, lines in enumerate(scenes):
        if not any(lines):
            raise SystemExit(f"第 {i} 段是空的")

    doc = {
        "schema": "kol-wilderness-scenes-v1",
        "tool": pathlib.Path(__file__).name,
        "source": f"Knights-of-legend/{ARCHIVE}.0 payload {PAYLOAD}",
        "source_sha256": hashlib.sha256(raw).hexdigest(),
        "decoded_bytes": len(body),
        "decoded_sha256": hashlib.sha256(body).hexdigest(),
        "contract": contract,
        "dgroup_offset": "0x2350",
        "note": "治療那一條（怪物編號 0xFF）的六段敘述。偶數段是遠景、奇數段是走近"
                "之後看到的；原版先印 `索引`、騎走那一問過了再印 `索引 + 1`（RE-112）。",
        "scenes": scenes,
    }
    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(doc, ensure_ascii=False, indent=1) + "\n",
                   encoding="utf-8")
    print(json.dumps({"out": str(out), "scenes": len(scenes),
                      "lines": sum(len(s) for s in scenes)}, ensure_ascii=False))
    return 0


raise SystemExit(main())
