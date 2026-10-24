#!/usr/bin/env python3
"""Docker限定：比較現有中文字模與粉圓，產生不接入正式遊戲的設定原型。"""
import argparse
import hashlib
import json
import os
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--root', type=Path, required=True)
p.add_argument('--reference', type=Path, required=True)
p.add_argument('--out', type=Path, required=True)
a = p.parse_args()
assert Path('/.dockerenv').exists() and not a.out.exists()
assert a.out.parent.stat().st_uid == os.getuid()
source = a.reference / 'workplace/full-hd-v1/manual-hd-v1/rounded-font/jf-openhuninn-2.1.ttf'
sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
assert sha(source) == '9d5bf4932d31fe94c18cd8cfddc98bc1b14ce10f4e354c682179db290a99c825'
a.out.mkdir()
faces = {s: ImageFont.truetype(str(source), s * 4) for s in (16, 20, 24, 28)}


def mask(ch, size, width=None, height=None):
    width, height = width or size, height or size
    # 共用基線保留標點位置，四倍取樣後縮回固定字格。
    face = faces[size]
    asc, _ = face.getmetrics()
    tile = Image.new('L', (width * 4, height * 4))
    ImageDraw.Draw(tile).text((0, (height - size) * 2 + size * 4 - asc), ch, font=face, fill=255)
    return tile.resize((width, height), Image.Resampling.LANCZOS)


def label(im, xy, text, size=20, color='#e6ebf5'):
    x, y = xy
    face = faces[size]
    for ch in text:
        advance = round(face.getlength(ch) / 4) if ord(ch) < 128 else size
        tile = Image.new('RGBA', (max(size, advance), size), color)
        tile.putalpha(mask(ch, size, max(size, advance)))
        im.alpha_composite(tile, (x, y))
        x += advance


m = json.loads((a.root / 'assets/ui-font/manifest.json').read_text())['fonts'][0]
atlas = Image.open(a.root / 'assets/ui-font' / m['file']).convert('RGBA')
glyphs = {g['codepoint']: g for g in m['glyphs']}
samples = ['武士傳說　角色與裝備', '生命值　疲勞　金幣 9000', '攻擊部位：上段、中段、下段。', '住宿保存隊伍，重開繼續遊戲。']
comparison = Image.new('RGBA', (1120, 440), '#10151e')
label(comparison, (24, 14), '字型比較原型：保留 16×18 字格，以下以 2 倍顯示', 24)
for col, rounded in enumerate((False, True)):
    x = 24 + col * 548
    label(comparison, (x, 64), '粉圓柔邊' if rounded else '現有中文字型', 24)
    small = Image.new('RGBA', (260, 128), '#fff4d4')
    for row, text in enumerate(samples):
        pen = 0
        for ch in text:
            g = glyphs.get(ord(ch))
            advance = g['advance'] if g else (8 if ord(ch) < 128 else 16)
            if rounded and ord(ch) >= 128:
                alpha = mask(ch, 16, 16, 18)
            elif g:
                gx, gy = g['sheet_column'] * 16, g['sheet_row'] * 18
                alpha = atlas.crop((gx, gy, gx + 16, gy + 18)).getchannel('A')
            else:
                pen += advance
                continue
            tile = Image.new('RGBA', (16, 18), '#30241d')
            tile.putalpha(alpha)
            small.alpha_composite(tile, (pen, row * 30 + 4))
            pen += advance
        assert pen <= small.width, (text, pen)
    comparison.alpha_composite(small.resize((520, 256), Image.Resampling.NEAREST), (x, 106))
label(comparison, (24, 392), '中文與全形標點改用粉圓；英數字及字距沿用現值。', 20)
comparison.convert('RGB').save(a.out / 'comparison.png')

original = a.root / 'docs/verification/assist-toolbar-20261024/toolbar.png'
background = Image.open(original).convert('RGBA')
assert background.size == (960, 648)
for rounded in (False, True):
    im = background.copy()
    d = ImageDraw.Draw(im)
    d.rectangle((400, 182, 942, 600), fill='#111d27', outline='#9e9478', width=2)
    label(im, (422, 200), '畫面設定', 24)
    for box, text, active in [((420, 240, 650, 278), '美術', False), ((664, 240, 922, 278), '字型', True)]:
        d.rectangle(box, fill='#285a73' if active else '#2a3446')
        label(im, (box[0] + 16, box[1] + 8), text)
    label(im, (422, 296), '遊戲中文字型', 20)
    for i, text in enumerate(('現有字型', '粉圓柔邊')):
        y = 330 + 56 * i
        active = bool(i) == rounded
        d.rectangle((420, y, 922, y + 44), fill='#285a73' if active else '#2a3446', outline='#b5a16a' if active else '#526477')
        label(im, (438, y + 11), text + ('　使用中' if active else ''))
    label(im, (422, 454), '切換後立即預覽，並記住選擇。', 20)
    label(im, (422, 488), '工具列目前已使用粉圓。', 20)
    label(im, (422, 520), '保留原有字距與操作位置。', 20)
    d.rectangle((718, 550, 922, 586), fill='#2a3446', outline='#b5a16a')
    label(im, (740, 558), '返回遊戲 Esc', 20)
    # 此圖為設定版面原型，背景遊戲文字未替換，避免誤認正式畫面。
    im.convert('RGB').save(a.out / ('settings-rounded.png' if rounded else 'settings-current.png'))

page = '''<!doctype html><html lang="zh-Hant"><meta charset="utf-8"><title>武士傳說字型原型</title>
<style>body{background:#10151e;color:#e6ebf5;font:18px sans-serif;max-width:1120px;margin:24px auto}button{padding:10px 20px;font:inherit;margin:8px;background:#285a73;color:white;border:1px solid #526477;cursor:pointer}img{max-width:100%;display:block}p{line-height:1.6}</style>
<h1>圓潤字型設定原型</h1><p>參照 Zeliard 的 jf open 粉圓 2.1。僅為比較與設定版面原型，不會改寫遊戲或存檔。</p>
<img src="comparison.png" alt="同字距的現有中文字型與粉圓比較">
<button onclick="document.querySelector('#settings').src='settings-current.png'">現有字型</button>
<button onclick="document.querySelector('#settings').src='settings-rounded.png'">粉圓柔邊</button>
<img id="settings" src="settings-rounded.png" alt="畫面 F10 的字型設定版面原型">
<p>上圖為設定版面示意，背景遊戲文字尚未替換。正式方案：在畫面 F10 加入字型分頁；提供現有字型／粉圓柔邊，預設維持現有字型；選擇另存偏好，不改角色存檔。EGA 與 HD 都可使用；英數字、字距與選項位置保持原值。本次不增加字級設定。</p></html>'''
(a.out / 'index.html').write_text(page)
record = dict(status='PROTOTYPE_ONLY', reference='~/cht/zeliard/tools/prepare_rounded_font.py', font_sha256=sha(source), original_background=dict(path=str(original.relative_to(a.root)), sha256=sha(original)), generator_sha256=sha(Path(__file__)), outputs={f.name: sha(f) for f in a.out.iterdir() if f.is_file()}, samples=samples, production_changed=False)
(a.out / 'receipt.json').write_text(json.dumps(record, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(record, ensure_ascii=False))
