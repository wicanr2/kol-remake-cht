#!/usr/bin/env python3
"""正常 DOS 玩家路徑的 KOLBOOK 角色建立 UI oracle 擷取器。

在容器內執行：唯讀掛載固定原版 ZIP，於可寫暫存區重建遊戲目錄，用 Xvfb 執行
DOSBox-X，並以 XTEST 送出腳本化按鍵。按鍵不是盲打時序，而是「等畫面靜止到指定
秒數 → 擷取該畫面 → 送下一鍵」，因此原版磁碟解碼時間的變動不會讓序列走偏。

輸出：去重畫格 PNG、畫格時間軸、每個 step 的檢查點，以及
CHARDATA／CHARPICS／SHLDPICS 的前後 SHA-256 收據。

設計限制：
- 原版 ZIP 唯讀，遊戲目錄是每次重建的副本；SHA-256 不符即失敗關閉。
- 只送鍵盤事件與讀畫面，不注入座標、不使用除錯入口、不跳過任何選單層。
- 畫格時間戳是主機牆鐘，逐次不同；可重生的比對對象是檢查點畫面的 PNG SHA-256。
  屬性擲骰畫面由 BIOS tick 播種，逐次不同是預期行為，不得當成擷取失敗。
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import pathlib
import signal
import struct
import subprocess
import sys
import time
import zipfile
import zlib

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from x11_key_inject import X11Client  # noqa: E402

EXPECTED_ZIP_SHA256 = "2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46"
CHARACTER_FILES = ("CHARDATA", "CHARPICS", "SHLDPICS")

DOSBOX_CONF = """[sdl]
autolock=false
output=surface
windowposition=0,0
fullscreen=false
[dosbox]
machine=ega
memsize=4
captures={captures}
[cpu]
core=normal
cputype=386
cycles={cycles}
[render]
aspect=false
scaler=none
[autoexec]
mount c {gamedir}
c:
kol.exe
exit
"""


# --------------------------------------------------------------------------
# 情境
# --------------------------------------------------------------------------
# step 欄位：
#   label   檢查點名稱（進入受控狀態、送出按鍵之前的畫面）
#   key     要送出的 keysym 名稱；None 代表只擷取不送鍵
#   settle  畫面需連續多少秒沒有變化才算靜止
#   timeout 等待靜止的上限秒數
#   delay   改為固定等待秒數，用於本來就持續動畫、不會靜止的畫面（例如 TITLE 標題序列）

def step(label: str, key: str | None, settle: float = 2.0, timeout: float = 120.0,
         delay: float | None = None) -> dict:
    return {"label": label, "key": key, "settle": settle, "timeout": timeout, "delay": delay}


# 由 KOL.EXE 正常啟動走到「New Character」名字輸入層的共同前置。
# 依 Apple II 參考卡的開始步驟：標題序列按任意鍵離開 → 書本 → 任意鍵翻到目錄頁 →
# 按項目編號 → 角色選項選單以反白條加 Return 選擇。TITLE 標題序列本身是循環動畫，
# 不會靜止，因此第一步用固定等待。走完 PRELUDE 後停在 New Character 的名字輸入層。
PRELUDE = [
    step("TITLE 標題序列", "space", delay=8.0),
    step("書本封面", "space", settle=2.5),
    step("目錄頁", "5", settle=2.5),
    step("角色選項選單（反白 LIST CHARACTERS）", "Down", settle=2.5),
    step("角色選項選單（反白 CREATE CHARACTER）", "Return", settle=2.0),
]

NAME_LETTERS = "abcdefghijklmnopq"     # 17 個，超過手冊宣稱的 16 上限


def _name_length_steps() -> list[dict]:
    steps: list[dict] = []
    for i, ch in enumerate(NAME_LETTERS):
        typed = NAME_LETTERS[:i] or "（空）"
        steps.append(step(f"名字欄內容 {typed}", ch, settle=1.2, timeout=60.0))
    steps += [
        step(f"送出全部 {len(NAME_LETTERS)} 個字母之後的名字欄", "BackSpace", settle=1.5),
        step("第一次 Backspace 之後", "BackSpace", settle=1.5),
        step("第二次 Backspace 之後", "Escape", settle=1.5),
        step("名字層取消之後", None, settle=3.0),
    ]
    return steps


# 走到角色清單的前置：目錄頁按 5 之後，角色選項選單預設就反白 LIST CHARACTERS，
# 因此不必先按 Down。
LIST_PRELUDE = [
    step("TITLE 標題序列", "space", delay=8.0),
    step("書本封面", "space", settle=2.5),
    step("目錄頁", "5", settle=2.5),
]

# 走到角色選項選單的前置。目錄頁按 5 之後有一段約六秒的翻頁動畫；用固定等待跨過去，
# 避免在動畫途中誤判畫面已靜止而把後續按鍵送進動畫裡（第一次試跑就是這樣整串失步）。
# 走完停在選單，反白條在第 0 項 LIST CHARACTERS。
MENU_PRELUDE = [
    step("TITLE 標題序列", "space", delay=8.0),
    step("書本封面", "space", delay=9.0),
    step("目錄頁", "5", delay=10.0),
]

# 角色選項選單的六個項目；index 是由 LIST CHARACTERS 起算要按幾次 Down。
MENU_ITEMS = ["LIST CHARACTERS", "CREATE CHARACTER", "DELETE CHARACTER",
              "EDIT PICTURE", "EDIT SHIELD", "HARBOR MASTER"]


def menu_steps(index: int) -> list[dict]:
    """由選單第 0 項走到第 index 項並按 Return。"""
    out = []
    for i in range(index):
        out.append(step(f"角色選項選單（反白 {MENU_ITEMS[i]}）", "Down", settle=2.0))
    out.append(step(f"角色選項選單（反白 {MENU_ITEMS[index]}）", "Return", settle=3.5))
    return out


DELETE_PRELUDE = list(MENU_PRELUDE) + menu_steps(2)

# 由 KOL.EXE 走到城鎮（半途之家旅店）的前置：目錄頁按 6 選 Play the game，
# 在共用的角色選擇清單選第一名角色當領隊。載入與開場對話都不會靜止，整段用固定等待。
# 第 14 名角色 ddduhh 住在城 1 的巨魔剋星客棧，身上 3000 金幣——買賣要看得到價錢。
SHOP_PRELUDE = [
    step("TITLE 標題序列", "space", delay=8.0),
    step("書本封面", "space", delay=9.0),
    step("目錄頁", "6", delay=12.0),
] + [
    step(f"角色清單往下第 {i + 1} 名", "Down", settle=1.2) for i in range(13)
] + [
    step("選第 14 名角色", "Return", delay=25.0),
    step("旅店老闆對話（第一段）", "space", delay=15.0),
]

TOWN_PRELUDE = [
    step("TITLE 標題序列", "space", delay=8.0),
    step("書本封面", "space", delay=9.0),
    step("目錄頁", "6", delay=12.0),
    step("角色選擇清單（反白第 1 名）", "Return", delay=25.0),
    step("旅店老闆對話（第一段）", "space", delay=15.0),
]

SCENARIOS: dict[str, dict] = {
    "list-characters": {
        "description": "從角色選項選單選 LIST CHARACTERS，擷取角色清單畫面，"
                       "再取消回到選單；全程不建立、不刪除、不保存任何角色。",
        "steps": list(LIST_PRELUDE) + [
            step("角色選項選單（反白 LIST CHARACTERS）", "Return", settle=3.0),
            step("角色清單", "Escape", settle=2.5),
            step("角色清單取消之後", None, settle=3.0),
        ],
    },
    "edit-picture-probe": {
        "description": "從角色選項選單進入 EDIT PICTURE，擷取進入後的畫面，"
                       "再逐次取消退回選單。全程不修改任何角色。",
        "steps": list(MENU_PRELUDE) + menu_steps(3) + [
            step("EDIT PICTURE 進入後", "Escape", settle=3.0),
            step("第一次取消之後", "Escape", settle=3.0),
            step("第二次取消之後", None, settle=3.0),
        ],
    },
    "edit-shield-probe": {
        "description": "從角色選項選單進入 EDIT SHIELD，擷取進入後的畫面，"
                       "再逐次取消退回選單。全程不修改任何角色。",
        "steps": list(MENU_PRELUDE) + menu_steps(4) + [
            step("EDIT SHIELD 進入後", "Escape", settle=3.0),
            step("第一次取消之後", "Escape", settle=3.0),
            step("第二次取消之後", None, settle=3.0),
        ],
    },
    "harbor-master-probe": {
        "description": "從角色選項選單進入 HARBOR MASTER，擷取進入後的畫面，"
                       "再逐次取消退回選單。全程不修改任何角色。",
        "steps": list(MENU_PRELUDE) + menu_steps(5) + [
            step("HARBOR MASTER 進入後", "Escape", settle=3.0),
            step("第一次取消之後", "Escape", settle=3.0),
            step("第二次取消之後", None, settle=3.0),
        ],
    },
    "edit-picture-enter": {
        "description": "從角色選項選單進入 EDIT PICTURE，在角色選擇清單選第一名角色，"
                       "擷取功能本身的畫面，再逐次取消退回；保存角色檔以檢查是否寫入。",
        "steps": list(MENU_PRELUDE) + menu_steps(3) + [
            step("角色選擇清單（反白第 1 名）", "Return", settle=4.0),
            step("EDIT PICTURE 畫面", "Escape", settle=3.5),
            step("第一次取消之後", "Escape", settle=3.0),
            step("第二次取消之後", "Escape", settle=3.0),
            step("第三次取消之後", None, settle=3.0),
        ],
    },
    "edit-shield-enter": {
        "description": "從角色選項選單進入 EDIT SHIELD，在角色選擇清單選第一名角色，"
                       "擷取功能本身的畫面，再逐次取消退回；保存角色檔以檢查是否寫入。",
        "steps": list(MENU_PRELUDE) + menu_steps(4) + [
            step("角色選擇清單（反白第 1 名）", "Return", settle=4.0),
            step("EDIT SHIELD 畫面", "Escape", settle=3.5),
            step("第一次取消之後", "Escape", settle=3.0),
            step("第二次取消之後", "Escape", settle=3.0),
            step("第三次取消之後", None, settle=3.0),
        ],
    },
    "harbor-master-enter": {
        "description": "從角色選項選單進入 HARBOR MASTER，在角色選擇清單選第一名角色，"
                       "擷取功能本身的畫面，再逐次取消退回；保存角色檔以檢查是否寫入。",
        "steps": list(MENU_PRELUDE) + menu_steps(5) + [
            step("角色選擇清單（反白第 1 名）", "Return", settle=4.0),
            step("HARBOR MASTER 畫面", "Escape", settle=3.5),
            step("第一次取消之後", "Escape", settle=3.0),
            step("第二次取消之後", "Escape", settle=3.0),
            step("第三次取消之後", None, settle=3.0),
        ],
    },
    "town-icon-bar": {
        "description": "在旅店的圖示列上左右移動再按 Return，測試圖示列是否吃方向鍵。"
                       "收據結論：方向鍵沒有作用，Return 直接執行目前白框的 DOOR。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列", "Right", settle=1.5),
            step("向右一次", "Right", settle=1.5),
            step("向右兩次", "Right", settle=1.5),
            step("向右三次", "Left", settle=1.5),
            step("向左一次", "Return", settle=3.0),
            step("按下 Return 之後", "Escape", settle=3.0),
            step("按 ESC 之後", None, settle=3.0),
        ],
    },
    "town-icon-move": {
        "description": "用參考卡指定的 < 與 > 在旅店圖示列上移動白框，逐格擷取六個圖示，"
                       "最後停在 COMPANION 但不按 Return。純觀察。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "greater", settle=1.5),
            step("白框第 2 格", "greater", settle=1.5),
            step("白框第 3 格", "greater", settle=1.5),
            step("白框第 4 格", "greater", settle=1.5),
            step("白框第 5 格", "greater", settle=1.5),
            step("白框第 6 格", "greater", settle=1.5),
            step("再往右一次（測是否繞回）", "less", settle=1.5),
            step("往左一次", None, settle=2.5),
        ],
    },
    "town-recruit-companion": {
        "description": "在旅店用 > 走到 COMPANION 圖示，從同伴清單選第一名角色加入隊伍，"
                       "並保存角色檔以檢查招募是否寫入磁碟。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "greater", settle=1.5),
            step("白框第 2 格", "greater", settle=1.5),
            step("白框第 3 格", "greater", settle=1.5),
            step("白框第 4 格", "greater", settle=1.5),
            step("白框第 5 格", "greater", settle=1.5),
            step("白框第 6 格（COMPANION）", "Return", settle=4.0),
            step("是否搜尋硬碟角色", "Return", settle=4.0),
            step("同伴清單（反白第 1 名）", "Down", settle=2.0),
            step("反白第 2 名", "Up", settle=2.0),
            step("回到第 1 名", "Return", settle=4.0),
            step("選定第一名同伴之後", None, settle=4.0),
        ],
    },
    "town-shop": {
        "description": "選第 14 名角色（ddduhh，3000 金幣，住在巨魔剋星客棧）進城 1，"
                       "走到魯德曼甲冑鋪按 TRADE 看商品清單。純觀察，不買也不保存。",
        "steps": list(SHOP_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "Return", settle=3.0),
            step("走出旅店", None, settle=2.0),
        ] + [
            step("往甲冑鋪第 1 步", "Left", settle=1.0),
            step("往甲冑鋪第 2 步", "Left", settle=1.0),
            step("往甲冑鋪第 3 步", "Left", settle=1.0),
            step("往甲冑鋪第 4 步", "Left", settle=1.0),
            step("往甲冑鋪第 5 步", "Left", settle=1.0),
            step("往甲冑鋪第 6 步", "Left", settle=1.0),
            step("往甲冑鋪第 7 步", "Left", settle=1.0),
            step("往甲冑鋪第 8 步", "Left", settle=1.0),
            step("往甲冑鋪第 9 步", "Left", settle=1.0),
            step("往甲冑鋪第 10 步", "Left", settle=1.0),
            step("往甲冑鋪第 11 步", "Left", settle=1.0),
            step("往甲冑鋪第 12 步", "Left", settle=1.0),
            step("往甲冑鋪第 13 步", "Down", settle=1.0),
            step("往甲冑鋪第 14 步", "Down", settle=1.0),
            step("往甲冑鋪第 15 步", "Left", settle=1.0),
            step("往甲冑鋪第 16 步", "Left", settle=1.0),
            step("往甲冑鋪第 17 步", "Down", settle=1.0),
            step("往甲冑鋪第 18 步", "Left", settle=1.0),
            step("往甲冑鋪第 19 步", "Left", settle=1.0),
            step("往甲冑鋪第 20 步", "Down", settle=1.0),
            step("往甲冑鋪第 21 步", "Left", settle=1.0),
            step("往甲冑鋪第 22 步", "Left", settle=1.0),
            step("往甲冑鋪第 23 步", "Down", settle=1.0),
            step("往甲冑鋪第 24 步", "Left", settle=1.0),
            step("往甲冑鋪第 25 步", "Left", settle=1.0),
            step("往甲冑鋪第 26 步", "Left", settle=1.0),
            step("往甲冑鋪第 27 步", "Left", settle=1.0),
            step("往甲冑鋪第 28 步", "Left", settle=1.0),
            step("往甲冑鋪第 29 步", "Left", settle=1.0),
            step("往甲冑鋪第 30 步", "Left", settle=1.0),
            step("往甲冑鋪第 31 步", "Left", settle=1.0),
            step("往甲冑鋪第 32 步", "Down", settle=1.0),
            step("往甲冑鋪第 33 步", "Down", settle=1.0),
            step("往甲冑鋪第 34 步", "Down", settle=1.0),
            step("往甲冑鋪第 35 步", "Down", settle=1.0),
            step("往甲冑鋪第 36 步", "Down", settle=1.0),
            step("往甲冑鋪第 37 步", "Left", settle=1.0),
        ] + [
            step("站到甲冑鋪門口", "Return", settle=3.5),
            step("進門敘述", "space", settle=2.5),
            step("讀完敘述", "greater", settle=1.2),
            step("白框第 2 格", "greater", settle=1.2),
            step("白框第 3 格", "greater", settle=1.2),
            step("白框第 4 格", "greater", settle=1.2),
            step("白框第 5 格（TRADE）", "Return", settle=4.0),
            step("商品清單", "Down", settle=2.0),
            step("清單往下一列", "Up", settle=2.0),
            step("回到第一列", "Return", settle=3.5),
            step("商品明細與「要買嗎」", "Return", settle=4.0),
            step("買下之後", "Escape", settle=3.0),
            step("離開清單", "greater", settle=1.5),
            step("白框第 2 格（MIRROR）", "Return", settle=4.0),
            step("角色資料頁：金幣與背包", "Escape", settle=3.0),
            step("離開資料頁", None, settle=3.0),
        ],
    },
    "town-train": {
        "description": "選第 14 名角色（ddduhh，住城 1 的巨魔剋星客棧）走到布瑞特堡壘，"
                       "按 AWARD 看訓練那一串：學費、技能清單、攻擊還是防禦、"
                       "要幾點技能。他身上沒有冒險點數，最後會被擋下來。純觀察，不保存。",
        "steps": list(SHOP_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "Return", settle=3.0),
            step("走出旅店", None, settle=2.0),
        ] + [
            step("往堡壘第 1 步", "Left", settle=1.0),
            step("往堡壘第 2 步", "Left", settle=1.0),
            step("往堡壘第 3 步", "Left", settle=1.0),
            step("往堡壘第 4 步", "Down", settle=1.0),
            step("往堡壘第 5 步", "Down", settle=1.0),
            step("往堡壘第 6 步", "Down", settle=1.0),
            step("往堡壘第 7 步", "Down", settle=1.0),
            step("往堡壘第 8 步", "Down", settle=1.0),
            step("往堡壘第 9 步", "Down", settle=1.0),
            step("往堡壘第 10 步", "Down", settle=1.0),
            step("往堡壘第 11 步", "Down", settle=1.0),
            step("往堡壘第 12 步", "Right", settle=1.0),
            step("往堡壘第 13 步", "Down", settle=1.0),
            step("往堡壘第 14 步", "Down", settle=1.0),
            step("往堡壘第 15 步", "Down", settle=1.0),
            step("往堡壘第 16 步", "Down", settle=1.0),
            step("往堡壘第 17 步", "Right", settle=1.0),
            step("往堡壘第 18 步", "Right", settle=1.0),
            step("往堡壘第 19 步", "Right", settle=1.0),
            step("往堡壘第 20 步", "Right", settle=1.0),
            step("往堡壘第 21 步", "Right", settle=1.0),
            step("往堡壘第 22 步", "Up", settle=1.0),
        ] + [
            step("站到堡壘門口", "Return", settle=3.5),
            step("進門敘述", "space", settle=2.5),
            step("讀完敘述", "greater", settle=1.2),
            step("白框第 2 格", "greater", settle=1.2),
            step("白框第 3 格", "greater", settle=1.2),
            step("白框第 4 格", "greater", settle=1.2),
            step("白框第 5 格（AWARD）", "Return", settle=4.0),
            step("學費那一問", "Return", settle=3.5),
            step("付完學費：要練哪一門", "space", settle=3.0),
            step("技能清單", "Return", settle=3.5),
            step("攻擊還是防禦", "Return", settle=3.5),
            step("要幾點技能", "1", settle=3.5),
            step("按下 1 之後", "Return", settle=3.0),
            step("再按 Return", "space", settle=3.0),
            step("讀完", None, settle=3.0),
        ],
    },
    "town-adjust": {
        "description": "同一條路走到魯德曼甲冑鋪，買下毛皮帽並穿上，再按 TRAIN "
                       "看甲冑調整那一支說什麼。純觀察，不保存。",
        "steps": list(SHOP_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "Return", settle=3.0),
            step("走出旅店", None, settle=2.0),
        ] + [
            step("往甲冑鋪第 1 步", "Left", settle=1.0),
            step("往甲冑鋪第 2 步", "Left", settle=1.0),
            step("往甲冑鋪第 3 步", "Left", settle=1.0),
            step("往甲冑鋪第 4 步", "Left", settle=1.0),
            step("往甲冑鋪第 5 步", "Left", settle=1.0),
            step("往甲冑鋪第 6 步", "Left", settle=1.0),
            step("往甲冑鋪第 7 步", "Left", settle=1.0),
            step("往甲冑鋪第 8 步", "Left", settle=1.0),
            step("往甲冑鋪第 9 步", "Left", settle=1.0),
            step("往甲冑鋪第 10 步", "Left", settle=1.0),
            step("往甲冑鋪第 11 步", "Left", settle=1.0),
            step("往甲冑鋪第 12 步", "Left", settle=1.0),
            step("往甲冑鋪第 13 步", "Down", settle=1.0),
            step("往甲冑鋪第 14 步", "Down", settle=1.0),
            step("往甲冑鋪第 15 步", "Left", settle=1.0),
            step("往甲冑鋪第 16 步", "Left", settle=1.0),
            step("往甲冑鋪第 17 步", "Down", settle=1.0),
            step("往甲冑鋪第 18 步", "Left", settle=1.0),
            step("往甲冑鋪第 19 步", "Left", settle=1.0),
            step("往甲冑鋪第 20 步", "Down", settle=1.0),
            step("往甲冑鋪第 21 步", "Left", settle=1.0),
            step("往甲冑鋪第 22 步", "Left", settle=1.0),
            step("往甲冑鋪第 23 步", "Down", settle=1.0),
            step("往甲冑鋪第 24 步", "Left", settle=1.0),
            step("往甲冑鋪第 25 步", "Left", settle=1.0),
            step("往甲冑鋪第 26 步", "Left", settle=1.0),
            step("往甲冑鋪第 27 步", "Left", settle=1.0),
            step("往甲冑鋪第 28 步", "Left", settle=1.0),
            step("往甲冑鋪第 29 步", "Left", settle=1.0),
            step("往甲冑鋪第 30 步", "Left", settle=1.0),
            step("往甲冑鋪第 31 步", "Left", settle=1.0),
            step("往甲冑鋪第 32 步", "Down", settle=1.0),
            step("往甲冑鋪第 33 步", "Down", settle=1.0),
            step("往甲冑鋪第 34 步", "Down", settle=1.0),
            step("往甲冑鋪第 35 步", "Down", settle=1.0),
            step("往甲冑鋪第 36 步", "Down", settle=1.0),
            step("往甲冑鋪第 37 步", "Left", settle=1.0),
        ] + [
            step("站到甲冑鋪門口", "Return", settle=3.5),
            step("進門敘述", "space", settle=2.5),
            step("讀完敘述", "greater", settle=1.2),
            step("白框第 2 格", "greater", settle=1.2),
            step("白框第 3 格", "greater", settle=1.2),
            step("白框第 4 格", "greater", settle=1.2),
            step("白框第 5 格（TRADE）", "Return", settle=4.0),
            step("商品清單", "Return", settle=3.5),
            step("商品明細", "Return", settle=4.0),
            step("買下並穿上", "space", settle=3.0),
            step("回到清單", "greater", settle=1.5),
            step("白框到 U-TURN", "Return", settle=3.5),
            step("回到建築", "greater", settle=1.2),
            step("白框第 2 格", "greater", settle=1.2),
            step("白框第 3 格", "greater", settle=1.2),
            step("白框第 4 格", "greater", settle=1.2),
            step("白框第 5 格", "greater", settle=1.2),
            step("白框第 6 格", "greater", settle=1.2),
            step("白框第 7 格（TRAIN）", "Return", settle=4.0),
            step("按下 TRAIN 之後", "space", settle=3.0),
            step("讀完", None, settle=3.0),
        ],
    },
    "town-mirror": {
        "description": "在旅店把白框移到 MIRROR 按下去，看角色資料頁的版面與"
                       "它吃哪些鍵。純觀察，不修改也不保存角色。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "greater", settle=1.5),
            step("白框第 2 格（MIRROR）", "Return", settle=4.0),
            step("按下 MIRROR 之後", "space", settle=3.0),
            step("按空白鍵之後", "Down", settle=3.0),
            step("按向下之後", "Right", settle=3.0),
            step("按向右之後", "Return", settle=3.0),
            step("按 Return 之後", "Escape", settle=3.0),
            step("按 Escape 之後", None, settle=4.0),
        ],
    },
    "town-rest-save": {
        "description": "在旅店把白框移到 REST 按下去，觀察「誰要留宿」的選單、"
                       "金幣不足會怎樣、以及選全隊之後那一問。會保存角色檔，"
                       "因此收據會記下檔案有變動——遊戲副本在容器 tmpfs，原版 ZIP 唯讀。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "greater", settle=1.5),
            step("白框第 2 格", "greater", settle=1.5),
            step("白框第 3 格", "greater", settle=1.5),
            step("白框第 4 格", "greater", settle=1.5),
            step("白框第 5 格（REST）", "Return", settle=4.0),
            step("按下 REST 之後", "Return", settle=4.0),
        ] + [
            step(f"再按 Return 第 {i} 次", "Return", settle=3.0) for i in range(2, 6)
        ] + [
            step("連按之後的畫面", "Escape", settle=3.0),
            step("按 Escape 之後", None, settle=4.0),
        ],
    },
    "town-rest-two": {
        "description": "先招募一名同伴再按 REST，看住宿費是不是按人頭算。"
                       "只看價目那一句，不回答，直接離開。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "greater", settle=1.5),
            step("白框第 2 格", "greater", settle=1.5),
            step("白框第 3 格", "greater", settle=1.5),
            step("白框第 4 格", "greater", settle=1.5),
            step("白框第 5 格", "greater", settle=1.5),
            step("白框第 6 格（COMPANION）", "Return", settle=4.0),
            step("是否搜尋硬碟角色", "Return", settle=4.0),
            step("同伴清單（反白第 1 名）", "Return", settle=4.0),
            step("加入第一名同伴之後", "greater", settle=1.5),
            step("白框第 2 格", "greater", settle=1.5),
            step("白框第 3 格", "greater", settle=1.5),
            step("白框第 4 格", "greater", settle=1.5),
            step("白框第 5 格（REST）", "Return", settle=4.0),
            step("兩人隊伍按下 REST：整隊嗎？", "Return", settle=4.0),
            step("回答整隊之後（觀察價目）", "Escape", settle=4.0),
            step("按 Escape 之後", None, settle=3.0),
        ],
    },
    "town-recruit-list": {
        "description": "在旅店打開同伴清單，一路往下按到底再往上按到底，"
                       "觀察清單超過一頁時怎麼捲、兩端會不會繞回；"
                       "最後按第一格圖示離開。純觀察，不加入任何同伴、不保存。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "greater", settle=1.5),
            step("白框第 2 格", "greater", settle=1.5),
            step("白框第 3 格", "greater", settle=1.5),
            step("白框第 4 格", "greater", settle=1.5),
            step("白框第 5 格", "greater", settle=1.5),
            step("白框第 6 格（COMPANION）", "Return", settle=4.0),
            step("是否搜尋硬碟角色", "Return", settle=4.0),
        ] + [
            step(f"往下第 {i} 名", "Down", settle=2.0) for i in range(1, 15)
        ] + [
            step("按到底之後", "Up", settle=2.0),
        ] + [
            step(f"往上第 {i} 名", "Up", settle=2.0) for i in range(2, 16)
        ] + [
            step("按回頂端之後", "greater", settle=2.0),
            step("白框移到第 2 格", "Return", settle=4.0),
            step("按下第 2 格圖示之後（觀察）", None, settle=4.0),
        ],
    },
    "town-enter-building": {
        "description": "由旅店按 DOOR 走到戶外地圖，先離開門口再走回來，觀察站在門口時"
                       "地圖視窗是不是換成建築物的圖；接著按 DOOR 進去，再出來走到"
                       "另一個門口，看第二棟建築的圖。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "Return", settle=3.5),
            step("戶外地圖（隊伍在旅店門口）", "Left", settle=2.5),
            step("往左一步，離開門口", "Right", settle=3.0),
            step("走回旅店門口", None, settle=3.0),
            step("門口狀態（觀察）", "Return", settle=3.5),
            step("按 DOOR 進旅店", None, settle=3.0),
            step("旅店畫面（觀察）", "Return", settle=3.5),
            step("按 DOOR 回到戶外", "Left", settle=2.5),
            step("往左一步", "Left", settle=2.5),
            step("往左兩步", "Left", settle=2.5),
            step("往左三步", "Down", settle=3.0),
            step("往下一步（另一個門口）", None, settle=3.0),
            step("第二個門口狀態（觀察）", "Return", settle=3.5),
            step("進入第二棟建築", None, settle=3.0),
            step("第二棟建築內部（觀察）", None, settle=3.0),
        ],
    },
    "town-map-edge": {
        "description": "由旅店按 DOOR 走到戶外地圖，往右兩步再一路往北走到地圖上緣，"
                       "看鏡頭被夾限時畫面變成什麼；接著按 U-TURN 回地圖。"
                       "路徑由 TOWN7 的格子陣列與不可通行位元圖算出，全程不進任何建築。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "Return", settle=3.5),
            step("戶外地圖（隊伍在旅店門口）", "Right", settle=2.5),
            step("往右一步", "Right", settle=2.5),
            step("往右兩步", "Up", settle=2.5),
        ] + [
            step(f"往北第 {i} 步", "Up", settle=2.5) for i in range(1, 8)
        ] + [
            step("走到地圖上緣（鏡頭已夾限）", "Up", settle=3.5),
            step("再往北一步：城門提示（觀察）", "greater", settle=2.5),
            step("白框移到 U-TURN", "Return", settle=3.5),
            step("按 U-TURN 之後回到地圖", None, settle=3.0),
        ],
    },
    "town-leave-to-wilderness": {
        "description": "走到城鎮地圖上緣跳出城門提示之後按 DOOR 離開城鎮，"
                       "觀察野外地圖的版面；接著在野外走幾步並按幾次圖示鍵。"
                       "純觀察，不進任何建築、不保存。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "Return", settle=3.5),
            step("戶外地圖（隊伍在旅店門口）", "Right", settle=2.5),
            step("往右一步", "Right", settle=2.5),
            step("往右兩步", "Up", settle=2.5),
        ] + [
            step(f"往北第 {i} 步", "Up", settle=2.5) for i in range(1, 8)
        ] + [
            step("走到地圖上緣（鏡頭已夾限）", "Up", settle=3.5),
            step("城門提示（白框在 DOOR）", "Return", settle=6.0),
            step("離開城鎮之後的第一張畫面（觀察）", "Up", settle=3.0),
            step("往北一步", "Up", settle=3.0),
            step("往北兩步", "Left", settle=3.0),
            step("往西一步", "Left", settle=3.0),
            step("往西兩步", "greater", settle=2.5),
            step("白框往右一格", "greater", settle=2.5),
            step("白框往右兩格", "Return", settle=4.0),
            step("按下該格圖示之後（觀察）", None, settle=4.0),
        ],
    },
    "world-road-follow": {
        "description": "離開城鎮走到野外的道路上，連按 ROAD 讓隊伍自動循路，"
                       "再按一次 SIGNPOST 換另一條路。純觀察，不進建築、不保存。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "Return", settle=3.5),
            step("戶外地圖（隊伍在旅店門口）", "Right", settle=2.5),
            step("往右一步", "Right", settle=2.5),
            step("往右兩步", "Up", settle=2.5),
        ] + [
            step(f"往北第 {i} 步", "Up", settle=2.5) for i in range(1, 8)
        ] + [
            step("走到地圖上緣（鏡頭已夾限）", "Up", settle=3.5),
            step("城門提示（白框在 DOOR）", "Return", settle=6.0),
            step("野外地圖的第一張畫面", "Up", settle=3.0),
        ] + [
            step(f"野外往北第 {i} 步", "Up", settle=3.0) for i in range(2, 13)
        ] + [
            step("站到路上（圖示列出現 ROAD 與 SIGNPOST）", "Return", settle=3.5),
        ] + [
            step(f"循路第 {i} 步", "Return", settle=3.5) for i in range(1, 7)
        ] + [
            step("循路第 7 步", "greater", settle=2.5),
            step("白框移到 SIGNPOST", "Return", settle=3.5),
            step("按下 SIGNPOST 之後（觀察）", None, settle=4.0),
        ],
    },
    "world-road-clock": {
        "description": "走到路上之後只按 ROAD，量「循路一步扣多少時刻」："
                       "野外先往北十二步（PLAINS，每步 2 × (8 + 20) = 56），"
                       "站到路上再連按 ROAD 六十次，看時段標籤在第幾次翻成 MID MORN、又在第幾次翻成 NOON。"
                       "純觀察，不保存。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "Return", settle=3.5),
            step("戶外地圖（隊伍在旅店門口）", "Right", settle=2.5),
            step("往右一步", "Right", settle=2.5),
            step("往右兩步", "Up", settle=2.5),
        ] + [
            step(f"往北第 {i} 步", "Up", settle=2.5) for i in range(1, 8)
        ] + [
            step("走到地圖上緣（鏡頭已夾限）", "Up", settle=3.5),
            step("城門提示（白框在 DOOR）", "Return", settle=6.0),
            step("野外地圖的第一張畫面", "Up", settle=3.0),
        ] + [
            step(f"野外往北第 {i} 步", "Up", settle=3.0) for i in range(2, 13)
        ] + [
            step(f"循路第 {i} 步", "Return", settle=2.0, timeout=12.0)
            for i in range(1, 61)
        ] + [
            step("觀察結束", None, settle=3.0, timeout=12.0),
        ],
    },
    "world-clock": {
        "description": "離開城鎮之後在野外原地左右走動：每一次方向鍵都會推進時刻，"
                       "但次格座標在同一格內來回，不會換格、不會換地形、也不會踏回"
                       "城鎮入口。用來量「時段標籤在第幾步變」。純觀察，不保存。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "Return", settle=3.5),
            step("戶外地圖（隊伍在旅店門口）", "Right", settle=2.5),
            step("往右一步", "Right", settle=2.5),
            step("往右兩步", "Up", settle=2.5),
        ] + [
            step(f"往北第 {i} 步", "Up", settle=2.5) for i in range(1, 8)
        ] + [
            step("走到地圖上緣（鏡頭已夾限）", "Up", settle=3.5),
            step("城門提示（白框在 DOOR）", "Return", settle=6.0),
        ] + [
            step(f"野外第 {i} 步", "Right" if i % 2 else "Left", settle=2.0)
            for i in range(1, 25)
        ] + [
            step("量完（觀察）", None, settle=3.0),
        ],
    },
    "world-camp": {
        "description": "離開城鎮之後在野外原地左右走動到時刻用完：每步扣 56 點，"
                       "第 76 步時刻變成負的，原版會切到 CAMPING! 那一張並等一個鍵。"
                       "同一格內來回，不換格所以不會踩到遭遇。純觀察，不保存。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "Return", settle=3.5),
            step("戶外地圖（隊伍在旅店門口）", "Right", settle=2.5),
            step("往右一步", "Right", settle=2.5),
            step("往右兩步", "Up", settle=2.5),
        ] + [
            step(f"往北第 {i} 步", "Up", settle=2.5) for i in range(1, 8)
        ] + [
            step("走到地圖上緣（鏡頭已夾限）", "Up", settle=3.5),
            step("城門提示（白框在 DOOR）", "Return", settle=6.0),
        ] + [
            # 天候一旦換成有雲層動畫的那幾組，畫面就不會靜止；逾時放短，
            # 讓每一步至多等 8 秒就送下一個鍵。
            step(f"野外第 {i} 步", "Right" if i % 2 else "Left",
                 settle=1.6, timeout=8.0)
            for i in range(1, 77)
        ] + [
            step("紮營畫面（等一個鍵）", "space", settle=4.0, timeout=20.0),
            step("收掉紮營之後", "Right", settle=3.0, timeout=20.0),
            step("再走一步（觀察）", None, settle=3.0, timeout=20.0),
        ],
    },
    "world-ambush": {
        "description": "在野外原地左右走動，等時段變換擲出伏擊（一天四次機會）；"
                       "每五步插一次空白鍵把訊息推過去，最後逐鍵觀察戰鬥畫面。"
                       "純觀察，不保存。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "Return", settle=3.5),
            step("戶外地圖（隊伍在旅店門口）", "Right", settle=2.5),
            step("往右一步", "Right", settle=2.5),
            step("往右兩步", "Up", settle=2.5),
        ] + [
            step(f"往北第 {i} 步", "Up", settle=2.5) for i in range(1, 8)
        ] + [
            step("走到地圖上緣（鏡頭已夾限）", "Up", settle=3.5),
            step("城門提示（白框在 DOOR）", "Return", settle=6.0),
        ] + [
            step("野外第 1 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 2 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 3 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 4 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 5 步", "Right", settle=1.6, timeout=8.0),
            step("第 5 步後推一次訊息", "space", settle=2.0, timeout=10.0),
            step("野外第 6 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 7 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 8 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 9 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 10 步", "Left", settle=1.6, timeout=8.0),
            step("第 10 步後推一次訊息", "space", settle=2.0, timeout=10.0),
            step("野外第 11 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 12 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 13 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 14 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 15 步", "Right", settle=1.6, timeout=8.0),
            step("第 15 步後推一次訊息", "space", settle=2.0, timeout=10.0),
            step("野外第 16 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 17 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 18 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 19 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 20 步", "Left", settle=1.6, timeout=8.0),
            step("第 20 步後推一次訊息", "space", settle=2.0, timeout=10.0),
            step("野外第 21 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 22 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 23 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 24 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 25 步", "Right", settle=1.6, timeout=8.0),
            step("第 25 步後推一次訊息", "space", settle=2.0, timeout=10.0),
            step("野外第 26 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 27 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 28 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 29 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 30 步", "Left", settle=1.6, timeout=8.0),
            step("第 30 步後推一次訊息", "space", settle=2.0, timeout=10.0),
            step("野外第 31 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 32 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 33 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 34 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 35 步", "Right", settle=1.6, timeout=8.0),
            step("第 35 步後推一次訊息", "space", settle=2.0, timeout=10.0),
            step("野外第 36 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 37 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 38 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 39 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 40 步", "Left", settle=1.6, timeout=8.0),
            step("第 40 步後推一次訊息", "space", settle=2.0, timeout=10.0),
            step("野外第 41 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 42 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 43 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 44 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 45 步", "Right", settle=1.6, timeout=8.0),
            step("第 45 步後推一次訊息", "space", settle=2.0, timeout=10.0),
            step("野外第 46 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 47 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 48 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 49 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 50 步", "Left", settle=1.6, timeout=8.0),
            step("第 50 步後推一次訊息", "space", settle=2.0, timeout=10.0),
            step("野外第 51 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 52 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 53 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 54 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 55 步", "Right", settle=1.6, timeout=8.0),
            step("第 55 步後推一次訊息", "space", settle=2.0, timeout=10.0),
            step("野外第 56 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 57 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 58 步", "Left", settle=1.6, timeout=8.0),
            step("野外第 59 步", "Right", settle=1.6, timeout=8.0),
            step("野外第 60 步", "Left", settle=1.6, timeout=8.0),
            step("第 60 步後推一次訊息", "space", settle=2.0, timeout=10.0),
            step("戰鬥第 1 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("戰鬥第 2 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("戰鬥第 3 鍵（Return）", "Return", settle=2.5, timeout=15.0),
            step("戰鬥第 4 鍵（greater）", "greater", settle=2.5, timeout=15.0),
            step("戰鬥第 5 鍵（greater）", "greater", settle=2.5, timeout=15.0),
            step("戰鬥第 6 鍵（Return）", "Return", settle=2.5, timeout=15.0),
            step("戰鬥第 7 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("戰鬥第 8 鍵（Up）", "Up", settle=2.5, timeout=15.0),
            step("戰鬥第 9 鍵（Right）", "Right", settle=2.5, timeout=15.0),
            step("戰鬥第 10 鍵（Return）", "Return", settle=2.5, timeout=15.0),
            step("戰鬥第 11 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("戰鬥第 12 鍵（1）", "1", settle=2.5, timeout=15.0),
            step("戰鬥第 13 鍵（Return）", "Return", settle=2.5, timeout=15.0),
            step("戰鬥第 14 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("戰鬥第 15 鍵（greater）", "greater", settle=2.5, timeout=15.0),
            step("戰鬥第 16 鍵（Return）", "Return", settle=2.5, timeout=15.0),
            step("戰鬥第 17 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("戰鬥第 18 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("觀察結束", None, settle=3.0, timeout=15.0),
        ],
    },
    "world-quest-site": {
        "description": "選第 2 名角色（Poontang，任務旗標 char[0x1E3] = 4，對應第 5 個"
                       "特殊地點 (48,21)），出城之後往東八格、往北兩格走到那一格，"
                       "觸發任務怪物並進戰鬥；沿途每六步插一次空白鍵推訊息。純觀察，不保存。",
        "steps": [
            step("TITLE 標題序列", "space", delay=8.0),
            step("書本封面", "space", delay=9.0),
            step("目錄頁", "6", delay=12.0),
            step("角色清單往下第 1 名", "Down", settle=1.2),
            step("選第 2 名角色（Poontang）", "Return", delay=25.0),
            step("旅店老闆對話（第一段）", "space", delay=15.0),
            step("旅店圖示列（白框在 DOOR）", "Return", settle=3.5),
            step("戶外地圖（隊伍在旅店門口）", "Right", settle=2.5),
            step("往右一步", "Right", settle=2.5),
            step("往右兩步", "Up", settle=2.5),
        ] + [
            step(f"往北第 {i} 步", "Up", settle=2.5) for i in range(1, 8)
        ] + [
            step("走到地圖上緣（鏡頭已夾限）", "Up", settle=3.5),
            step("城門提示（白框在 DOOR）", "Return", settle=6.0),
        ] + [
            step("往東第 1 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 2 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 3 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 4 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 5 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 6 步", "Right", settle=1.4, timeout=8.0),
            step("推一次訊息（東 6）", "space", settle=1.6, timeout=10.0),
            step("往東第 7 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 8 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 9 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 10 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 11 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 12 步", "Right", settle=1.4, timeout=8.0),
            step("推一次訊息（東 12）", "space", settle=1.6, timeout=10.0),
            step("往東第 13 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 14 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 15 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 16 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 17 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 18 步", "Right", settle=1.4, timeout=8.0),
            step("推一次訊息（東 18）", "space", settle=1.6, timeout=10.0),
            step("往東第 19 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 20 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 21 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 22 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 23 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 24 步", "Right", settle=1.4, timeout=8.0),
            step("推一次訊息（東 24）", "space", settle=1.6, timeout=10.0),
            step("往東第 25 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 26 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 27 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 28 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 29 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 30 步", "Right", settle=1.4, timeout=8.0),
            step("推一次訊息（東 30）", "space", settle=1.6, timeout=10.0),
            step("往東第 31 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 32 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 33 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 34 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 35 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 36 步", "Right", settle=1.4, timeout=8.0),
            step("推一次訊息（東 36）", "space", settle=1.6, timeout=10.0),
            step("往東第 37 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 38 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 39 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 40 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 41 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 42 步", "Right", settle=1.4, timeout=8.0),
            step("推一次訊息（東 42）", "space", settle=1.6, timeout=10.0),
            step("往東第 43 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 44 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 45 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 46 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 47 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 48 步", "Right", settle=1.4, timeout=8.0),
            step("推一次訊息（東 48）", "space", settle=1.6, timeout=10.0),
            step("往東第 49 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 50 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 51 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 52 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 53 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 54 步", "Right", settle=1.4, timeout=8.0),
            step("推一次訊息（東 54）", "space", settle=1.6, timeout=10.0),
            step("往東第 55 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 56 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 57 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 58 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 59 步", "Right", settle=1.4, timeout=8.0),
            step("往東第 60 步", "Right", settle=1.4, timeout=8.0),
            step("推一次訊息（東 60）", "space", settle=1.6, timeout=10.0),
            step("往東第 61 步", "Right", settle=1.4, timeout=8.0),
            step("往北第 1 步", "Up", settle=1.4, timeout=8.0),
            step("往北第 2 步", "Up", settle=1.4, timeout=8.0),
            step("往北第 3 步", "Up", settle=1.4, timeout=8.0),
            step("往北第 4 步", "Up", settle=1.4, timeout=8.0),
            step("往北第 5 步", "Up", settle=1.4, timeout=8.0),
            step("推一次訊息（北 5）", "space", settle=1.6, timeout=10.0),
            step("往北第 6 步", "Up", settle=1.4, timeout=8.0),
            step("往北第 7 步", "Up", settle=1.4, timeout=8.0),
            step("往北第 8 步", "Up", settle=1.4, timeout=8.0),
            step("往北第 9 步", "Up", settle=1.4, timeout=8.0),
            step("往北第 10 步", "Up", settle=1.4, timeout=8.0),
            step("往北第 11 步", "Up", settle=1.4, timeout=8.0),
            step("推一次訊息（北 11）", "space", settle=1.6, timeout=10.0),
            step("往北第 12 步", "Up", settle=1.4, timeout=8.0),
            step("戰鬥第 1 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("戰鬥第 2 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("戰鬥第 3 鍵（Return）", "Return", settle=2.5, timeout=15.0),
            step("戰鬥第 4 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("戰鬥第 5 鍵（greater）", "greater", settle=2.5, timeout=15.0),
            step("戰鬥第 6 鍵（greater）", "greater", settle=2.5, timeout=15.0),
            step("戰鬥第 7 鍵（Return）", "Return", settle=2.5, timeout=15.0),
            step("戰鬥第 8 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("戰鬥第 9 鍵（Up）", "Up", settle=2.5, timeout=15.0),
            step("戰鬥第 10 鍵（Right）", "Right", settle=2.5, timeout=15.0),
            step("戰鬥第 11 鍵（Return）", "Return", settle=2.5, timeout=15.0),
            step("戰鬥第 12 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("戰鬥第 13 鍵（1）", "1", settle=2.5, timeout=15.0),
            step("戰鬥第 14 鍵（Return）", "Return", settle=2.5, timeout=15.0),
            step("戰鬥第 15 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("戰鬥第 16 鍵（greater）", "greater", settle=2.5, timeout=15.0),
            step("戰鬥第 17 鍵（Return）", "Return", settle=2.5, timeout=15.0),
            step("戰鬥第 18 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("戰鬥第 19 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("戰鬥第 20 鍵（space）", "space", settle=2.5, timeout=15.0),
            step("觀察結束", None, settle=3.0, timeout=15.0),
        ],
    },
    "world-enter-town": {
        "description": "離開城鎮走到野外之後，往東走五步跨進隔壁那一格，再往西一步"
                       "踏回城鎮入口那一格：畫面應換成該地點的外觀圖，圖示列剩 "
                       "DOOR 與 U-TURN；按 DOOR 回到城鎮。純觀察，不進建築、不保存。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "Return", settle=3.5),
            step("戶外地圖（隊伍在旅店門口）", "Right", settle=2.5),
            step("往右一步", "Right", settle=2.5),
            step("往右兩步", "Up", settle=2.5),
        ] + [
            step(f"往北第 {i} 步", "Up", settle=2.5) for i in range(1, 8)
        ] + [
            step("走到地圖上緣（鏡頭已夾限）", "Up", settle=3.5),
            step("城門提示（白框在 DOOR）", "Return", settle=6.0),
            step("野外地圖的第一張畫面", "Right", settle=3.0),
        ] + [
            step(f"野外往東第 {i} 步", "Right", settle=3.0) for i in range(2, 6)
        ] + [
            step("往東第五步（次格繞回，世界座標往東一格）", "Left", settle=3.0),
            step("往西一步：踏回城鎮入口那一格", "Return", settle=5.0),
            step("按下 DOOR 之後（應回到城鎮地圖）", None, settle=4.0),
        ],
    },
    "town-listen": {
        "description": "在旅店按 SPACEBAR 讓圖示列出現，再把白框移到 LISTEN 連按幾次，"
                       "觀察開場白與傳聞；最後移回 DOOR 離開。純觀察，不保存。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店的 SPACEBAR 圖示列", "space", settle=3.5),
            step("圖示列（DOOR／MIRROR／LISTEN／MOUTH／REST／COMPANION）", "greater", settle=2.0),
            step("白框到 MIRROR", "greater", settle=2.0),
            step("白框到 LISTEN", "Return", settle=4.0),
            step("第一次 LISTEN", "Return", settle=4.0),
            step("第二次 LISTEN", "Return", settle=4.0),
            step("第三次 LISTEN", "Return", settle=4.0),
            step("第四次 LISTEN", "Return", settle=4.0),
            step("第五次 LISTEN", "Return", settle=4.0),
            step("第六次 LISTEN", "Return", settle=4.0),
            step("第七次 LISTEN（應該已經講完）", None, settle=4.0),
        ],
    },
    "town-mouth": {
        "description": "在旅店按 SPACEBAR 讓圖示列出現，白框移到 MOUTH 按 Return 開問話輸入框，"
                       "逐鍵打進一個關鍵字並退格一次，觀察輸入框的起點、字往右長的方式與"
                       "筆尖那條橫槓；最後按 Return 送出。純觀察，不保存。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店的 SPACEBAR 圖示列", "space", settle=3.5),
            step("圖示列（白框在 DOOR）", "greater", settle=2.0),
            step("白框到 MIRROR", "greater", settle=2.0),
            step("白框到 LISTEN", "greater", settle=2.0),
            step("白框到 MOUTH（開問話輸入框）", "Return", settle=4.0),
            step("問話輸入框（還沒打字）", "e", settle=1.5),
            step("打了 e", "l", settle=1.5),
            step("打了 el", "v", settle=1.5),
            step("打了 elv", "e", settle=1.5),
            step("打了 elve", "BackSpace", settle=1.5),
            step("退格之後（elv）", "e", settle=1.5),
            step("補回 elve", "Return", settle=4.0),
            step("送出之後", "space", settle=3.0),
            step("讀完", None, settle=3.0),
        ],
    },
    "town-outdoor-move": {
        "description": "由旅店按 DOOR 走到戶外地圖，再用方向鍵移動幾步並在圖示列上移動，"
                       "用來判斷地圖視窗是不是一張固定圖。純觀察，不進任何建築。",
        "steps": list(TOWN_PRELUDE) + [
            step("旅店圖示列（白框在 DOOR）", "Return", settle=3.5),
            step("戶外地圖", "Right", settle=2.0),
            step("往右一步", "Right", settle=2.0),
            step("往右兩步", "Down", settle=2.0),
            step("往下一步", "Down", settle=2.0),
            step("往下兩步", "Up", settle=2.0),
            step("往上一步", "greater", settle=2.0),
            step("圖示列往右一格", "greater", settle=2.0),
            step("圖示列往右兩格", None, settle=3.0),
        ],
    },
    "play-the-game": {
        "description": "從目錄頁選 6（Play the game），在角色選擇清單選第一名角色當領隊，"
                       "觀察遊戲實際開始之後的畫面。純觀察，不改任何角色。",
        # 載入與開場動畫都不會靜止，因此整段用固定等待。
        "steps": [
            step("TITLE 標題序列", "space", delay=8.0),
            step("書本封面", "space", delay=9.0),
            step("目錄頁", "6", delay=12.0),
            step("角色選擇清單（反白第 1 名）", "Return", delay=25.0),
            step("開始遊戲之後（第一張）", "space", delay=15.0),
            step("按鍵之後（第二張）", "space", delay=15.0),
            step("按鍵之後（第三張）", "space", delay=15.0),
            step("按鍵之後（第四張）", None, delay=15.0),
        ],
    },
    "delete-character-probe": {
        "description": "從角色選項選單進入 DELETE CHARACTER，擷取刪除畫面之後直接取消，"
                       "不刪除任何角色。用來看清楚刪除流程的畫面與按鍵。",
        "steps": list(DELETE_PRELUDE) + [
            step("刪除角色畫面", "Escape", settle=3.0),
            step("取消之後", None, settle=3.0),
        ],
    },
    "delete-character": {
        "description": "實際刪除一名角色：進入 DELETE CHARACTER 後上下移動反白條，"
                       "選第二名角色按 Return 叫出確認視窗，在視窗內移動一次再送出，"
                       "並保存刪除前後的角色檔以便逐位元比對。",
        "steps": list(DELETE_PRELUDE) + [
            step("刪除清單（反白第 1 名）", "Down", settle=1.5),
            step("刪除清單（反白第 2 名）", "Down", settle=1.5),
            step("刪除清單（反白第 3 名）", "Up", settle=1.5),
            step("刪除清單（回到第 2 名）", "Return", settle=3.0),
            step("確認視窗（預設反白）", "Down", settle=2.0),
            step("確認視窗（反白第二項）", "Return", settle=4.0),
            step("送出確認之後", None, settle=4.0),
        ],
    },
    "delete-character-confirm": {
        "description": "在確認視窗回答 YES 實際刪除第二名角色，保存刪除前後的角色檔；"
                       "隨後再次進入刪除清單觀察刪除後的名冊，並在確認視窗回答 NO "
                       "驗證取消分支。",
        # 確認視窗不吃方向鍵與 Return（delete-character 收據：送 Down／Return 後畫面
        # 完全沒變），視窗上寫的是 YES/NO，因此用字母鍵回答。
        "steps": list(DELETE_PRELUDE) + [
            step("刪除清單（反白第 1 名）", "Down", settle=1.5),
            step("刪除清單（反白第 2 名）", "Return", settle=3.0),
            step("確認視窗（回答 YES）", "y", settle=4.0),
            step("刪除完成後的畫面", "Down", settle=2.5),
            step("選單（反白第二項）", "Down", settle=2.0),
            step("選單（反白第三項）", "Return", settle=3.5),
            step("刪除後的清單", "Return", settle=3.0),
            step("確認視窗（回答 NO）", "n", settle=4.0),
            step("回答 NO 之後", None, settle=4.0),
        ],
    },
    "accept-and-save-ranger": {
        "description": "同 accept-and-save，但職業層往下一項（RANGER）。用來把"
                       "「武器 selector → ITEM.LST 記錄索引」的對應再取兩個資料點，"
                       "避免只靠單一職業推廣。",
        "steps": list(PRELUDE) + [
            step("名字層", "r", settle=1.5),
            step("名字欄 r", "a", settle=1.2),
            step("名字欄 ra", "n", settle=1.2),
            step("名字欄 ran", "Return", settle=1.5),
            step("種族層（選 HUMAN）", "Return", settle=2.5),
            step("性別層（選 MALE）", "Return", settle=2.5),
            step("職業層（反白第一項）", "Down", settle=2.0),
            step("職業層（選第二項）", "Return", settle=3.0),
            step("屬性頁（反白 REROLL STATS）", "Down", settle=3.0),
            step("屬性頁（反白 ACCEPT STATS）", "Return", settle=3.0),
            step("圖像層（selector 0）", "Return", settle=3.0),
            step("接受圖像之後", None, settle=4.0),
        ],
    },
    "prelude": {
        "description": "正常玩家路徑走到 New Character 名字輸入層。",
        "steps": list(PRELUDE) + [step("新角色名字層", None, settle=3.0)],
    },
    "name-length": {
        "description": "在名字欄連續送 17 個可見 ASCII 字母，再逐次 Backspace，"
                       "量測此 DOS build 實際接受並顯示的字元數；最後取消不保存。",
        "steps": list(PRELUDE) + _name_length_steps(),
    },
    "blank-name": {
        "description": "名字欄不輸入任何字元直接按 Return，檢查是否前進到種族層；"
                       "隨後逐層取消，不保存任何角色。",
        "steps": list(PRELUDE) + [
            step("空白名字欄", "Return", settle=2.5),
            step("空名送出後的畫面", "Escape", settle=2.5),
            step("第一次取消之後", "Escape", settle=2.0),
            step("第二次取消之後", "Escape", settle=2.0),
            step("第三次取消之後", None, settle=3.0),
        ],
    },
    "escape-layers": {
        "description": "逐層驗證 ESC 的返回目的地：種族→名字、性別→種族、職業→性別、"
                       "屬性→職業，最後退出建立流程且不保存。名字層在取消後會清空，"
                       "因此回到名字層要重新輸入。",
        "steps": list(PRELUDE) + [
            step("名字層", "t", settle=1.5),
            step("名字欄 t", "e", settle=1.2),
            step("名字欄 te", "s", settle=1.2),
            step("名字欄 tes", "t", settle=1.2),
            step("名字欄 test", "Return", settle=1.5),
            step("種族層", "Escape", settle=2.5),
            step("種族層 ESC 之後", "t", settle=1.5),
            step("重新輸入 t", "e", settle=1.2),
            step("重新輸入 te", "s", settle=1.2),
            step("重新輸入 tes", "t", settle=1.2),
            step("重新輸入 test", "Return", settle=1.5),
            step("種族層（第二次）", "Return", settle=2.5),
            step("性別層", "Escape", settle=2.5),
            step("性別層 ESC 之後", "Return", settle=2.5),
            step("種族層（第三次）", "Return", settle=2.5),
            step("職業層", "Escape", settle=2.5),
            step("職業層 ESC 之後", "Return", settle=2.5),
            step("性別層（第三次）", "Return", settle=3.0),
            step("屬性頁", "Escape", settle=3.0),
            step("屬性頁 ESC 之後", "Escape", settle=2.5),
            step("再退一層", "Escape", settle=2.5),
            step("再退一層之二", "Escape", settle=2.5),
            step("再退一層之三", None, settle=3.0),
        ],
    },
    "accept-and-save": {
        "description": "完整走完建立流程：名字 zed、HUMAN、MALE、第一個職業、"
                       "ACCEPT STATS、接受第一張圖像，量測三個角色檔的實際寫入。",
        "steps": list(PRELUDE) + [
            step("名字層", "z", settle=1.5),
            step("名字欄 z", "e", settle=1.2),
            step("名字欄 ze", "d", settle=1.2),
            step("名字欄 zed", "Return", settle=1.5),
            step("種族層（選 HUMAN）", "Return", settle=2.5),
            step("性別層（選 MALE）", "Return", settle=2.5),
            step("職業層（選第一項）", "Return", settle=3.0),
            step("屬性頁（反白 REROLL STATS）", "Down", settle=3.0),
            step("屬性頁（反白 ACCEPT STATS）", "Return", settle=3.0),
            step("圖像層（selector 0）", "Return", settle=3.0),
            step("接受圖像之後", None, settle=4.0),
        ],
    },
    "picture-cycle": {
        "description": "在圖像層驗證 selector 的循環：先向左一次檢查由 0 回繞，"
                       "再連續向右走完整圈並回到起點；全程不接受圖像、不保存。",
        "steps": list(PRELUDE) + [
            step("名字層", "z", settle=1.5),
            step("名字欄 z", "e", settle=1.2),
            step("名字欄 ze", "d", settle=1.2),
            step("名字欄 zed", "Return", settle=1.5),
            step("種族層（選 HUMAN）", "Return", settle=2.5),
            step("性別層（選 MALE）", "Return", settle=2.5),
            step("職業層（選第一項）", "Return", settle=3.0),
            step("屬性頁（反白 REROLL STATS）", "Down", settle=3.0),
            step("屬性頁（反白 ACCEPT STATS）", "Return", settle=3.0),
            step("圖像層 selector 0", "Left", settle=1.0, timeout=60.0),
            step("向左一次之後（預期回繞到最後一張）", "Right", settle=1.0, timeout=60.0),
        ] + [
            step(f"向右第 {i} 次之前", "Right", settle=0.8, timeout=60.0)
            for i in range(1, 82)
        ] + [
            step("走完一圈之後", "Escape", settle=2.5),
            step("圖像層取消之後", None, settle=3.0),
        ],
    },
    "reroll": {
        "description": "走到屬性頁後連續兩次 REROLL STATS，比對屬性數值變化；"
                       "全程不接受屬性、不保存角色。",
        "steps": list(PRELUDE) + [
            step("名字層", "t", settle=1.5),
            step("名字欄 t", "e", settle=1.2),
            step("名字欄 te", "s", settle=1.2),
            step("名字欄 tes", "t", settle=1.2),
            step("名字欄 test", "Return", settle=1.5),
            step("種族層（選 HUMAN）", "Return", settle=2.5),
            step("性別層（選 MALE）", "Return", settle=2.5),
            step("職業層（選第一項）", "Return", settle=3.0),
            step("屬性頁初次擲骰（反白 REROLL STATS）", "Return", settle=3.0),
            step("第一次 REROLL 之後", "Return", settle=3.0),
            step("第二次 REROLL 之後", "Escape", settle=3.0),
            step("屬性頁取消之後", "Escape", settle=2.5),
            step("再取消一層", "Escape", settle=2.5),
            step("再取消一層之二", "Escape", settle=2.5),
            step("再取消一層之三", None, settle=3.0),
        ],
    },
}


# world-quest-execute 的前綴就是 world-quest-site 走到戰場那一段：一路走到第 5 個
# 任務地點、答應「Will you partake?」、按過滿版怪物圖進戰場。取到「戰鬥第 4 鍵」為止。
_quest_steps = SCENARIOS["world-quest-site"]["steps"]
SCENARIOS_QUEST_PREFIX = _quest_steps[:[s["label"] for s in _quest_steps]
                                      .index("戰鬥第 4 鍵（space）") + 1]

# 執行段的訊息畫面。進戰場之後只做一件事：讓 Poontang 選休息、答應
# 「Begin next round?」，再一路按鍵推完整個執行段的訊息。
#
# 為什麼不照 remake 的路徑走：原版的敵人**會動**（`sub_2374C` 的 AI），
# 每一手都印一句再等鍵，而 remake 的敵人那一手還沒實作（RE-081）。兩邊的
# 按鍵數與訊息內容因此對不起來。這一份要的是**訊息框的版面**——文字起點、
# 行高、斷行與逐格清底（RE-087）——那幾項與誰動的無關。
SCENARIOS["world-quest-execute"] = {
    "description": "同 world-quest-site 走到第 5 個任務地點並進戰場，接著讓"
                   "Poontang 選最後一格（休息）、答應「Begin next round?」，"
                   "再連按二十次空白鍵推完執行段的訊息。取的是訊息框的版面，"
                   "不是逐手的內容——原版敵人會動而 remake 還沒實作。純觀察，不保存。",
    "steps": SCENARIOS_QUEST_PREFIX + [
        step("指令列往右第 1 格", "greater", settle=2.0, timeout=15.0),
        step("指令列往右第 2 格", "greater", settle=2.0, timeout=15.0),
        step("指令列往右第 3 格", "greater", settle=2.0, timeout=15.0),
        step("指令列往右第 4 格", "greater", settle=2.0, timeout=15.0),
        step("指令列往右第 5 格", "greater", settle=2.0, timeout=15.0),
        step("指令列往右第 6 格（休息）", "greater", settle=2.0, timeout=15.0),
        step("選休息", "Return", settle=2.5, timeout=15.0),
        step("Begin next round? 答應", "Return", settle=2.5, timeout=15.0),
    ] + [
        step(f"推執行段訊息第 {i} 次", "space", settle=2.0, timeout=15.0)
        for i in range(1, 21)
    ] + [
        step("觀察結束", None, settle=3.0, timeout=15.0),
    ],
}



def sha256_file(path: pathlib.Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as fh:
        for block in iter(lambda: fh.read(1 << 20), b""):
            h.update(block)
    return h.hexdigest()


def extract_game(zip_path: pathlib.Path, dest: pathlib.Path) -> dict:
    digest = sha256_file(zip_path)
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit(f"ZIP SHA-256 不符：{digest}")
    dest.mkdir(parents=True, exist_ok=True)
    count = 0
    with zipfile.ZipFile(zip_path) as zf:
        for info in zf.infolist():
            if info.is_dir():
                continue
            (dest / info.filename.split("/")[-1]).write_bytes(zf.read(info))
            count += 1
    return {"zip_sha256": digest, "entry_count": count}


def character_hashes(gamedir: pathlib.Path) -> dict:
    return {name: {"bytes": (gamedir / name).stat().st_size,
                   "sha256": sha256_file(gamedir / name)}
            for name in CHARACTER_FILES}


# --------------------------------------------------------------------------
# XWD framebuffer → PNG
# --------------------------------------------------------------------------

def xwd_crop_rgb(data: bytes, rect: dict) -> bytes | None:
    """從 Xvfb 的 XWD framebuffer 取指定矩形的 RGB 像素。"""
    if len(data) < 100:
        return None
    f = struct.unpack(">25I", data[:100])
    hdr_size, version, depth = f[0], f[1], f[3]
    pm_w, pm_h = f[4], f[5]
    bpp, bytes_per_line, ncolors = f[11], f[12], f[19]
    if version != 7 or depth != 24 or bpp != 32:
        return None
    x0, y0, w, h = rect["x"], rect["y"], rect["width"], rect["height"]
    if x0 < 0 or y0 < 0 or x0 + w > pm_w or y0 + h > pm_h or w <= 0 or h <= 0:
        return None
    px_off = hdr_size + ncolors * 12
    rows = []
    for y in range(y0, y0 + h):
        base = px_off + y * bytes_per_line + x0 * 4
        line = data[base:base + w * 4]
        if len(line) < w * 4:
            return None
        out = bytearray(w * 3)
        for x in range(w):
            out[x * 3] = line[x * 4 + 2]
            out[x * 3 + 1] = line[x * 4 + 1]
            out[x * 3 + 2] = line[x * 4]
        rows.append(bytes(out))
    return b"".join(rows)


def write_png(path: pathlib.Path, w: int, h: int, rgb: bytes) -> str:
    raw = b"".join(b"\x00" + rgb[y * w * 3:(y + 1) * w * 3] for y in range(h))

    def chunk(tag: bytes, body: bytes) -> bytes:
        payload = tag + body
        return (struct.pack(">I", len(body)) + payload
                + struct.pack(">I", zlib.crc32(payload) & 0xFFFFFFFF))

    png = (b"\x89PNG\r\n\x1a\n"
           + chunk(b"IHDR", struct.pack(">IIBBBBB", w, h, 8, 2, 0, 0, 0))
           + chunk(b"IDAT", zlib.compress(raw, 9))
           + chunk(b"IEND", b""))
    path.write_bytes(png)
    return hashlib.sha256(png).hexdigest()


# --------------------------------------------------------------------------
# 擷取
# --------------------------------------------------------------------------

class Capture:
    def __init__(self, client: X11Client, window: int, fb_path: pathlib.Path,
                 framedir: pathlib.Path, start: float):
        self.client = client
        self.window = window
        self.fb_path = fb_path
        self.framedir = framedir
        self.start = start
        self.seen: dict[str, int] = {}
        self.timeline: list[dict] = []
        self.last_digest: str | None = None

    @staticmethod
    def diff_ratio(a: bytes, b: bytes) -> float:
        """兩張畫面的差異比例（不同的 bit 數／總 bit 數）；長度不同視為完全不同。

        這個比例只用來判斷畫面何時算靜止。跨輪比對用的是
        `tools/compare_character_ui_runs.py` 的 byte 比例，兩者不是同一個量。
        """
        if len(a) != len(b) or not a:
            return 1.0
        xor = int.from_bytes(a, "big") ^ int.from_bytes(b, "big")
        return xor.bit_count() / (len(a) * 8)

    def sample(self) -> tuple[str, int, bytes] | None:
        rect = self.client.get_geometry(self.window)
        try:
            data = self.fb_path.read_bytes()
        except OSError:
            return None
        rgb = xwd_crop_rgb(data, rect)
        if rgb is None:
            return None
        digest = hashlib.sha256(
            rgb + f"{rect['width']}x{rect['height']}".encode()).hexdigest()
        if digest != self.last_digest:
            self.last_digest = digest
            if digest in self.seen:
                index = self.seen[digest]
                self.timeline.append({"t": round(time.time() - self.start, 2),
                                      "frame": index, "repeat": True})
            else:
                index = len(self.seen)
                self.seen[digest] = index
                name = f"frame-{index:03d}.png"
                png_sha = write_png(self.framedir / name,
                                    rect["width"], rect["height"], rgb)
                self.timeline.append({
                    "t": round(time.time() - self.start, 2),
                    "frame": index,
                    "file": f"frames/{name}",
                    "size": [rect["width"], rect["height"]],
                    "png_sha256": png_sha,
                })
        return digest, self.seen[digest], rgb

    def wait_stable(self, settle: float, timeout: float, interval: float,
                    tolerance: float) -> dict:
        """等到畫面連續 settle 秒內的變化都小於 tolerance。

        名字欄的閃爍游標會讓嚴格逐 byte 比對永遠不穩定，因此以差異比例判斷；
        容差只影響「何時送下一鍵」，畫格本身仍以完整雜湊去重保存。
        """
        start_wait = time.time()
        deadline = start_wait + timeout
        last_change = start_wait
        ref = None
        current = None
        while time.time() < deadline:
            got = self.sample()
            if got is not None:
                current = got
                if ref is None or self.diff_ratio(ref, got[2]) > tolerance:
                    ref = got[2]
                    last_change = time.time()
                elif time.time() - last_change >= settle:
                    return {"stable": True, "frame": got[1],
                            "waited_s": round(time.time() - start_wait, 2)}
            time.sleep(interval)
        return {"stable": False, "frame": current[1] if current else None,
                "waited_s": round(timeout, 2)}


def resolve_steps(args) -> tuple[str, str, list[dict]]:
    if args.steps_json:
        spec = json.loads(pathlib.Path(args.steps_json).read_text())
        return (spec.get("name", "adhoc"), spec.get("description", ""),
                [step(s["label"], s.get("key"), s.get("settle", 2.0),
                      s.get("timeout", 120.0), s.get("delay"))
                 for s in spec["steps"]])
    if args.scenario not in SCENARIOS:
        raise SystemExit(f"未知情境：{args.scenario}")
    sc = SCENARIOS[args.scenario]
    return args.scenario, sc["description"], sc["steps"]


def run(args) -> int:
    work = pathlib.Path(args.work)
    gamedir = work / "game"
    fbdir = work / "fb"
    outdir = pathlib.Path(args.out)
    framedir = outdir / "frames"
    for d in (fbdir, framedir, work / "capture"):
        d.mkdir(parents=True, exist_ok=True)

    scenario_name, description, steps = resolve_steps(args)
    inventory = extract_game(pathlib.Path(args.zip), gamedir)
    pre = character_hashes(gamedir)

    conf_text = DOSBOX_CONF.format(captures=str(work / "capture"),
                                   gamedir=str(gamedir), cycles=args.cycles)
    conf_path = work / "dosbox-x.conf"
    conf_path.write_text(conf_text)

    ver = subprocess.run([args.dosbox, "-version"], capture_output=True, text=True)
    version_line = next((l for l in (ver.stdout + ver.stderr).splitlines() if l.strip()), "unknown")

    env = dict(os.environ, HOME=str(work))
    xvfb = subprocess.Popen(
        ["Xvfb", args.display, "-screen", "0", "800x600x24",
         "-fbdir", str(fbdir), "-nolisten", "tcp"],
        stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, env=env)
    fb_path = fbdir / "Xvfb_screen0"
    deadline = time.time() + 10
    while not fb_path.exists() and time.time() < deadline:
        time.sleep(0.1)
    if not fb_path.exists():
        xvfb.terminate()
        raise SystemExit("Xvfb framebuffer 未建立")

    checkpoints: list[dict] = []
    dosbox = None
    client = None
    try:
        with (outdir / "dosbox.log").open("wb") as logfh:
            dosbox = subprocess.Popen(
                [args.dosbox, "-conf", str(conf_path), "-nomenu"],
                stdout=logfh, stderr=subprocess.STDOUT,
                env=dict(env, DISPLAY=args.display, SDL_VIDEODRIVER="x11"))
            start = time.time()

            client = None
            window = None
            while time.time() - start < 30:
                time.sleep(0.5)
                try:
                    c = X11Client(args.display)
                except OSError:
                    continue
                kids = c.query_children()
                if kids:
                    client, window = c, kids[-1]
                    break
                c.close()
            if client is None:
                raise SystemExit("找不到 DOSBox-X 視窗")
            client.set_input_focus(window)

            cap = Capture(client, window, fb_path, framedir, start)
            if args.observe:
                end = time.time() + args.observe
                while time.time() < end:
                    cap.sample()
                    time.sleep(args.interval)
                steps = []
            for index, st in enumerate(steps):
                if st.get("delay") is not None:
                    end = time.time() + st["delay"]
                    while time.time() < end:
                        cap.sample()
                        time.sleep(args.interval)
                    got = cap.sample()
                    result = {"stable": True, "frame": got[1] if got else None,
                              "waited_s": st["delay"]}
                else:
                    result = cap.wait_stable(st["settle"], st["timeout"],
                                             args.interval, args.tolerance)
                checkpoints.append({
                    "index": index,
                    "label": st["label"],
                    "key": st["key"],
                    "settle_s": st["settle"],
                    "fixed_delay_s": st.get("delay"),
                    "reached_stable": result["stable"],
                    "t": round(time.time() - start, 2),
                    "frame": result["frame"],
                })
                if st["key"] is not None:
                    client.fake_key(st["key"])
                    time.sleep(args.post_key_delay)
    finally:
        if client is not None:
            client.close()
        if dosbox is not None and dosbox.poll() is None:
            dosbox.send_signal(signal.SIGTERM)
            try:
                dosbox.wait(timeout=15)
            except subprocess.TimeoutExpired:
                dosbox.kill()
        xvfb.terminate()
        try:
            xvfb.wait(timeout=10)
        except subprocess.TimeoutExpired:
            xvfb.kill()

    post = character_hashes(gamedir)
    changed = [n for n in CHARACTER_FILES if pre[n]["sha256"] != post[n]["sha256"]]
    if args.save_character_files:
        # 執行後的角色檔含原版模板 bytes，只作本機離線比對，不得提交。
        savedir = outdir / "character-files"
        savedir.mkdir(parents=True, exist_ok=True)
        for name in CHARACTER_FILES:
            (savedir / name).write_bytes((gamedir / name).read_bytes())
    frame_index = {f["frame"]: f for f in cap.timeline if "png_sha256" in f}

    receipt = {
        "tool": "tools/dosbox_character_ui_oracle.py",
        "scenario": scenario_name,
        "description": description,
        "generated_utc": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "dosbox_x_version": version_line,
        "dosbox_conf_sha256": hashlib.sha256(conf_text.encode()).hexdigest(),
        "dosbox_cycles": args.cycles,
        "input_method": "X11 XTEST FakeInput，畫面靜止後逐鍵送出",
        "sample_interval_s": args.interval,
        "stability_tolerance": args.tolerance,
        "original_zip": inventory,
        "character_files_before": pre,
        "character_files_after": post,
        "character_files_changed": changed,
        "character_files_saved": bool(args.save_character_files),
        "distinct_frames": len(cap.seen),
        "checkpoints": [
            dict(cp, file=frame_index.get(cp["frame"], {}).get("file"),
                 png_sha256=frame_index.get(cp["frame"], {}).get("png_sha256"))
            for cp in checkpoints
        ],
        "timeline": cap.timeline,
    }
    (outdir / "receipt.json").write_text(
        json.dumps(receipt, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({
        "scenario": scenario_name,
        "distinct_frames": len(cap.seen),
        "checkpoints": len(checkpoints),
        "unstable": [c["label"] for c in checkpoints if not c["reached_stable"]],
        "changed": changed,
    }, ensure_ascii=False))
    return 0


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--work", default="/work")
    p.add_argument("--out", default="/out")
    p.add_argument("--scenario", default="explore")
    p.add_argument("--steps-json", default=None, help="外部 step 定義檔，覆蓋 --scenario")
    p.add_argument("--cycles", default="fixed 3000")
    p.add_argument("--interval", type=float, default=0.3)
    p.add_argument("--post-key-delay", type=float, default=0.4)
    p.add_argument("--save-character-files", action="store_true",
                   help="把執行後的 CHARDATA／CHARPICS／SHLDPICS 複製到輸出目錄（僅本機比對用）")
    p.add_argument("--tolerance", type=float, default=0.002,
                   help="判定畫面靜止時允許的差異比例（閃爍游標用）")
    p.add_argument("--observe", type=float, default=0.0,
                   help="只觀察並錄畫格 N 秒，不送任何按鍵")
    p.add_argument("--display", default=":77")
    p.add_argument("--dosbox", default="/opt/dosbox-x/bin/dosbox-x")
    return run(p.parse_args())


if __name__ == "__main__":
    sys.exit(main())
