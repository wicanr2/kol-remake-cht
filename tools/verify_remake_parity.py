#!/usr/bin/env python3
"""remake 畫面對原版同狀態畫面的對拍。

判準分兩層，因為繁體中文重製不可能在文字上逐像素等於原版英文字模：

- **素材／版面層**：允許差異區以外的每個像素都必須相同。差一個像素就是失敗。
- **允許差異區**：由計畫明文列出並附理由，例如原版畫的滑鼠游標、原版自己繪製的
  文字區。允許不等於忽略——收據會記錄每個允許區內實際差了多少像素。
- **必須相同區**：優先於允許差異區。用來在一大片「文字可以不同」的區域裡，仍然
  釘住素材必須逐像素相同的矩形，例如角色圖像的 blit 位置。

計畫是 JSON：

    {"pairs": [{"label": "...", "remake": "...", "original": "...",
                "allow": [{"name": "...", "x": 0, "y": 0, "width": 0, "height": 0}]}]}

  python3 tools/verify_remake_parity.py --plan 計畫.json --out 收據.json
"""

from __future__ import annotations

import argparse
import json
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from png_io import read_png  # noqa: E402


def compare_pair(entry: dict, root: pathlib.Path) -> dict:
    a = read_png(root / entry["remake"])
    b = read_png(root / entry["original"])
    if (a.width, a.height) != (b.width, b.height):
        return {"label": entry["label"], "pass": False,
                "reason": f"尺寸不同 {a.width}x{a.height} vs {b.width}x{b.height}"}
    ra, rb = a.rgb(), b.rgb()

    allow = entry.get("allow", [])
    require = entry.get("require", [])
    allowed = bytearray(a.width * a.height)
    for region in allow:
        for y in range(region["y"], region["y"] + region["height"]):
            base = y * a.width
            for x in range(region["x"], region["x"] + region["width"]):
                allowed[base + x] = 1
    for region in require:          # 必須相同區覆蓋允許差異區
        for y in range(region["y"], region["y"] + region["height"]):
            base = y * a.width
            for x in range(region["x"], region["x"] + region["width"]):
                allowed[base + x] = 0

    outside = 0
    first_outside = None
    per_region = {r["name"]: 0 for r in allow}
    region_index = []
    for r in allow:
        region_index.append((r["name"], r["x"], r["y"],
                             r["x"] + r["width"], r["y"] + r["height"]))
    for i in range(a.width * a.height):
        o = i * 3
        if ra[o:o + 3] == rb[o:o + 3]:
            continue
        x, y = i % a.width, i // a.width
        if allowed[i]:
            for name, x0, y0, x1, y1 in region_index:
                if x0 <= x < x1 and y0 <= y < y1:
                    per_region[name] += 1
                    break
        else:
            outside += 1
            if first_outside is None:
                first_outside = {"x": x, "y": y}

    return {
        "label": entry["label"],
        "remake": entry["remake"],
        "original": entry["original"],
        "size": [a.width, a.height],
        "pass": outside == 0,
        "differing_outside_allowed": outside,
        "first_outside_pixel": first_outside,
        "allowed_regions": [
            {"name": r["name"], "reason": r.get("reason"),
             "rect": [r["x"], r["y"], r["width"], r["height"]],
             "differing_pixels": per_region[r["name"]]}
            for r in allow
        ],
        "required_regions": [
            {"name": r["name"], "reason": r.get("reason"),
             "rect": [r["x"], r["y"], r["width"], r["height"]]}
            for r in require
        ],
    }


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--plan", required=True)
    p.add_argument("--root", default=".")
    p.add_argument("--out", default=None)
    args = p.parse_args()

    root = pathlib.Path(args.root)
    plan = json.loads(pathlib.Path(args.plan).read_text())
    rows = [compare_pair(e, root) for e in plan["pairs"]]
    failures = [r for r in rows if not r["pass"]]
    result = {
        "tool": "tools/verify_remake_parity.py",
        "plan": plan.get("name", args.plan),
        "description": plan.get("description"),
        "pair_count": len(rows),
        "pass_count": len(rows) - len(failures),
        "failure_count": len(failures),
        "pairs": rows,
    }
    text = json.dumps(result, ensure_ascii=False, indent=2) + "\n"
    if args.out:
        pathlib.Path(args.out).write_text(text)
    print(json.dumps({"pairs": len(rows), "pass": result["pass_count"],
                      "fail": result["failure_count"]}, ensure_ascii=False))
    return 0 if not failures else 1


if __name__ == "__main__":
    sys.exit(main())
