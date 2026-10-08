import collections
import hashlib
import json
import pathlib
import sys
sys.path.insert(0, '/src/tools')
from build_asset_contact_sheets import read_image
from png_io import write_png_rgb

base = pathlib.Path('assets/original-derived-png/ega')
out = pathlib.Path('workplace/cjk-edge-20261001')

def white_mask(img):
    rgb = img.rgb()
    return bytes(tuple(rgb[i:i+3]) != (255, 255, 255) for i in range(0, len(rgb), 3))

references = []
for town, tile in [(3, 39), (5, 144), (7, 12)]:
    ref = read_image(base / f'TOWN{town}/town{town}-tiles.png').crop(tile % 16 * 16, tile // 16 * 16, 16, 16)
    path = out / f'free-reference-town{town}-tile{tile}.png'
    write_png_rgb(path, ref.width, ref.height, ref.rgb())
    references.append({'town': town, 'tile': tile, 'mask': white_mask(ref),
                       'file': path.name, 'sha256': hashlib.sha256(path.read_bytes()).hexdigest()})
result = []
for town in range(1, 8):
    atlas_path = base / f'TOWN{town}/town{town}-tiles.png'
    map_path = base / f'TOWN{town}/town{town}-map.png'
    atlas, grid = read_image(atlas_path), read_image(map_path)
    counts = collections.Counter(grid.rgb()[::3])
    matches = [i for i in range(256) if white_mask(atlas.crop(i % 16 * 16, i // 16 * 16, 16, 16)) in [r['mask'] for r in references]]
    result.append({'town': town, 'matching_tiles': matches,
                   'map_references': {str(i): counts[i] for i in matches},
                   'atlas_sha256': hashlib.sha256(atlas_path.read_bytes()).hexdigest(),
                   'map_sha256': hashlib.sha256(map_path.read_bytes()).hexdigest()})
report = {'schema': 'kol-baked-text-free-audit-v1', 'references': [{k:v for k,v in r.items() if k != 'mask'} for r in references],
          'method': '逐 16×16 圖塊比較白底／非白字形遮罩；不使用 OCR 推測，逐地圖灰階索引計數',
          'towns': result}
(out / 'free-tile-audit.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(result, ensure_ascii=False))
