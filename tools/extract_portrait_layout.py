#!/usr/bin/env python3
"""把角色立繪的合成版面從 KOLTOWN 的資料段抽成 typed data（RE-126）。

資料頁的物品那一頁與隊伍那一頁都由 `sub_18F3A` 把好幾張圖疊起來畫：底下是體型
（`PERPICS`），上面依序是腿、身、頭（`LEGSPICS`／`TORPICS`／`HEADPICS`），
再上去是盾、武器與手（`SHLDPICS`／`WEAPPICS`／`HANDPICS`）。每一層畫在哪、多大，
全部來自 DGROUP 0x168…0x68C 的一串表。

那一串表**首尾相接、中間沒有空隙**——0x168 的 6 筆身體位置接 0x17A 的 54×3 尺寸、
接 0x2BE 的 54×3 位置、接 0x402 的 28 筆武器尺寸…一路到 0x680 的六個位移。
筆數與抽出來的 PNG 組數也一一對上（`PERPICS` 6 組、頭身腿各 54 組、`SHLDPICS`
4 組、`WEAPPICS` 28 組、`HANDPICS` 12 組）。這兩件事一起，就是版面讀對了的證據。

  tools/run_python_tool.sh tools/extract_portrait_layout.py \
      --exe workplace/analysis/koltown/KOLTOWN-unpacked.EXE \
      --out assets/original-derived-data/portrait-layout.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib

# 反組譯用的是 IDA 線性位址；這個檔案是重建出來的 unpacked MZ，載入基底 0x10000，
# header 佔 0x2400 個 byte，所以 DGROUP 的 offset 換算成檔案位置就是加這個數。
EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
DGROUP_EA = 0x2DED0
LOAD_BASE = 0x10000
HEADER_BYTES = 0x2400
FILE_BASE = DGROUP_EA - LOAD_BASE + HEADER_BYTES

# 體型組合：`[+0x1A] × 4 + [+0x24]`（體型檔次與種族），共 6 種。
BUILDS = 6
# 每個體型組合有九套外觀（`+0x10E + 16×層` 的低 nibble），三層各一張圖。
STYLES = 9
LAYERS = 3          # 0 = 頭、1 = 身、2 = 腿（畫的順序是 2 → 0）
WEAPONS = 28        # WEAPPICS 的組數
SHIELDS = 4         # SHLDPICS 的組數
HANDS = 12          # HANDPICS 的組數
WEAPON_KINDS = 8    # `+0x3D` 除以 0x20 得到的武器種類

TABLES = {
    "body": 0x168,          # 3 byte × 6：x1、y1、x2（y2 固定 0x9A）
    "layer_size": 0x17A,    # 6 byte × 54：三層各 (寬, 高)
    "layer_pos": 0x2BE,     # 6 byte × 54：三層各 (x, y)
    "weapon_size": 0x402,   # 2 byte × 28
    "shield_size": 0x43A,   # 2 byte × 4
    "shield_pos": 0x442,    # 8 byte × 6：每個體型組合四種盾各 (x, y)
    "weapon_pos": 0x472,    # 0x38 byte × 6：每個體型組合 28 把武器各 (x, y)
    "hand_size": 0x5C2,     # 2 byte × 12
    "hand_pos": 0x5DA,      # 2 byte × 12
    "weapon_base": 0x5F2,   # byte × 8：武器種類 → WEAPPICS 的起始 payload 索引
    "names": 0x5FA,         # 5 byte × 6：head、tor、legs、weap、shld、hand
    "party_layer": 0x66A,   # byte × 6：隊伍頁由後往前畫的層次
    "mode": 0x670,          # 5 byte × 3：pics（EGA）、pcga、herc
    "party_x": 0x680,       # word × 6：隊伍頁第 n 個畫的人往右挪多少
}
# 表分成幾串。第一串 0x168…0x618 是立繪自己的版面，一路首尾相接；中間隔著兩張
# 與立繪無關的表（0x618 的十六個 word 與 0x636 的十二個月份名），0x66A 起才是隊伍頁
# 那三張——其中 0x67F 是一個填充 byte，把後面那張 word 表推到偶數位址。
RUNS = ((0x168, 0x618), (0x66A, 0x67F), (0x680, 0x68C))


def cstr(raw: bytes) -> str:
    return raw.split(b"\x00", 1)[0].decode("ascii")


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--exe", required=True, help="重建出來的 KOLTOWN-unpacked.EXE")
    ap.add_argument("--out", required=True, help="輸出的 JSON")
    args = ap.parse_args()

    exe = pathlib.Path(args.exe)
    raw = exe.read_bytes()
    digest = hashlib.sha256(raw).hexdigest()
    if digest != EXPECTED_SHA256:
        raise SystemExit(f"來源不是預期的那一份：SHA-256 {digest}")

    def dg(off: int, n: int) -> bytes:
        lo = FILE_BASE + off
        block = raw[lo:lo + n]
        if len(block) != n:
            raise SystemExit(f"DGROUP 0x{off:04x} 取不到 {n} 個 byte")
        return block

    # 先驗一次：這一串表要首尾相接，而且模式後綴要真的是那三個字。
    order = sorted(TABLES.items(), key=lambda kv: kv[1])
    spans = {
        "body": 3 * BUILDS,
        "layer_size": 6 * BUILDS * STYLES,
        "layer_pos": 6 * BUILDS * STYLES,
        "weapon_size": 2 * WEAPONS,
        "shield_size": 2 * SHIELDS,
        "shield_pos": 8 * BUILDS,
        "weapon_pos": 2 * WEAPONS * BUILDS,
        "hand_size": 2 * HANDS,
        "hand_pos": 2 * HANDS,
        "weapon_base": WEAPON_KINDS,
        "names": 5 * 6,
        "party_layer": BUILDS,
        "mode": 5 * 3,
        "party_x": 2 * BUILDS,
    }
    for (name, off), (_, nxt) in zip(order, order[1:] + [("", -1)]):
        end = off + spans[name]
        want = next((hi for lo, hi in RUNS if lo <= off < hi), None)
        if want is None:
            raise SystemExit(f"{name} 在 0x{off:04x}，不在任何一串表裡")
        expect = nxt if nxt >= 0 and nxt < want else want
        if end != expect:
            raise SystemExit(f"{name} 從 0x{off:04x} 到 0x{end:04x}，"
                             f"但下一張表在 0x{expect:04x}——版面讀錯了")
    modes = [cstr(dg(TABLES["mode"] + 5 * i, 5)) for i in range(3)]
    if modes != ["pics", "pcga", "herc"]:
        raise SystemExit(f"模式後綴是 {modes}，不是預期的三個")

    names = [cstr(dg(TABLES["names"] + 5 * i, 5)) for i in range(6)]

    body = []
    for v in range(BUILDS):
        x1, y1, x2 = dg(TABLES["body"] + 3 * v, 3)
        body.append({"x": x1, "y": y1, "x2": x2, "y2": 0x9A})

    layers = []
    for v in range(BUILDS * STYLES):
        pos = dg(TABLES["layer_pos"] + 6 * v, 6)
        size = dg(TABLES["layer_size"] + 6 * v, 6)
        layers.append([{"x": pos[2 * k], "y": pos[2 * k + 1],
                        "w": size[2 * k], "h": size[2 * k + 1]}
                       for k in range(LAYERS)])

    weapon_size = dg(TABLES["weapon_size"], 2 * WEAPONS)
    weapon_pos = dg(TABLES["weapon_pos"], 2 * WEAPONS * BUILDS)
    weapons = [[{"x": weapon_pos[0x38 * v + 2 * i], "y": weapon_pos[0x38 * v + 2 * i + 1],
                 "w": weapon_size[2 * i], "h": weapon_size[2 * i + 1]}
                for i in range(WEAPONS)] for v in range(BUILDS)]

    shield_size = dg(TABLES["shield_size"], 2 * SHIELDS)
    shield_pos = dg(TABLES["shield_pos"], 8 * BUILDS)
    shields = [[{"x": shield_pos[8 * v + 2 * i], "y": shield_pos[8 * v + 2 * i + 1],
                 "w": shield_size[2 * i], "h": shield_size[2 * i + 1]}
                for i in range(SHIELDS)] for v in range(BUILDS)]

    hand_size = dg(TABLES["hand_size"], 2 * HANDS)
    hand_pos = dg(TABLES["hand_pos"], 2 * HANDS)
    hands = [{"x": hand_pos[2 * i], "y": hand_pos[2 * i + 1],
              "w": hand_size[2 * i], "h": hand_size[2 * i + 1]}
             for i in range(HANDS)]

    party_x = dg(TABLES["party_x"], 2 * BUILDS)
    doc = {
        "schema": "kol-portrait-layout-v1",
        "source": exe.name,
        "source_sha256": digest,
        "dgroup_ea": f"0x{DGROUP_EA:x}",
        "tables": {k: f"0x{v:04x}" for k, v in TABLES.items()},
        "archive_names": names,
        "mode_suffix": modes,
        "body": body,
        "layers": layers,
        "weapons": weapons,
        "shields": shields,
        "hands": hands,
        "weapon_base": list(dg(TABLES["weapon_base"], WEAPON_KINDS)),
        "party_layer": list(dg(TABLES["party_layer"], BUILDS)),
        "party_x": [int.from_bytes(party_x[2 * i:2 * i + 2], "little", signed=True)
                    for i in range(BUILDS)],
    }
    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(doc, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")
    print(f"寫出 {out}：{BUILDS} 個體型、{len(layers)} 套外觀、"
          f"{WEAPONS} 把武器、{SHIELDS} 面盾、{HANDS} 隻手")
    return 0


raise SystemExit(main())
