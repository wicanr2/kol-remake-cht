#!/usr/bin/env python3
"""用原版角色圖像選擇畫面驗證抽取出來的 80 張 CHARPICS 模板。

`picture-cycle` 情境會讓原版把 80 張模板依序畫在同一個位置，因此每一張抽取資產都有
一張原版自己畫的同狀態畫面可對。比對只取該 blit 矩形，差異必須為 0；任何一張不符即
整體失敗。這是 SPEC-002 抽取器的原版 parity 收據，不是 remake 的 parity。

  python3 tools/verify_charpics_against_capture.py --run 擷取目錄 \\
      --assets assets/original-derived-png/ega/CHARPICS --out 收據.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from png_io import read_png  # noqa: E402

BLIT = {"x": 208, "y": 106, "width": 40, "height": 38}
TEMPLATE_COUNT = 80


def selector_frames(receipt: dict) -> dict[int, str]:
    """把 picture-cycle 的檢查點對回 selector 編號。

    步驟固定為：selector 0 → 向左一次（回繞到 79）→ 向右一次（回到 0）→
    連續向右。因此 checkpoint 14 是 selector 0、15 是 79、16 起是 0,1,2,…
    """
    cps = {c["index"]: c for c in receipt["checkpoints"]}
    if cps[14]["label"] != "圖像層 selector 0":
        raise SystemExit("這份收據不是 picture-cycle 情境")
    mapping = {0: cps[14]["file"], 79: cps[15]["file"]}
    for k in range(1, 79):
        cp = cps[16 + k]
        if cp["label"] != f"向右第 {k + 1} 次之前":
            raise SystemExit(f"checkpoint {16 + k} 標籤不符：{cp['label']}")
        mapping[k] = cp["file"]
    if cps[95]["file"] != cps[15]["file"]:
        raise SystemExit("向右 80 次未回到向左一次的同一張畫面，循環契約不成立")
    if cps[96]["file"] != cps[14]["file"]:
        raise SystemExit("向右 81 次未回到 selector 0，循環契約不成立")
    return mapping


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--run", required=True, help="picture-cycle 擷取輸出目錄")
    p.add_argument("--assets", required=True, help="CHARPICS 抽取資產目錄")
    p.add_argument("--out", default=None)
    args = p.parse_args()

    run = pathlib.Path(args.run)
    assets = pathlib.Path(args.assets)
    receipt = json.loads((run / "receipt.json").read_text())
    mapping = selector_frames(receipt)
    if sorted(mapping) != list(range(TEMPLATE_COUNT)):
        raise SystemExit("selector 對照未涵蓋 0..79")

    rows = []
    for selector in range(TEMPLATE_COUNT):
        frame = read_png(run / mapping[selector])
        crop = frame.crop(BLIT["x"], BLIT["y"], BLIT["width"], BLIT["height"]).rgb()
        asset_path = assets / f"charpics-{selector:03d}.png"
        asset = read_png(asset_path)
        if (asset.width, asset.height) != (BLIT["width"], BLIT["height"]):
            raise SystemExit(f"{asset_path} 尺寸不是 {BLIT['width']}x{BLIT['height']}")
        arg = asset.rgb()
        differing = sum(1 for i in range(0, len(arg), 3)
                        if arg[i:i + 3] != crop[i:i + 3])
        rows.append({
            "selector": selector,
            "frame": mapping[selector],
            "asset": f"CHARPICS/charpics-{selector:03d}.png",
            "asset_png_sha256": hashlib.sha256(asset_path.read_bytes()).hexdigest(),
            "differing_pixels": differing,
        })

    failures = [r for r in rows if r["differing_pixels"]]
    result = {
        "tool": "tools/verify_charpics_against_capture.py",
        "purpose": "抽取資產對原版執行畫面的同狀態 parity",
        "capture_run": receipt["scenario"],
        "capture_generated_utc": receipt["generated_utc"],
        "dosbox_conf_sha256": receipt["dosbox_conf_sha256"],
        "blit_rectangle": BLIT,
        "template_count": TEMPLATE_COUNT,
        "exact_count": TEMPLATE_COUNT - len(failures),
        "failure_count": len(failures),
        "templates": rows,
    }
    text = json.dumps(result, ensure_ascii=False, indent=2) + "\n"
    if args.out:
        pathlib.Path(args.out).write_text(text)
    print(json.dumps({"exact": result["exact_count"], "failures": result["failure_count"]},
                     ensure_ascii=False))
    return 0 if not failures else 1


if __name__ == "__main__":
    sys.exit(main())
