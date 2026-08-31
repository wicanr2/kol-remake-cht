"""量化 COMBPICS 測試 PNG 與魔旅士原版截圖；只輸出安全 metadata。"""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageChops, ImageStat


ARTICLE_URL = (
    "https://morrisjfwong.com/blog/"
    "knights-of-legend-%E6%AD%A6%E5%A3%AB%E5%82%B3%E8%AA%AA-dos-game/"
)
CASES = (
    {
        "archive_index": 0,
        "probe": "combpics-0-ega-candidate.png",
        "probe_sha256": "6f2d94d3133add2194a8278491f5068e12ae7440b31f1b81218b6a70a6060ede",
        "oracle": "morris-kol-flee-001a-s.jpg",
        "oracle_url": "https://morrisjfwong.com/blog/wp-content/uploads/2010/03/kol-flee-001a-s.jpg",
        "oracle_sha256": "29752354cdcb994896f4f0545dfb6a6e795e095845c3edb0be5b4ca4c02505b0",
    },
    {
        "archive_index": 1,
        "probe": "combpics-1-ega-candidate.png",
        "probe_sha256": "090b6ffeeae5218269a2c6fc3b7b1e4be31515269fd00386c6d1f85cb18bf6d9",
        "oracle": "morris-war-lose-001a-s.jpg",
        "oracle_url": "https://morrisjfwong.com/blog/wp-content/uploads/2010/04/war-lose-001a-s.jpg",
        "oracle_sha256": "bda145b799afdce2193aba543eb3ec7a1ab8b059f7c07b438e691ac48b5a5faa",
    },
)
CROP = (7, 16, 273, 145)


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def mean_absolute_rgb_difference(left: Image.Image, right: Image.Image) -> float:
    difference = ImageStat.Stat(ImageChops.difference(left, right))
    return sum(difference.mean) / 3


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("probe_directory", type=Path)
    parser.add_argument("oracle_directory", type=Path)
    parser.add_argument("output_path", type=Path)
    args = parser.parse_args()

    probes = []
    oracles = []
    for case in CASES:
        probe_path = args.probe_directory / case["probe"]
        oracle_path = args.oracle_directory / case["oracle"]
        if digest(probe_path) != case["probe_sha256"]:
            raise SystemExit(f"測試 PNG SHA-256 不符：{probe_path}")
        if digest(oracle_path) != case["oracle_sha256"]:
            raise SystemExit(f"原版截圖 SHA-256 不符：{oracle_path}")
        probe = Image.open(probe_path).convert("RGB")
        oracle_full = Image.open(oracle_path).convert("RGB")
        if probe.size != (296, 144) or oracle_full.size != (288, 189):
            raise SystemExit("圖片尺寸不符合預先登記的對照契約")
        oracle = oracle_full.crop(CROP)
        probes.append(probe.resize(oracle.size, Image.Resampling.BICUBIC))
        oracles.append(oracle)

    matrix = [
        [mean_absolute_rgb_difference(probe, oracle) for oracle in oracles]
        for probe in probes
    ]
    if not (matrix[0][0] < matrix[0][1] / 2 and matrix[1][1] < matrix[1][0] / 2):
        raise SystemExit("正配圖片與錯配圖片的差異分離不足")

    result = {
        "schema": "kol-combpics-visual-oracle-v1",
        "source_article": ARTICLE_URL,
        "rights": "external copyrighted screenshots; repository stores metadata only",
        "probe_size": [296, 144],
        "oracle_full_size": [288, 189],
        "oracle_crop_xyxy": list(CROP),
        "oracle_crop_size": [CROP[2] - CROP[0], CROP[3] - CROP[1]],
        "resampling": "Pillow BICUBIC",
        "mean_absolute_rgb_difference_matrix": matrix,
        "acceptance": "each registered match is less than half its cross-mismatch",
        "cases": [dict(case) for case in CASES],
        "claim_limit": (
            "confirms scene identity, geometry and orientation; JPEG screenshot does not prove "
            "byte-exact RGB equality"
        ),
    }
    args.output_path.parent.mkdir(parents=True, exist_ok=True)
    args.output_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
