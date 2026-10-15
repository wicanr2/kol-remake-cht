from pathlib import Path
from PIL import Image, ImageChops, ImageDraw
import hashlib, json, html

w = Path('/out')
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
manifest = json.loads((w/'complete-creature-prototype.json').read_text())
receipt = json.loads((w/'full-gallery-v1.json').read_text())
rows = {(r['source']['archive'], r['source']['index']): r for r in manifest['portraits']}
preserved = {(r['archive'], r['index']) for r in manifest['preserved']}
groups = {}
for frame in receipt['frames']:
    assert sha(w/'full-gallery-v1'/frame['png']) == frame['sha256']
    groups.setdefault(frame['id'], {})[frame['variant']] = frame
assert len(groups) == 110 and len(receipt['frames']) == 440
records = []
for name, frames in groups.items():
    assert set(frames) == {'baseline-ega', 'baseline-hd', 'prototype-ega', 'prototype-hd'}
    base = frames['baseline-ega']
    assert all(f['inn_sha256'] == base['inn_sha256'] and f['text'] == base['text'] and f['han_characters'] > 0 for f in frames.values()), name
    imgs = {k: Image.open(w/'full-gallery-v1'/f['png']).convert('RGBA') for k, f in frames.items()}
    assert imgs['baseline-ega'].size == (960, 600) and imgs['prototype-hd'].size == (1920, 1200)
    assert imgs['baseline-ega'].tobytes() == imgs['prototype-ega'].tobytes(), name
    diff = ImageChops.difference(imgs['baseline-hd'], imgs['prototype-hd'])
    assert diff.getextrema()[3] == (0, 0), name
    key = (base['archive'], base['index'])
    record = dict(id=name, kind=base['kind'], archive=key[0], index=key[1], ega_rgba_equal=True,
                  full_inn_state_equal=True, chinese_draw_log_equal=True, han_characters=base['han_characters'],
                  native_ui=base['native_ui'], alpha_difference=0)
    if key in preserved:
        value = max(v[1] for v in diff.getextrema())
        assert value <= 1, (name, 'preserved', value)
        record.update(preserved=True, complete_hd_max=value)
    else:
        row = rows[key]
        rect = {'npc': (62, 398, 560, 754), 'stable': (672, 516, 912, 744), 'sheet': (1584, 12, 1824, 240)}[base['kind']]
        width, height = rect[2]-rect[0], rect[3]-rect[1]
        source = Image.open(w/row['file']).convert('RGBA').crop(row['rect'])
        ew, eh = width, height
        if row['kind'] == 'mount':
            if source.width*height > source.height*width:
                eh = max(1, source.height*width//source.width)
            else:
                ew = max(1, source.width*height//source.height)
        expected = Image.new('RGBA', (width, height), (0, 0, 0, 255))
        sample = source.transform((ew, eh), Image.Transform.AFFINE,
                                  (source.width/ew, 0, 0, 0, source.height/eh, 0), Image.Resampling.BILINEAR)
        expected.paste(sample, ((width-ew)//2, (height-eh)//2))
        actual = imgs['prototype-hd'].crop(rect)
        native_diff = ImageChops.difference(expected, actual)
        native_max = max(v[1] for v in native_diff.getextrema())
        assert native_max <= 1 and native_diff.getextrema()[3] == (0, 0), (name, 'native', native_max)
        outside = diff.copy()
        outside.paste((0, 0, 0, 0), rect)
        outside_max = max(v[1] for v in outside.getextrema())
        assert outside_max <= 1, (name, 'outside', outside_max)
        wrong_row = next(r for k, r in rows.items() if r['kind'] == row['kind'] and k != key)
        wrong = Image.open(w/wrong_row['file']).convert('RGBA').crop(wrong_row['rect']).resize((width, height))
        assert max(v[1] for v in ImageChops.difference(wrong, actual).getextrema()) > 10, name
        record.update(preserved=False, rectangle=rect, native_fit=[ew, eh], native_max=native_max,
                      outside_max=outside_max, wrong_identity_negative_rejected=True)
    records.append(record)

out = dict(status='PASS controlled actual-game complete NPC and mount pixel checks; visual review separate',
           method='Actual game.render/game.presentHD; independent Pillow9.4 affine bilinear pixel-center oracle',
           scenes=110, frames=440, npc_scenes=102, stable_scenes=4, sheet_scenes=4, source_identities=106,
           new_npcs=98, preserved_npcs=4, mount_artworks=4, max_channel_difference=1,
           source_manifest_sha256=sha(w/'complete-creature-prototype.json'),
           capture_receipt_sha256=sha(w/'full-gallery-v1.json'),
           original_oracle=False, normal_player_path=False, gui_F10=False, v3=False,
           rng='Static building entries and prepared mount pages; gameplay RND not invoked', records=records)
(w/'full-gallery-check.json').write_text(json.dumps(out, ensure_ascii=False, indent=2)+'\n')
parts = ['<!doctype html><meta charset="utf-8"><title>七城 NPC 與坐騎遊戲畫面</title>',
         '<style>body{background:#20171b;color:#eeddb9;font:18px sans-serif;margin:24px}section{margin:28px 0}.pair{display:grid;grid-template-columns:1fr 1fr;gap:12px}img{width:100%;height:auto}a{color:#e4bf79}</style>',
         '<h1>七城 NPC 與坐騎：實際遊戲 EGA／HD</h1><p>110 個受控場景，包含 102 位 NPC、四種馬廄及四種人物資料頁。完整組仍為隔離原型，尚未正式接入。</p>',
         '<p><a href="full-gallery-check.json">像素檢查</a> · <a href="full-gallery-v1.json">完整場景與中文收據</a> · <a href="complete-creature-prototype.json">來源與原生裁切</a></p>']
for name, fs in groups.items():
    title = html.escape(' / '.join(fs['prototype-hd']['text'][1:3]))
    parts.append(f'<section><h2>{html.escape(name)}：{title}</h2><div class="pair"><div>EGA<img loading="lazy" src="full-gallery-v1/{fs["prototype-ega"]["png"]}"></div><div>HD<img loading="lazy" src="full-gallery-v1/{fs["prototype-hd"]["png"]}"></div></div></section>')
(w/'full-runtime-review.html').write_text('\n'.join(parts)+'\n')
selected = [f'npc-t{i}-p02' for i in range(1, 8)] + ['npc-t5-p22', 'npc-t5-p28', 'npc-t6-p10', 'npc-t7-p10', 'mount-stable-3', 'mount-sheet-3']
for i, name in enumerate(selected):
    frame = groups[name]['prototype-hd']
    original = Image.open(w/'full-gallery-v1'/frame['png']).convert('RGB')
    if i < 11:
        original = original.crop((48, 348, 1780, 912))
    original.thumbnail((866, 540), Image.Resampling.BILINEAR)
    original.save(w/(name+'-visual.png'))
print(json.dumps({k: v for k, v in out.items() if k != 'records'}, ensure_ascii=False))
