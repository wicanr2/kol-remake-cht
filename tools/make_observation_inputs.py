#!/usr/bin/env python3
"""產生戰後結算的觀察用替代輸入。

第 5 個任務地點原本是十二隻 3 級 BINDERAK，單人隊伍打不贏，也就看不到打贏之後的
獎勵與戰利品畫面。這支工具從正式的 typed data 產生三份替代檔：地點改成一隻怪、
同一個地點再把隊伍挪到戰場邊上（看逃走那一問，RE-103），以及怪物本身壓到一擊就倒。

產物只給觀察用，**不是原版資料，不得當作對拍來源**；輸出寫進 `workplace/observe/`，
那個目錄不進版控。收據見 `docs/verification/battle-spoils/receipt.json`。

  tools/run_python_tool.sh tools/make_observation_inputs.py
"""

from __future__ import annotations

import copy
import json
from pathlib import Path

SITES_IN = Path("assets/original-derived-data/quest-sites.json")
MONSTERS_IN = Path("assets/original-derived-data/monsters.json")
OUT_DIR = Path("workplace/observe")
SITE_INDEX = 5
MONSTER_NAME = "BINDERAK"
NOTE = "觀察用替代輸入，不是原版資料，不得當作對拍來源。"


def one_enemy(sites: dict) -> dict:
    out = copy.deepcopy(sites)
    for s in out["sites"]:
        if s["index"] != SITE_INDEX:
            continue
        # 座標、任務旗標與敘述都不動，只換怪的數量、等級與位置。
        s["count"] = 1
        s["level"] = 1
        s["monster_y"] = [42]
        s["monster_x"] = [15]
    out["note_observation_only"] = (
        f"{NOTE}第 {SITE_INDEX} 個地點改成一隻 1 級怪、緊貼隊伍。")
    return out


def party_on_edge(sites: dict) -> dict:
    """隊伍改站在戰場可走區的最外圈，好看得到逃走那一問（RE-103）。

    原版的逃走判定要「有人站到 4 或 0x3B」才問，而正常玩法要從中央走二十幾格
    才到得了邊上，一回合走一格。站位換掉之後第一個回合結束就問得出來。
    """
    out = one_enemy(sites)
    for s in out["sites"]:
        if s["index"] != SITE_INDEX:
            continue
        s["party_x"] = [4] + s["party_x"][1:]
    out["note_observation_only"] = (
        f"{NOTE}第 {SITE_INDEX} 個地點改成一隻 1 級怪，隊伍第一格站到戰場上緣。")
    return out


def weak_monster(monsters: dict) -> dict:
    out = copy.deepcopy(monsters)
    for m in out["monsters"]:
        if m["name"] != MONSTER_NAME:
            continue
        # 壓到一擊就倒、幾乎打不中人。金冠骰式（記錄 +0x0F）不動，
        # 那是要觀察的東西之一。
        m["health"] = 1
        m["size"] = 1
        m["strength"] = 1
        m["attack_skill"] = 1
        m["parry_skill"] = 1
    out["note_observation_only"] = (
        f"{NOTE}{MONSTER_NAME} 的體力、體型、力量與攻防技能壓到 1。")
    return out


def main() -> None:
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    for name, data in (
        ("quest-sites-one-enemy.json", one_enemy(json.loads(SITES_IN.read_text()))),
        ("quest-sites-edge.json", party_on_edge(json.loads(SITES_IN.read_text()))),
        ("monsters-weak.json", weak_monster(json.loads(MONSTERS_IN.read_text()))),
    ):
        path = OUT_DIR / name
        path.write_text(json.dumps(data, ensure_ascii=False, indent=1) + "\n",
                        encoding="utf-8")
        print(path)


main()
