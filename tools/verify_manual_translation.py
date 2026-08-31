"""驗證手冊逐頁狀態、來源雜湊錨點與已校對 Markdown 的基本完整性。"""

from __future__ import annotations

import argparse
import csv
import json
import re
from pathlib import Path


MARKER = re.compile(r"<!--\s*(handbook|reference-card)/page/(\d{3})(?:｜[^>]*)?\s*-->")
NUMBER = re.compile(r"(?<![\w.,-])(\d{1,3}(?:,\d{3})+|\d+)(?![\w,])")
CJK = re.compile(r"[\u3400-\u9fff]")
WORD = re.compile(r"[A-Za-z][A-Za-z'’-]*")
ENGLISH_RUN = re.compile(r"(?:\b[A-Za-z][A-Za-z'’.-]*\b[ ,]+){7}\b[A-Za-z]")
CJK_VALUES = {
    "零": 0, "一": 1, "二": 2, "兩": 2, "三": 3, "四": 4,
    "五": 5, "六": 6, "七": 7, "八": 8, "九": 9,
}
CJK_UNITS = {"十": 10, "百": 100, "千": 1000, "萬": 10000}
CJK_NUMBER = re.compile(r"[零一二兩三四五六七八九十百千萬]+")


def cjk_to_int(text: str) -> int | None:
    total = section = 0
    digit = None
    for char in text:
        if char in CJK_VALUES:
            digit = CJK_VALUES[char]
        elif char == "萬":
            total += ((section + (digit or 0)) or 1) * 10000
            section, digit = 0, None
        else:
            section += (digit if digit is not None else 1) * CJK_UNITS[char]
            digit = None
    value = total + section + (digit or 0)
    return value or None


def numbers(text: str) -> set[str]:
    result = {token.replace(",", "") for token in NUMBER.findall(text)}
    for run in CJK_NUMBER.findall(text):
        value = cjk_to_int(run)
        if value is not None:
            result.add(str(value))
    return result


def expand_ranges(ranges: list[list[int]]) -> set[int]:
    pages: set[int] = set()
    for start, end in ranges:
        if start > end:
            raise ValueError(f"反向頁碼範圍：{start}..{end}")
        pages.update(range(start, end + 1))
    return pages


def markdown_segments(root: Path, outputs: list[str]) -> dict[str, str]:
    segments: dict[str, list[str]] = {}
    for name in outputs:
        text = (root / name).read_text(encoding="utf-8")
        matches = list(MARKER.finditer(text))
        for index, match in enumerate(matches):
            key = f"{match.group(1)}/page/{match.group(2)}"
            end = matches[index + 1].start() if index + 1 < len(matches) else len(text)
            segments.setdefault(key, []).append(text[match.end():end].strip())
    return {key: "\n".join(parts) for key, parts in segments.items()}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("root", type=Path)
    parser.add_argument("source_catalog", type=Path)
    parser.add_argument("status", type=Path)
    parser.add_argument("local_source", type=Path)
    args = parser.parse_args()
    source_catalog = json.loads(args.source_catalog.read_text(encoding="utf-8"))
    status = json.loads(args.status.read_text(encoding="utf-8"))
    with args.local_source.open(encoding="utf-8", newline="") as handle:
        sources = {row["key"]: row for row in csv.DictReader(handle, delimiter="\t")}
    catalog = {item["key"]: item for item in source_catalog["pages"]}
    if set(catalog) != set(sources):
        raise SystemExit("來源清冊與本機原文 key 集合不一致")
    for key, item in catalog.items():
        if sources[key]["source_sha256"] != item["source_sha256"]:
            raise SystemExit(f"{key}: 本機原文雜湊漂移")

    outputs = sorted({path for doc in status["documents"].values() for path in doc["outputs"]})
    segments = markdown_segments(args.root, outputs)
    reviewed_total = translated_total = todo_total = 0
    for document, doc in status["documents"].items():
        reviewed = set(doc["reviewed_pages"])
        translated = set(doc["translated_pages"])
        todo = expand_ranges(doc["todo_ranges"])
        blank = set(doc["blank_pages"])
        expected = set(range(1, doc["physical_pages"] + 1))
        if reviewed & translated or reviewed & todo or translated & todo:
            raise SystemExit(f"{document}: 狀態頁碼重疊")
        if reviewed | translated | todo != expected:
            raise SystemExit(f"{document}: 狀態頁碼未完整覆蓋")
        if not blank <= reviewed:
            raise SystemExit(f"{document}: 空白頁必須經目視後標 reviewed")
        reviewed_total += len(reviewed)
        translated_total += len(translated)
        todo_total += len(todo)
        for page in reviewed - blank:
            key = f"{document}/page/{page:03d}"
            translation = segments.get(key, "")
            if not translation:
                raise SystemExit(f"{key}: reviewed 但沒有 Markdown 區塊")
            source = sources[key]["source"]
            source_numbers = {n for n in numbers(source) if int(n) > 1}
            missing = sorted(source_numbers - numbers(translation), key=int)
            if missing:
                raise SystemExit(f"{key}: 譯文缺少原文數字 {missing}")
            words = len(WORD.findall(source))
            ratio = len(CJK.findall(translation)) / max(words, 1)
            if words >= 50 and not 0.65 <= ratio <= 3.0:
                raise SystemExit(f"{key}: 中英篇幅比 {ratio:.2f} 可疑")
            if ENGLISH_RUN.search(translation):
                raise SystemExit(f"{key}: 疑似殘留整段英文")
    print(
        f"pages={len(catalog)} reviewed={reviewed_total} translated={translated_total} "
        f"todo={todo_total} outputs={len(outputs)}"
    )


if __name__ == "__main__":
    main()
