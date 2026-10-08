#!/usr/bin/env python3
"""在 Docker 內核實 SPEC-005 的三份原版名冊邊界收據；不宣稱重製版 parity。"""
import argparse
import hashlib
import json
import pathlib
import re

DEFAULT = pathlib.Path(__file__).resolve().parents[1] / "docs/verification/character-ui/roster-boundaries-20261001"
REVISION = "968ca774705853276ae55c8d1163c369481ac71f"
FILES = {"CHARDATA", "CHARPICS", "SHLDPICS"}


def require(value, message):
    if not value:
        raise ValueError(message)


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def verify(folder):
    cases = {}
    for case, count in [("full", 46), ("duplicate", 24), ("charset", 223)]:
        d = json.loads((folder / f"{case}-receipt.json").read_text())
        require(d["schema"] == "kol-book-roster-reviewed-v1" and d["scenario"] == case and d["status"] == "passed", f"{case} 收據未通過")
        require(d["evidence_level"] == "已證實", "缺證據等級")
        require(d["tool"]["dosgolem_execution_revision"] == REVISION, "未固定已驗證工具")
        require(d["source"]["sha256"] == "2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46", "原始 ZIP 不符")
        entries = {x["entry"].rsplit("/", 1)[-1]: x for x in d["source"]["entries"]}
        require(entries["KOLBOOK.EXE"]["sha256"] == "2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832", "原版程式不符")
        require(entries["CHARDATA"]["bytes"] == 16 * 768, "原版槽位數不符")
        for name in FILES:
            require(entries[name]["sha256"] == d["before"][name], "初始三檔未對上來源")
        events = d["seed"]["events"]
        require(d["seed"]["original"] == 1 and d["seed"]["remake"] is None and len(events) == 1 and events[0]["fixed_arg"] == 1, "種子未預先固定")
        require(events[0]["ida"] == "0x18922" and events[0]["caller_ida"] == "0x1015b", "種子定位不符")
        steps = d["steps"]
        require(len(steps) == count and [s["number"] for s in steps] == list(range(count)), "輸入序列不完整")
        for s in steps:
            require(set(s["file_sha256"]) == FILES and len(s["slot_names"]) == 16, "狀態缺角色資料")
            require(all(re.fullmatch("[0-9a-f]{64}", h) for h in s["file_sha256"].values()), "資料雜湊格式錯誤")
        require(d["frames"], "缺目視畫格")
        for frame in d["frames"]:
            path = folder / frame["path"]
            require(path.resolve().parent == folder.resolve(), "畫格路徑越界")
            require(frame["status"] == "visually-reviewed" and frame["rights"] == "private-github-only-original-runtime-png", "目視／權利未確認")
            require(digest(path) == frame["sha256"], "PNG 雜湊不符")
            require(path.read_bytes()[:8] == b"\x89PNG\r\n\x1a\n", "不是 PNG")
            require(frame["source_sha256"] == entries["KOLBOOK.EXE"]["sha256"], "PNG 未連原版來源")
        commits = [s for s in steps if s["label"] == "接受肖像並提交"]
        require(all(s["file_sha256"] == d["before"] for s in steps[:commits[0]["number"]]), "提交前發生寫入")
        if case == "full":
            require(len(commits) == 3 and commits[0]["slot_names"][14:] == ["NOVA", ""]
                    and commits[1]["slot_names"][14:] == ["NOVA", "LUNA"], "不是正常填滿兩個原空槽")
            require(d["changed_slots"] == {"CHARDATA": [14, 15], "CHARPICS": [], "SHLDPICS": []}, "全滿差分超出空槽")
            require(commits[2]["number"] == 43 and d["detail"]["notice_step"] == 43, "滿槽訊息時點不符")
            require(all(s["file_sha256"] == commits[1]["file_sha256"] for s in steps[commits[1]["number"]:]), "滿槽拒絕後仍有寫入")
            require(d["after"] == commits[1]["file_sha256"], "拒絕後資料變動")
        elif case == "duplicate":
            require(len(commits) == 1 and commits[0]["number"] == 21, "重名提交時點不符")
            require(d["changed_slots"] == {"CHARDATA": [0], "CHARPICS": [0], "SHLDPICS": []}, "重名未限於原角色槽")
            require(commits[0]["slot_names"][0] == "Theodoric" and commits[0]["slot_names"][14:] == ["", ""], "重名另占空槽")
            require(d["after"]["CHARDATA"] != d["before"]["CHARDATA"], "未實際覆寫")
        else:
            rows = d["detail"]["ascii"]
            require([r["codepoint"] for r in rows] == [f"U+{n:04X}" for n in range(32, 127)], "ASCII 95 項漏驗")
            for row in rows:
                ch = chr(int(row["codepoint"][2:], 16))
                accepted = ch == " " or "A" <= ch <= "Z" or "a" <= ch <= "z"
                require(row["key"] == ch and row["accepted"] == accepted and row["result"] == (ch if accepted else ""), "字元結果與已證實契約不同")
                label = f"ASCII {row['codepoint']}：結束輸入並讀取名字"
                state = next(s for s in steps if s["label"] == label)
                require(ord(ch) in state["read_ascii"] and state["keys_pending"] == state["dos_pending"] == 0, "字元未被 DOS 實際消費")
                require(state["resolved_name_buffer"] == row["result"] and state["file_sha256"] == d["before"], "字元結果／提交前資料不符")
            require(sum(r["accepted"] for r in rows) == 53, "接受字元數不符")
            require(d["changed_slots"] == {"CHARDATA": [14], "CHARPICS": [], "SHLDPICS": []}, "混合名字差分不符")
            require(len(commits) == 1 and commits[0]["slot_names"][14] == "MiXeD Name", "大小寫與空白未保存")
        require(steps[-1]["file_sha256"] == d["after"], "最終資料摘要不同")
        cases[case] = d
    return cases


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--receipts", type=pathlib.Path, default=DEFAULT)
    args = parser.parse_args()
    try:
        verify(args.receipts)
    except (ValueError, KeyError, OSError, StopIteration, json.JSONDecodeError) as error:
        parser.exit(1, f"名冊收據核實失敗：{error}\n")
    print("三份名冊收據通過：固定種子、95 項字元、提交時點、資料差分與四張 PNG 雜湊一致。")


if __name__ == "__main__":
    main()

