#!/usr/bin/env python3
"""唯讀盤點《武士傳說》固定研究輸入，輸出可重生 JSON。"""

from __future__ import annotations

import argparse
import hashlib
import json
import struct
import zipfile
from pathlib import Path


INPUTS = (
    "Knights-of-legend.zip",
    "knightslegend-manual.pdf",
    "knightslegend-refcard.pdf",
    "KOL-MAP-001a4s2.jpg",
    "knightslegend.jpg",
    "knightslegend-back.jpg",
)


def sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def mz_header(data: bytes) -> dict[str, int | str] | None:
    if len(data) < 28 or data[:2] != b"MZ":
        return None
    fields = struct.unpack_from("<14H", data, 0)
    return {
        "magic": "MZ",
        "bytes_in_last_page": fields[1],
        "pages": fields[2],
        "relocations": fields[3],
        "header_paragraphs": fields[4],
        "min_extra_paragraphs": fields[5],
        "max_extra_paragraphs": fields[6],
        "initial_ss": fields[7],
        "initial_sp": fields[8],
        "checksum": fields[9],
        "initial_ip": fields[10],
        "initial_cs": fields[11],
        "relocation_table_offset": fields[12],
        "overlay_number": fields[13],
    }


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("root", type=Path)
    parser.add_argument("output", type=Path)
    args = parser.parse_args()

    result: dict[str, object] = {
        "schema": "kol-original-inventory-v1",
        "inputs": [],
        "zip": {},
    }
    for name in INPUTS:
        path = args.root / name
        data = path.read_bytes()
        result["inputs"].append(
            {"name": name, "bytes": len(data), "sha256": sha256(data)}
        )

    archive = args.root / INPUTS[0]
    entries: list[dict[str, object]] = []
    with zipfile.ZipFile(archive) as zf:
        bad = zf.testzip()
        if bad is not None:
            raise SystemExit(f"ZIP CRC failure: {bad}")
        for info in zf.infolist():
            data = b"" if info.is_dir() else zf.read(info)
            entry: dict[str, object] = {
                "name": info.filename,
                "directory": info.is_dir(),
                "bytes": info.file_size,
                "compressed_bytes": info.compress_size,
                "compression_method": info.compress_type,
                "crc32": f"{info.CRC:08x}",
                "dos_timestamp": "%04d-%02d-%02dT%02d:%02d:%02d" % info.date_time,
                "sha256": None if info.is_dir() else sha256(data),
            }
            header = mz_header(data)
            if header is not None:
                entry["mz_header"] = header
            entries.append(entry)
        result["zip"] = {
            "entry_count": len(entries),
            "comment_hex": zf.comment.hex(),
            "crc_check": "pass",
            "entries": entries,
        }

    args.output.write_text(
        json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )


if __name__ == "__main__":
    main()
