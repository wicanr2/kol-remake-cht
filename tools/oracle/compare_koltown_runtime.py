#!/usr/bin/env python3
"""比較固定 derivative 與 DOSBox-X runtime load image，並正規化 MZ relocation。"""

from __future__ import annotations

import hashlib
import json
import struct
import sys
from pathlib import Path


DERIVATIVE_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
ANCHORS = [0x00000, 0x08000, 0x10000, 0x18000, 0x1DED8]
ANCHOR_LENGTH = 32


def sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def ranges(offsets: list[int]) -> list[tuple[int, int]]:
    result: list[list[int]] = []
    for offset in offsets:
        if not result or offset != result[-1][1] + 1:
            result.append([offset, offset])
        else:
            result[-1][1] = offset
    return [(start, end) for start, end in result]


def main() -> None:
    if len(sys.argv) != 6:
        raise SystemExit(
            "用法：compare_koltown_runtime.py DERIVATIVE.EXE RUNTIME.bin RECEIPT.json "
            "PRIVATE.json SUMMARY.json"
        )
    derivative_path, runtime_path, receipt_path, private_path, summary_path = map(
        Path, sys.argv[1:]
    )
    derivative = derivative_path.read_bytes()
    if sha256(derivative) != DERIVATIVE_SHA256:
        raise RuntimeError("derivative SHA-256 不符")
    runtime = bytearray(runtime_path.read_bytes())
    receipt = json.loads(receipt_path.read_text(encoding="utf-8"))
    if sha256(runtime) != receipt["load_image_sha256"]:
        raise RuntimeError("runtime dump SHA-256 與收據不符")

    header_size = struct.unpack_from("<H", derivative, 0x08)[0] * 16
    relocation_count = struct.unpack_from("<H", derivative, 0x06)[0]
    relocation_offset = struct.unpack_from("<H", derivative, 0x18)[0]
    load_image = derivative[header_size:]
    if len(load_image) != len(runtime):
        raise RuntimeError(f"load image 長度不符：{len(load_image)} != {len(runtime)}")
    load_segment = int(receipt["load_segment"], 16)

    relocation_mismatches = []
    for index in range(relocation_count):
        offset, segment = struct.unpack_from(
            "<HH", derivative, relocation_offset + index * 4
        )
        linear = segment * 16 + offset
        if linear + 1 >= len(runtime):
            raise RuntimeError(f"relocation {index} 超出 load image：{linear:#x}")
        derivative_word = struct.unpack_from("<H", load_image, linear)[0]
        runtime_word = struct.unpack_from("<H", runtime, linear)[0]
        expected = (derivative_word + load_segment) & 0xFFFF
        if runtime_word != expected:
            relocation_mismatches.append({
                "index": index,
                "load_offset": linear,
                "ida_ea": f"0x{0x10000 + linear:x}",
                "derivative_word": derivative_word,
                "runtime_word": runtime_word,
                "expected_runtime_word": expected,
            })
        struct.pack_into("<H", runtime, linear, (runtime_word - load_segment) & 0xFFFF)

    differing_offsets = [
        index for index, (left, right) in enumerate(zip(load_image, runtime))
        if left != right
    ]
    diff_runs = []
    for start, end in ranges(differing_offsets):
        diff_runs.append({
            "load_start": start,
            "load_end": end,
            "ida_start_ea": f"0x{0x10000 + start:x}",
            "ida_end_ea": f"0x{0x10000 + end:x}",
            "length": end - start + 1,
            "derivative_hex": load_image[start:end + 1].hex(),
            "runtime_normalized_hex": runtime[start:end + 1].hex(),
        })

    anchors = []
    for offset in ANCHORS:
        left = load_image[offset:offset + ANCHOR_LENGTH]
        right = runtime[offset:offset + ANCHOR_LENGTH]
        anchors.append({
            "load_offset": offset,
            "ida_ea": f"0x{0x10000 + offset:x}",
            "length": ANCHOR_LENGTH,
            "matched": left == right,
            "sha256": sha256(left),
        })

    summary = {
        "schema": "kol-koltown-runtime-compare-summary-v1",
        "derivative_name": derivative_path.name,
        "derivative_sha256": sha256(derivative),
        "runtime_name": runtime_path.name,
        "runtime_sha256": receipt["load_image_sha256"],
        "address_space": "load-image offsets; IDA EA = 0x10000 + load offset",
        "mz_header_bytes": header_size,
        "load_image_bytes": len(load_image),
        "runtime_load_segment": receipt["load_segment"],
        "relocation_count": relocation_count,
        "relocation_contract_match_count": relocation_count - len(relocation_mismatches),
        "relocation_contract_mismatch_count": len(relocation_mismatches),
        "normalized_diff_byte_count": len(differing_offsets),
        "normalized_diff_run_count": len(diff_runs),
        "anchors": anchors,
        "diff_runs": [
            {key: value for key, value in run.items() if not key.endswith("_hex")}
            for run in diff_runs
        ],
    }
    private = dict(summary)
    private["schema"] = "kol-koltown-runtime-compare-private-v1"
    private["relocation_mismatches"] = relocation_mismatches
    private["diff_runs"] = diff_runs

    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
