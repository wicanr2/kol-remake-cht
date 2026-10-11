#!/usr/bin/env python3
"""Docker內核對核准PNG並重生SPEC-039首批素材清冊；不重繪或改尺寸。"""
import hashlib
import json
from pathlib import Path
from PIL import Image

ROOT = Path('/src')
assert Path('/.dockerenv').exists()
approved = ROOT / 'docs/verification/hd-first-slice-20261011'
dest = ROOT / 'assets/hd/first-slice'
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()


def rgb_key(pic, x, y, w, h, scale=3):
    rgb = pic.convert('RGB')
    pixels = bytes(v for row in range(h) for col in range(w)
                   for v in rgb.getpixel((x + col*scale, y + row*scale)))
    return hashlib.sha256(pixels).hexdigest()


reference = {
    'docs/screenshots/zh-town-door.png': Image.open(ROOT / 'docs/screenshots/zh-town-door.png'),
    'docs/verification/cjk-closeout-20261011/zh-party.png':
        Image.open(ROOT / 'docs/verification/cjk-closeout-20261011/zh-party.png'),
}
files = {}
for name, size in [('inn-painted-v1.png', (1774, 887)), ('party-painted-v1.png', (1536, 1024))]:
    p = dest / name
    assert p.is_file() and sha(p) == sha(approved / name), name
    assert Image.open(p).size == size, name
    files[name] = dict(sha256=sha(p), width=size[0], height=size[1],
                       source_entry=str((approved / name).relative_to(ROOT)),
                       evidence='使用者2026-10-11核准原型；原樣接入',
                       rights='private GitHub 可保存；無公開散布權')
party = reference['docs/verification/cjk-closeout-20261011/zh-party.png']
keys = [rgb_key(party, 48+i*144, 6, 40, 38) for i in range(6)]
assert len(set(keys)) == 6
manifest = dict(schema='kol-hd-first-slice/1', rights='private GitHub 可保存；無公開散布權',
    source_zip_sha256=json.loads((approved/'manifest.json').read_text())['source_zip_sha256'],
    decision='2026-10-11：同意囉，開始吧；首批原型接入，其他批次未核准',
    files=files, facade_rgb_sha256=rgb_key(reference['docs/screenshots/zh-town-door.png'], 312,174,192,96),
    portrait_rgb_sha256=keys, sources={p:sha(ROOT/p) for p in reference},
    approved_manifest_sha256=sha(approved/'manifest.json'),
    prompts_sha256=sha(approved/'prompts.json'),
    tool='tools/build_hd_first_slice.py / Pillow；RGB逐像素指紋，不轉檔',
    matching='依960×600畫格中的3倍原圖逐像素取樣；角色名與隊伍位置不作runtime鍵',
    resolution=[1920,1200], spec='docs/spec/039-hd-art-ui.md#首批接入契約')
(dest/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
print('PASS approved PNGs unchanged;',len(keys),'unique portrait fingerprints')
