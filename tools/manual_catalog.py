"""從固定 KOL 說明書 PDF 建立逐頁文字雜湊清冊；英文全文只寫 ignored 工作區。"""

from __future__ import annotations

import argparse
import csv
import hashlib
import json
import re
import subprocess
from pathlib import Path


SOURCES = (
    {
        "id": "handbook",
        "filename": "knightslegend-manual.pdf",
        "sha256": "4c80f5d4c7221f9d3376546d249d52e2fc7fc0f73b48ee9daf6e20ce87d7086a",
        "pages": 79,
    },
    {
        "id": "reference-card",
        "filename": "knightslegend-refcard.pdf",
        "sha256": "4815bc6c3a51ff7abaa610a071f08e05eb2dd180e5130b0674de8b4a94b2abd3",
        "pages": 5,
    },
)
WORD = re.compile(r"[A-Za-z0-9][A-Za-z0-9'’-]*")
EXPECTED_BLANK = {"handbook/page/076", "handbook/page/077"}


def sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def extract_page(pdf: Path, page: int) -> str:
    result = subprocess.run(
        ["pdftotext", "-f", str(page), "-l", str(page), "-raw", str(pdf), "-"],
        check=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    text = result.stdout.decode("utf-8").replace("\f", "")
    lines = [line.rstrip() for line in text.splitlines()]
    while lines and not lines[-1]:
        lines.pop()
    return "\n".join(lines) + "\n"


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("root", type=Path)
    parser.add_argument("catalog", type=Path)
    parser.add_argument("local_source", type=Path)
    args = parser.parse_args()

    records = []
    local_rows = []
    inputs = []
    for source in SOURCES:
        pdf = args.root / source["filename"]
        raw = pdf.read_bytes()
        digest = sha256(raw)
        if digest != source["sha256"]:
            raise SystemExit(f"{pdf.name} SHA-256 不符：{digest}")
        inputs.append({
            "id": source["id"],
            "filename": source["filename"],
            "sha256": digest,
            "physical_pages": source["pages"],
        })
        for page in range(1, source["pages"] + 1):
            key = f"{source['id']}/page/{page:03d}"
            text = extract_page(pdf, page)
            record = {
                "key": key,
                "source_sha256": sha256(text.encode("utf-8")),
                "characters": len(text),
                "words": len(WORD.findall(text)),
                "lines": len(text.splitlines()),
                "text_layer_nonempty": bool(text.strip()),
                "page_kind": "expected_blank" if key in EXPECTED_BLANK else "content",
            }
            records.append(record)
            local_rows.append({"key": key, "source_sha256": record["source_sha256"], "source": text})

    result = {
        "schema": "kol-manual-source-catalog-v1",
        "extractor": "Poppler pdftotext -raw",
        "extractor_version": subprocess.run(
            ["pdftotext", "-v"], stderr=subprocess.PIPE, text=True, check=True
        ).stderr.splitlines()[0],
        "inputs": inputs,
        "page_count": len(records),
        "all_expected_text_layers_present": all(
            item["text_layer_nonempty"] or item["page_kind"] == "expected_blank"
            for item in records
        ),
        "pages": records,
    }
    args.catalog.parent.mkdir(parents=True, exist_ok=True)
    args.local_source.parent.mkdir(parents=True, exist_ok=True)
    args.catalog.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    with args.local_source.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.DictWriter(
            handle, fieldnames=("key", "source_sha256", "source"), delimiter="\t", lineterminator="\n"
        )
        writer.writeheader()
        writer.writerows(local_rows)
    print(
        f"pages={len(records)} "
        f"all_expected_text_layers_present={result['all_expected_text_layers_present']}"
    )


if __name__ == "__main__":
    main()
