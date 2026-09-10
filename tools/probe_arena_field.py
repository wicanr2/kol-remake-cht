"""競技場地圖候選；只輸出到本機研究工作目錄，不接正式資產。"""
import hashlib
import json
from pathlib import Path
import zipfile
from extract_battle_tiles import rle_decode
from probe_extract_town_map_png import encode_gray_png
from probe_town_data import lzw_decode, payloads, read_volumes
from inventory_image_archives import EXPECTED_ZIP_SHA256

source = Path('/src/Knights-of-legend.zip')
assert hashlib.sha256(source.read_bytes()).hexdigest() == EXPECTED_ZIP_SHA256
with zipfile.ZipFile(source) as archive:
    contract, parts = payloads(read_volumes(archive, 'TOWN0'), 'TOWN0')
    assert len(parts) > 0x55
    raw = lzw_decode(parts[0x55])
cells = rle_decode(raw)
assert all(x < 80 for x in cells)
png = encode_gray_png(cells, 64, 64)
out = Path('/out')
assert out.stat().st_uid == 1000
(out/'arena-map.png').write_bytes(png)
(out/'arena-map.json').write_text(json.dumps(dict(schema='kol-arena-map-probe-v1',
    source_zip_sha256=EXPECTED_ZIP_SHA256, archive='TOWN0', archive_count=len(parts),
    payload=0x55, raw_hex=raw.hex(), raw_sha256=hashlib.sha256(raw).hexdigest(),
    cells_sha256=hashlib.sha256(cells).hexdigest(), png_sha256=hashlib.sha256(png).hexdigest(),
    rights='僅本機研究；尚未視覺驗證'), ensure_ascii=False, indent=2), encoding='utf-8')
print(json.dumps(dict(payload=0x55,archive_count=len(parts),raw_bytes=len(raw),cells=len(cells),tiles=sorted(set(cells)))))
