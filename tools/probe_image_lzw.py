"""一次性驗證 KOLTOWN 圖片 LZW 流；只輸出雜湊與統計，不輸出素材。"""

from __future__ import annotations

import argparse
import hashlib
import json
import zipfile
from pathlib import Path

from inventory_image_archives import (
    EXPECTED_ZIP_SHA256,
    is_candidate,
    logical_archive_contract,
)


def decode(source: bytes) -> bytes:
    # sub_1A26B 在 byte 邊界會預讀下一 byte；本地零值只承接最後 padding bits。
    padded = source + b"\0\0"
    bit_position = 0
    code_width = 9
    dictionary = {value: bytes([value]) for value in range(256)}
    next_code = 0x101

    def read_code(width: int) -> int:
        nonlocal bit_position
        value = 0
        for bit_index in range(width):
            source_bit = bit_position + bit_index
            value |= (
                (padded[source_bit // 8] >> (source_bit % 8)) & 1
            ) << bit_index
        bit_position += width
        return value

    previous_code = read_code(code_width)
    if previous_code > 0xFF:
        raise ValueError(f"首碼超出 literal 範圍：{previous_code:#x}")
    previous = dictionary[previous_code]
    output = bytearray(previous)

    while bit_position // 8 < len(source):
        code = read_code(code_width)
        if code == 0x100:
            code_width += 1
            if code_width > 16:
                raise ValueError("碼寬超過原函式的 16-bit 上限")
            code = read_code(code_width)

        if code < next_code:
            entry = dictionary[code]
        elif code == next_code:
            entry = previous + previous[:1]
        else:
            raise ValueError(
                f"無效字典碼：code={code:#x}, next={next_code:#x}, width={code_width}"
            )
        output.extend(entry)
        dictionary[next_code] = previous + entry[:1]
        next_code += 1
        previous = entry

    return bytes(output)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("zip_path", type=Path)
    parser.add_argument("output_path", type=Path)
    args = parser.parse_args()

    zip_bytes = args.zip_path.read_bytes()
    digest = hashlib.sha256(zip_bytes).hexdigest()
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit(f"非預期 ZIP SHA-256：{digest}")

    records = []
    with zipfile.ZipFile(args.zip_path) as archive:
        files = {
            info.filename: archive.read(info)
            for info in archive.infolist()
            if not info.is_dir()
        }
        for name in sorted(files, key=str.upper):
            if not name.upper().endswith(".0") or not is_candidate(name):
                continue
            contract = logical_archive_contract(name[:-2], files)
            if contract is None or not contract["all_payload_bounds_valid"]:
                continue
            outputs = []
            for item in contract["items"]:
                source = files[item["volume_name"]]
                offset = item["offset"]
                length = item["compressed_bytes"]
                decoded = decode(source[offset : offset + length])
                outputs.append(
                    {
                        "index": item["index"],
                        "volume_index": item["volume_index"],
                        "volume_name": item["volume_name"],
                        "compressed_offset": offset,
                        "compressed_bytes": length,
                        "decoded_bytes": len(decoded),
                        "decoded_sha256": hashlib.sha256(decoded).hexdigest(),
                    }
                )
            records.append({"name": contract["root_name"], "outputs": outputs})

    result = {
        "schema": "kol-image-lzw-probe-summary-v2",
        "input_name": args.zip_path.name,
        "input_sha256": digest,
        "status": "mechanical decode only; dimensions, pixels and palette not validated",
        "archive_count": len(records),
        "image_count": sum(len(record["outputs"]) for record in records),
        "records": records,
    }
    args.output_path.parent.mkdir(parents=True, exist_ok=True)
    args.output_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
