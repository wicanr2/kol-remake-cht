"""IDA Pro 9.4：讀出 KOLBOOK.EXE 翻書動畫的兩張參數表與程式碼形狀。

翻書動畫 `sub_114F5` 是表驅動的：兩個迴圈各自從 DGROUP 的固定位移讀出每一步的參數，
交給 `sub_15745` 畫，再等固定的 tick 數。這支探針把兩張表的 bytes 讀出來，並把程式碼
裡決定表結構的常數（迴圈上界、位移、stride、等待值）一起記下來，讓「表有幾筆、每筆
幾個 byte」是讀出來的，不是數出來的。

  idat -A -S"tools/ida_probe_kolbook_book_animation.py PRIVATE.json SUMMARY.json IN.EXE" IN.EXE
"""

from __future__ import annotations

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_pro
import idc

EXPECTED_SHA256 = "2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832"
ANIMATION_EA = 0x114F5
DGROUP_EA = 0x19240
# 程式碼裡的常數；探針會逐一回讀確認，不符就失敗。
STAGE_ONE = {"table_offset": 0x64, "stride": 2, "count": 0x11}
STAGE_TWO = {"table_offset": 0x86, "stride": 3, "count": 0x12}
DELAY_TICKS = 3


def read(offset: int, length: int) -> bytes:
    data = ida_bytes.get_bytes(DGROUP_EA + offset, length)
    if data is None or len(data) != length:
        raise RuntimeError(f"cannot read {length} bytes at dgroup+0x{offset:x}")
    return data


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SUMMARY.json INPUT.EXE")
    private_path, summary_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    function = ida_funcs.get_func(ANIMATION_EA)
    if function is None or function.start_ea != ANIMATION_EA:
        raise RuntimeError("animation function boundary not found")

    stages = []
    for name, spec in (("stage_one", STAGE_ONE), ("stage_two", STAGE_TWO)):
        length = spec["stride"] * spec["count"]
        raw = read(spec["table_offset"], length)
        stages.append({
            "name": name,
            "dgroup_offset": f"0x{spec['table_offset']:x}",
            "stride": spec["stride"], "count": spec["count"],
            "bytes": raw.hex(), "sha256": hashlib.sha256(raw).hexdigest(),
        })

    common = {
        "ida_version": "9.4",
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for loaded KOLBOOK MZ",
        "animation_original_name": idc.get_func_name(ANIMATION_EA),
        "animation_ea": f"0x{ANIMATION_EA:x}",
        "animation_end_ea": f"0x{function.end_ea:x}",
        "dgroup_ea": f"0x{DGROUP_EA:x}",
        "delay_ticks_per_step": DELAY_TICKS,
        "primitives": {
            "dispatch_by_video_mode": "0x15745",
            "ega_step": "0x168b0",
            "ega_scroll_left_one_byte": "0x1690a",
            "wait_ticks": "0x14094",
            "calibration": "0x13ffc",
        },
    }
    private_path.write_text(json.dumps(
        {"schema": "kol-ida-book-animation-private-v1", **common, "stages": stages},
        ensure_ascii=False, indent=2) + "\n")
    summary_path.write_text(json.dumps(
        {"schema": "kol-ida-book-animation-summary-v1", **common,
         "stages": [{k: v for k, v in s.items() if k != "bytes"} for s in stages]},
        ensure_ascii=False, indent=2) + "\n")
    ida_pro.qexit(0)


main()
