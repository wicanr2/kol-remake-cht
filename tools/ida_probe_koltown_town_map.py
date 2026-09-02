"""IDA Pro 9.4：找出 KOLTOWN 載入 `TOWNn` 城鎮資料與繪製地圖的路徑。

這是**探索用**探針，不是鎖定斷言：先掃出所有含 `town` 的字串、它們的 xref，
以及參照者所屬函式的反組譯，讓下一輪可以寫成有 EXPECTED_* 斷言的鎖定探針。
另外掃出所有「立即值 0x80／128」與「shl cl,7」形態的位置，因為磚塊間距是 128 bytes。

  idat -A -S"tools/ida_probe_koltown_town_map.py OUT.json INPUT.EXE" INPUT.EXE
"""

from __future__ import annotations

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_lines
import idaapi
import idautils
import idc

EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
# 16-bit real mode 下，程式碼是把字串的 DGROUP 位移當立即值載入（`mov ax, 1776h`），
# IDA 不會因此產生 data xref——只看 xref 會得到「沒有人用這些字串」的假零。
DGROUP_EA = 0x2DED0


def disassembly(ea: int) -> str:
    rendered = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
    return rendered.split(" ;", 1)[0].rstrip()


def strings_matching(needles: tuple[bytes, ...]) -> list[dict]:
    out = []
    for item in idautils.Strings():
        raw = ida_bytes.get_bytes(item.ea, item.length) or b""
        low = raw.lower()
        if any(n in low for n in needles):
            refs = [f"0x{x:x}" for x in idautils.DataRefsTo(item.ea)]
            out.append({
                "ea": f"0x{item.ea:x}",
                "length": item.length,
                "text": raw.rstrip(b"\0").decode("ascii", "replace"),
                "bytes": raw.hex(),
                "data_refs_to": refs,
                "ref_functions": sorted({
                    f"0x{ida_funcs.get_func(x).start_ea:x}"
                    for x in idautils.DataRefsTo(item.ea)
                    if ida_funcs.get_func(x) is not None}),
            })
    return out


def immediate_users(offsets: set[int]) -> list[dict]:
    """掃全部指令，找出把這些 DGROUP 位移當立即值載入的地方。"""
    out = []
    for fn_ea in idautils.Functions():
        for ea in idautils.FuncItems(fn_ea):
            for op in range(2):
                if idc.get_operand_type(ea, op) != idc.o_imm:
                    continue
                value = idc.get_operand_value(ea, op)
                if value in offsets:
                    out.append({
                        "ea": f"0x{ea:x}",
                        "function": f"0x{fn_ea:x}",
                        "immediate": f"0x{value:x}",
                        "disassembly": disassembly(ea),
                    })
    return out


def function_slice(start: int, limit: int = 400) -> dict:
    fn = ida_funcs.get_func(start)
    if fn is None:
        return {"start_ea": f"0x{start:x}", "error": "no function"}
    items = list(idautils.FuncItems(start))
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{fn.end_ea:x}",
        "instruction_count": len(items),
        "callers": [f"0x{x:x}" for x in idautils.CodeRefsTo(start, False)],
        "instructions": [
            {"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}
            for ea in items[:limit]
        ],
    }


