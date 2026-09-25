#!/usr/bin/env python3
"""以中文模式重播玩家路徑計畫，彙整畫面上仍含英文字母的字串。

每條計畫沿用 `check_player_paths.py` 的旗標，另加 `-lang zh-Hant` 與 `-text-log`，
只截最後一個檢查點。輸出在 `workplace/cjk-leaks/`：

  tools/scan_cjk_text_leaks.py run town-shop world-quest   # 重播指定計畫
  tools/scan_cjk_text_leaks.py run --all --jobs 3          # 全部，分三個容器
  tools/scan_cjk_text_leaks.py run-battle                  # 重播 docs/verification/battle-*/replay.spec
  tools/scan_cjk_text_leaks.py report                      # 彙整已有的紀錄

玩家自取的角色名、數字與標點不算漏譯；報表把「只含玩家名稱」的字串另列，
其餘含拉丁字母的字串都要人判。這支只看得到字模繪製的字串，畫在原版圖片裡
的英文字樣要另外看畫格。在主機上執行，重播本身在 `run_kol_shots.sh` 的容器裡。
"""

from __future__ import annotations

import argparse
import concurrent.futures
import json
import pathlib
import re
import subprocess
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from check_player_paths import PLANS, spec_line  # noqa: E402

OUT = pathlib.Path("workplace/cjk-leaks")
LATIN = re.compile(r"[A-Za-z]")


def plan_names() -> list[str]:
    return sorted(f.stem for f in PLANS.glob("*.json") if not f.stem.endswith("-receipt"))


def line_for(name: str, all_checkpoints: bool = False) -> str:
    plan = json.loads((PLANS / f"{name}.json").read_text(encoding="utf-8"))
    shot_dir = OUT / name
    shot_dir.mkdir(parents=True, exist_ok=True)
    steps = sorted({c["step"] for c in plan.get("checkpoints", [])})
    if steps and not all_checkpoints:
        plan = {**plan, "checkpoints": [{"step": steps[-1]}]}
    line = spec_line(plan, shot_dir)
    head, rest = line.split("|", 1)
    return "|".join([head, "-lang", "zh-Hant", "-text-log", str(shot_dir / "text.jsonl"), rest])


def battle_lines() -> list[str]:
    """戰鬥情境重播規格：輸出改寫到 OUT/battle-<情境>-<行>/，再加語言與文字紀錄。"""
    lines = []
    for spec in sorted(pathlib.Path("docs/verification").glob("battle-*/replay.spec")):
        rows = [r for r in spec.read_text(encoding="utf-8").splitlines()
                if r.strip() and not r.startswith("#")]
        for i, row in enumerate(rows):
            out = OUT / f"{spec.parent.name}-{i}"
            out.mkdir(parents=True, exist_ok=True)
            fields = row.split("|")
            fields = [re.sub(r"^workplace/tmp/[^/]+", str(out), f) for f in fields]
            fields = [f for f in fields if not f.startswith("-lang")]
            lines.append("|".join([fields[0], "-lang", "zh-Hant", "-text-log",
                                   str(out / "text.jsonl"), *fields[1:]]))
    return lines


def run_lines(index: int, lines: list[str]) -> tuple[int, str]:
    spec = OUT / f"batch-{index}.spec"
    spec.write_text("".join(line + "\n" for line in lines), encoding="utf-8")
    return _run_spec(spec)


FRAMES = False


def run_batch(index: int, names: list[str]) -> tuple[int, str]:
    return run_lines(index, [line_for(n, FRAMES) for n in names])


def _run_spec(spec: pathlib.Path) -> tuple[int, str]:
    proc = subprocess.run(["tools/run_kol_shots.sh", str(spec)], capture_output=True, text=True,
                          env={**__import__("os").environ, "KOL_GO_TIMEOUT": "7200"})
    return proc.returncode, proc.stderr[-1500:]


def names_in_saves() -> set[str]:
    names = set()
    for path in pathlib.Path("workplace/saves").glob("*.json"):
        try:
            doc = json.loads(path.read_text())
        except (ValueError, UnicodeDecodeError):
            continue
        for member in doc.get("characters", []) or doc.get("roster", []) or []:
            if isinstance(member, dict) and member.get("name"):
                names.add(member["name"].strip())
    return names


def report() -> int:
    players = names_in_saves()
    seen: dict[tuple[str, str], set[str]] = {}
    for log in sorted(OUT.glob("*/text.jsonl")):
        for raw in log.read_text(encoding="utf-8").splitlines():
            row = json.loads(raw)
            if LATIN.search(row["text"]):
                seen.setdefault((row["font"], row["text"]), set()).add(log.parent.name)
    player_only, leaks = [], []
    upper = {name.upper() for name in players}
    for (font, text), where in sorted(seen.items()):
        item = {"font": font, "text": text, "plans": sorted(where)}
        # 玩家名可能整段大寫，或被名牌截成前五字（至少四字才算，避免誤吞單字母）。
        key = text.strip().upper()
        is_player = key in upper or (len(key) >= 4 and any(n.startswith(key) for n in upper))
        (player_only if is_player else leaks).append(item)
    result = {"plans_logged": sorted(p.parent.name for p in OUT.glob("*/text.jsonl")),
              "player_names": sorted(players), "latin_strings": len(leaks),
              "player_name_only": player_only, "leaks": leaks}
    (OUT / "report.json").write_text(json.dumps(result, ensure_ascii=False, indent=1) + "\n",
                                     encoding="utf-8")
    print(f"計畫 {len(result['plans_logged'])} 條；含英文字母 {len(leaks)} 筆，"
          f"只含玩家名稱 {len(player_only)} 筆 → {OUT / 'report.json'}")
    return 0


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__,
                                formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = p.add_subparsers(dest="cmd", required=True)
    r = sub.add_parser("run")
    r.add_argument("names", nargs="*")
    r.add_argument("--all", action="store_true")
    r.add_argument("--jobs", type=int, default=1)
    r.add_argument("--frames", action="store_true", help="保留計畫的全部檢查點畫格，供目視檢查")
    sub.add_parser("report")
    b = sub.add_parser("run-battle")
    b.add_argument("--jobs", type=int, default=1)
    args = p.parse_args()
    if args.cmd == "report":
        return report()
    if args.cmd == "run-battle":
        OUT.mkdir(parents=True, exist_ok=True)
        lines = battle_lines()
        groups = [lines[i::args.jobs] for i in range(args.jobs)]
        with concurrent.futures.ThreadPoolExecutor(args.jobs) as pool:
            codes = list(pool.map(run_lines, [100 + i for i in range(args.jobs)], groups))
        for code, err in codes:
            if code != 0:
                print(err, file=sys.stderr)
        report()
        return 1 if any(c for c, _ in codes) else 0
    global FRAMES
    FRAMES = args.frames
    names = plan_names() if args.all else args.names
    if not names:
        p.error("要指定計畫名，或加 --all")
    OUT.mkdir(parents=True, exist_ok=True)
    batches = [names[i::args.jobs] for i in range(args.jobs)]
    failed = False
    with concurrent.futures.ThreadPoolExecutor(args.jobs) as pool:
        for i, (code, err) in enumerate(pool.map(run_batch, range(args.jobs), batches)):
            if code != 0:
                failed = True
                print(f"batch-{i} 失敗：{err}", file=sys.stderr)
    report()
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
