#!/usr/bin/env python3
"""在 Docker 中核對正式翻譯文字的字模涵蓋率；不把研究註記當成玩家文字。"""
import argparse
import csv
import hashlib
import json
from pathlib import Path


def strings(value, location):
    if isinstance(value, str):
        yield location, value
    elif isinstance(value, list):
        for index, item in enumerate(value):
            yield from strings(item, f"{location}[{index}]")
    elif isinstance(value, dict):
        for key, item in value.items():
            yield from strings(item, f"{location}.{key}")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path("."))
    parser.add_argument("--report", type=Path)
    args = parser.parse_args()
    hashes = {}

    def read(name):
        raw = (args.root / name).read_bytes()
        hashes[name] = hashlib.sha256(raw).hexdigest()
        return json.loads(raw)

    manifest = read("assets/ui-font/manifest.json")
    fonts = [f for f in manifest["fonts"] if f["name"] == "zh-hant-16"]
    if len(fonts) != 1:
        raise SystemExit("找不到唯一的 zh-hant-16 字型")
    supported = {g["codepoint"] for g in fonts[0]["glyphs"]}
    paths = ["docs/i18n/newchar-zh-Hant.json", "docs/i18n/town-zh-Hant.json", "docs/i18n/ui-zh-Hant.json"]
    creation, town, ui = [read(p) for p in paths]
    entries = []
    for key in ("labels", "races", "secondary", "professions", "ui"):
        for i, row in enumerate(creation[key]):
            entries.append((f"{paths[0]}.{key}[{i}].zh", row["zh"]))
    entries.extend(strings(town["towns"], paths[1] + ".towns"))
    # 城鎮檔中其他可能使用的翻譯區塊也要涵蓋，排除明確的文件中繼資料。
    for key, value in town.items():
        if key not in {"schema", "note", "status", "line_limits", "towns", "source"}:
            entries.extend(strings(value, paths[1] + "." + key))
    for bank, records in ui["banks"].items():
        for number, row in records.items():
            entries.append((f"{paths[2]}.banks.{bank}.{number}.zh", row["zh"]))
    for names_path in ["docs/i18n/display-names-zh-Hant.tsv", "docs/i18n/world-prompts-zh-Hant.tsv"]:
        names_raw = (args.root / names_path).read_bytes()
        hashes[names_path] = hashlib.sha256(names_raw).hexdigest()
        with (args.root / names_path).open(encoding="utf-8", newline="") as handle:
            rows = csv.DictReader(handle, delimiter="\t")
            if rows.fieldnames != ["key", "en", "zh-Hant", "status", "source"]:
                raise SystemExit(f"翻譯 TSV 欄位不符：{names_path}")
            for row in rows:
                entries.append((f"{names_path}.{row['key']}", row["zh-Hant"]))
    symbols_path = "docs/i18n/ui-symbols-zh-Hant.json"
    symbols = read(symbols_path)
    for key, value in symbols["symbols"].items():
        entries.append((f"{symbols_path}.symbols.{key}", value))
    words_path = "docs/i18n/spell-words-zh-Hant.json"
    words = read(words_path)
    for part in ("race", "stat", "severity", "range", "duration"):
        for i, row in enumerate(words[part]):
            entries.append((f"{words_path}.{part}[{i}].zh", row["zh"]))
    entries.append((f"{words_path}.gloss-prefix", "持續"))
    title_path = "docs/i18n/title-zh-Hant.json"
    title = read(title_path)
    for key, row in title["backs"].items():
        entries.append((f"{title_path}.backs.{key}.zh", row["zh"]))
    entries.append((f"{title_path}.cover.zh", title["cover"]["zh"]))
    missing = []
    for location, value in entries:
        absent = sorted({c for c in value if c.isprintable() and not c.isspace() and ord(c) not in supported})
        if absent:
            missing.append({"location": location, "characters": "".join(absent), "text": value})
    report = {"schema": "kol-cjk-coverage-v1", "inputs_sha256": hashes,
              "checked_strings": len(entries), "missing": missing,
              "scope": "翻譯資料的字元涵蓋；不證明 PNG 墨跡、排版、程式動態字串或正常玩家路徑"}
    output = json.dumps(report, ensure_ascii=False, indent=2)
    if args.report:
        args.report.write_text(output + "\n", encoding="utf-8")
    print(output)
    return bool(missing)


if __name__ == "__main__":
    raise SystemExit(main())