# 進一步的錨點：payload 載入器與磚塊緩衝區。
PAYLOAD_LOADER = 0x1A4FE      # sub_1A4FE(name, index, seg, off, extra)：解一個 payload
FILE_LOADER = 0x19D98         # sub_19D98(name, seg, off)：整個檔案讀進記憶體
TILE_BUFFER_SEG = 0x30C8A     # word_30C8A：三個磚圖 payload 落在這個 segment
# 想看清楚的函式：地圖區段載入器與用 ES 讀磚塊緩衝區的繪製路徑。
DUMP_FUNCTIONS = (0x11302, 0x11703, 0x11722, 0x1103F)
MAP_BUFFER = 0x2FBC4   # byte_2FBC4：sub_11A93 把 town%d 的 payload 直接讀進這個 DGROUP 陣列
# sub_115D1 這個繪製迴圈用到的全域：鏡頭座標、四個夾限值與格子索引暫存。
LOOP_GLOBALS = {
    "camera_x_word_2FB50": 0x2FB50,
    "camera_y_word_2FB56": 0x2FB56,
    "clamp_word_2DF66": 0x2DF66,
    "clamp_word_2DF68": 0x2DF68,
    "clamp_word_2DF6A": 0x2DF6A,
    "clamp_word_2DF6C": 0x2DF6C,
    "cell_index_word_3AC5A": 0x3AC5A,
    "screen_cache_2A48": 0x2A48,
    "sprite_source_word_39824": 0x39824,
    "sprite_base_word_2DF48": 0x2DF48,
    "sprite_frame_word_2DF24": 0x2DF24,
    "overlay_flags_6AE4": 0x6AE4,
    "door_table_unk_32300": 0x32300,
    "payload20_unk_3B26E": 0x3B26E,
    "building_index_word_2FBC2": 0x2FBC2,
}
SPRINTF = 0x2D3FA             # sub_2D3FA：組檔名用的 sprintf
# 磚塊間距是 128 bytes ＝ 8 個 paragraph。real mode 常見寫法是把索引乘 8 加到 segment，
# 而不是乘 128 加到 offset——所以掃「乘 128」掃不到，要改掃乘 8／乘 16 的位移鏈。
SHIFT_KINDS = {3: "x8-paragraph", 4: "x16"}


def call_sites(target: int, window: int = 14) -> list[dict]:
    """列出所有呼叫點，並附上呼叫前 window 條指令——引數是 push 上去的。"""
    out = []
    for ea in idautils.CodeRefsTo(target, False):
        fn = ida_funcs.get_func(ea)
        if fn is None:
            continue
        items = list(idautils.FuncItems(fn.start_ea))
        try:
            idx = items.index(ea)
        except ValueError:
            continue
        start = max(0, idx - window)
        out.append({
            "call_ea": f"0x{ea:x}",
            "function": f"0x{fn.start_ea:x}",
            "context": [{"ea": f"0x{e:x}", "disassembly": disassembly(e)}
                        for e in items[start:idx + 1]],
        })
    return out


def global_users(addr: int) -> list[dict]:
    out = []
    for ea in idautils.DataRefsTo(addr):
        fn = ida_funcs.get_func(ea)
        out.append({
            "ea": f"0x{ea:x}",
            "function": f"0x{fn.start_ea:x}" if fn else None,
            "disassembly": disassembly(ea),
        })
    return out


def format_strings() -> list[dict]:
    """把每個 sprintf 呼叫點的格式字串立即值解析成實際文字。

    先前把 `mov ax, 1874h` 的目標猜成 `twn` 那組字串，是因為只看了鄰近的字串清單。
    這裡直接讀 DGROUP+位移的位元組，不再用猜的。
    """
    out = []
    for ea in idautils.CodeRefsTo(SPRINTF, False):
        fn = ida_funcs.get_func(ea)
        if fn is None:
            continue
        items = list(idautils.FuncItems(fn.start_ea))
        idx = items.index(ea) if ea in items else -1
        if idx < 0:
            continue
        for e in items[max(0, idx - 10):idx]:
            if idc.print_insn_mnem(e) != "mov" or idc.get_operand_type(e, 1) != idc.o_imm:
                continue
            value = idc.get_operand_value(e, 1)
            if not 0x1000 <= value <= 0x8000:
                continue
            raw = ida_bytes.get_bytes(DGROUP_EA + value, 32) or b""
            text = raw.split(b"\0", 1)[0]
            if not text or any(b < 0x20 or b > 0x7E for b in text):
                continue
            out.append({
                "call_ea": f"0x{ea:x}", "function": f"0x{fn.start_ea:x}",
                "load_ea": f"0x{e:x}", "dgroup_offset": f"0x{value:x}",
                "text": text.decode("ascii"),
            })
    return out


