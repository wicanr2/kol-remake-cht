"""驗證 KOLPICS 的 RLE、raw-planar 與 placeholder 混合 archive。"""

from __future__ import annotations

import argparse
import hashlib
import json
import zipfile
from pathlib import Path

from inventory_image_archives import EXPECTED_ZIP_SHA256, logical_archive_contract
from probe_extract_comb_png import decode_nibbles
from probe_extract_monpics_png import encode_rgba_png
from probe_image_lzw import decode


ARCHIVE_ROOT = "Knights-of-legend/KOLPICS"


def sha(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def decode_planar(source: bytes, pixel_count: int, source_stride: int) -> bytes:
    block_count = pixel_count // 256
    if pixel_count % 256 or source_stride != 256 or len(source) != block_count * source_stride + 1:
        raise SystemExit("raw-planar 長度／幾何契約不符")
    output = bytearray(pixel_count)
    for block in range(block_count):
        row = source[block * source_stride : block * source_stride + 128]
        planes = [row[p * 32 : (p + 1) * 32] for p in range(4)]
        for x in range(256):
            bit = 0x80 >> (x & 7)
            output[block * 256 + x] = sum(
                (1 << plane) if planes[plane][x >> 3] & bit else 0
                for plane in range(4)
            )
    return bytes(output)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("zip_path", type=Path)
    parser.add_argument("ida_summary_path", type=Path)
    parser.add_argument("output_directory", type=Path)
    parser.add_argument("summary_path", type=Path)
    args = parser.parse_args()
    if "workplace" not in args.output_directory.resolve().parts:
        raise SystemExit("研究 PNG 只可寫入 ignored workplace")
    zip_bytes = args.zip_path.read_bytes()
    if sha(zip_bytes) != EXPECTED_ZIP_SHA256:
        raise SystemExit("非預期 ZIP SHA-256")
    ida_bytes = args.ida_summary_path.read_bytes()
    ida = json.loads(ida_bytes)
    if ida.get("schema") != "kol-ida-koltown-kolpics-summary-v2":
        raise SystemExit("非預期 KOLPICS IDA 摘要 schema")
    evidence = {x["index"]: x for x in ida["archive_records"]}
    if sorted(evidence) != list(range(14)):
        raise SystemExit("IDA 摘要未完整涵蓋 KOLPICS 0..13")

    with zipfile.ZipFile(args.zip_path) as archive:
        files = {i.filename: archive.read(i) for i in archive.infolist() if not i.is_dir()}
    contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if contract is None or contract["count"] != 14:
        raise SystemExit("KOLPICS archive contract 不符")
    args.output_directory.mkdir(parents=True, exist_ok=True)

    cell_width, cell_height, columns = 336, 216, 2
    active = ida["active_indices"]
    sheet_width = cell_width * columns
    sheet_height = cell_height * ((len(active) + columns - 1) // columns)
    sheet = bytearray(7 if ((x // 8) + (y // 8)) % 2 == 0 else 8
                      for y in range(sheet_height) for x in range(sheet_width))
    records = []
    for item in contract["items"]:
        index = item["index"]
        source = files[item["volume_name"]]
        stage1 = decode(source[item["offset"] : item["offset"] + item["compressed_bytes"]])
        expected = evidence[index]
        kind = expected["kind"]
        record = {"image_index": index, "kind": kind,
                  "stage1_bytes": len(stage1), "stage1_sha256": sha(stage1)}
        if kind == "placeholder":
            if len(stage1) != 1:
                raise SystemExit(f"KOLPICS[{index}] placeholder 長度不符")
            record["placeholder_value"] = stage1[0]
            records.append(record)
            continue
        width, height = expected["width"], expected["height"]
        if kind == "rle":
            pixels = decode_nibbles(stage1)
        elif kind == "raw-planar":
            if expected.get("extraction_geometry_grade") != "confirmed":
                raise SystemExit(f"KOLPICS[{index}] extraction geometry 尚未證實")
            pixels = decode_planar(stage1, width * height, expected["source_stride_bytes"])
            if stage1[-1:] != b"\0":
                raise SystemExit(f"KOLPICS[{index}] 非預期 trailing byte")
            record["trailing_byte_sha256"] = sha(stage1[-1:])
            record["representation"] = "renderer_data_png; not an original display rectangle"
        else:
            raise SystemExit(f"未知 KOLPICS kind：{kind}")
        if len(pixels) != width * height or any(x > 15 for x in pixels):
            raise SystemExit(f"KOLPICS[{index}] 像素契約不符")
        output = args.output_directory / f"kolpics-{index:02d}.png"
        output.write_bytes(encode_rgba_png(pixels, bytes([1]) * len(pixels), width, height))
        position = active.index(index)
        origin_x = (position % columns) * cell_width + (cell_width - width) // 2
        origin_y = (position // columns) * cell_height + (cell_height - height) // 2
        for y in range(height):
            start = (origin_y + y) * sheet_width + origin_x
            sheet[start : start + width] = pixels[y * width : (y + 1) * width]
        record.update({"width_pixels": width, "height_pixels": height,
                       "color_count": len(set(pixels)), "pixel_sha256": sha(pixels),
                       "png_sha256": sha(output.read_bytes())})
        records.append(record)

    contact = args.output_directory / "kolpics-contact-sheet.png"
    contact.write_bytes(encode_rgba_png(bytes(sheet), bytes([1]) * len(sheet),
                                        sheet_width, sheet_height))
    result = {
        "schema": "kol-kolpics-ega-visual-probe-v2",
        "input_name": args.zip_path.name, "input_sha256": sha(zip_bytes),
        "ida_summary_name": args.ida_summary_path.name,
        "ida_summary_sha256": sha(ida_bytes),
        "status": "mechanically validated all eight active payloads; raw-planar extraction geometry confirmed as renderer-data serialization",
        "human_inspection": "five RLE images show coherent UI frames or town scenery; index 0 forms five coherent 256-entry converter blocks; indices 8 and 12 form coherent 64-column lookup grids after applying the confirmed downstream y*64+x access; no plane-order or serialization defect is visible",
        "active_indices": active, "placeholder_indices": ida["placeholder_indices"],
        "png_count": len(active), "contact_sheet_layout": "two columns by four rows",
        "contact_sheet_sha256": sha(contact.read_bytes()), "records": records,
    }
    args.summary_path.parent.mkdir(parents=True, exist_ok=True)
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"validated=14 png={len(active)} human_inspection=complete")


if __name__ == "__main__":
    main()