def shift_chains() -> list[dict]:
    """找出連續 `shl reg,1` 形成的乘 8／乘 16，以及 `mov cl,n` ＋ `shl reg,cl`。"""
    out = []
    for fn_ea in idautils.Functions():
        items = list(idautils.FuncItems(fn_ea))
        run_reg, run_len, run_start = None, 0, None
        for ea in items:
            mnem = idc.print_insn_mnem(ea)
            if mnem in ("shl", "sal") and idc.get_operand_type(ea, 1) == idc.o_imm \
                    and idc.get_operand_value(ea, 1) == 1:
                reg = idc.print_operand(ea, 0)
                if reg == run_reg:
                    run_len += 1
                else:
                    run_reg, run_len, run_start = reg, 1, ea
                if run_len in SHIFT_KINDS:
                    out.append({"function": f"0x{fn_ea:x}", "ea": f"0x{run_start:x}",
                                "register": reg, "shift": run_len,
                                "kind": SHIFT_KINDS[run_len]})
            else:
                run_reg, run_len, run_start = None, 0, None
    return out


def tile_stride_arithmetic() -> list[dict]:
    """找出「把索引乘上 128」的地方——磚塊間距是 128 bytes，繪製迴圈一定會算這個。"""
    out = []
    for fn_ea in idautils.Functions():
        for ea in idautils.FuncItems(fn_ea):
            mnem = idc.print_insn_mnem(ea)
            if mnem not in ("shl", "sal", "imul", "mov"):
                continue
            text = disassembly(ea)
            hit = None
            if mnem in ("shl", "sal") and idc.get_operand_type(ea, 1) == idc.o_imm \
                    and idc.get_operand_value(ea, 1) == 7:
                hit = "shift-7"
            elif mnem == "imul" and idc.get_operand_type(ea, 1) == idc.o_imm \
                    and idc.get_operand_value(ea, 1) == 128:
                hit = "imul-128"
            elif mnem == "mov" and idc.get_operand_type(ea, 1) == idc.o_imm \
                    and idc.get_operand_value(ea, 1) == 7 \
                    and idc.print_operand(ea, 0) in ("cl", "cx"):
                hit = "mov-cl-7"
            if hit:
                out.append({"ea": f"0x{ea:x}", "function": f"0x{fn_ea:x}",
                            "kind": hit, "disassembly": text})
    return out


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    town_strings = strings_matching((b"town", b"twn", b"region"))
    for s in town_strings:
        s["dgroup_offset"] = f"0x{int(s['ea'], 16) - DGROUP_EA:x}"
    offsets = {int(s["ea"], 16) - DGROUP_EA for s in town_strings}
    users = immediate_users(offsets)

    seen = set()
    functions = []
    for u in users:
        if u["function"] in seen:
            continue
        seen.add(u["function"])
        functions.append(function_slice(int(u["function"], 16)))

    receipt = {
        "schema": "kol-ida-koltown-town-map-explore-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：字串、xref 與參照函式的反組譯；尚未鎖定任何斷言。",
        "town_strings": town_strings,
        "immediate_users": users,
        "payload_loader_calls": call_sites(PAYLOAD_LOADER),
        "file_loader_calls": call_sites(FILE_LOADER),
        "tile_buffer_users": global_users(TILE_BUFFER_SEG),
        "map_buffer_users": global_users(MAP_BUFFER),
        "loop_globals": {name: global_users(addr)
                         for name, addr in LOOP_GLOBALS.items()},
        "tile_blit_calls": call_sites(0x1CAC2, 18),
        "tile_stride_arithmetic": tile_stride_arithmetic(),
        "format_strings": format_strings(),
        "shift_chains": shift_chains(),
        "dumped_functions": [function_slice(ea, 1400) for ea in DUMP_FUNCTIONS],
        "referencing_functions": functions,
        "function_total": len(list(idautils.Functions())),
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
