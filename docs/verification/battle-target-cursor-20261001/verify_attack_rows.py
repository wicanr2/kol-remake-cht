"""在 Docker 中核對完整攻擊／防禦列；--export 保存已驗PNG與收據。"""
from pathlib import Path
import hashlib
import json
import os
import shutil
import sys
import zipfile

from attack_rows_pixels import ROOT, OUT, ASSETS, WEAPON, GUARD, expected, difference
sys.path.insert(0, '/src/tools')
from png_io import read_png

dest = ROOT / 'docs/verification/battle-target-cursor-20261001'
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
assert dest.stat().st_uid == OUT.stat().st_uid == os.getuid()
base = json.loads((dest / 'remake-receipt.json').read_text())
changed_sources = [r['path'] for r in base['inputs'] if sha(ROOT / r['path']) != r['sha256']]
assert changed_sources == ['internal/ui/battle_cjk.go'], changed_sources
before_source = OUT / 'battle_cjk-before-attack-frame.go'
before_row = next(r for r in base['inputs'] if r['path'] == 'internal/ui/battle_cjk.go')
assert sha(before_source) == before_row['sha256']

manifest_path = ROOT / 'assets/original-derived-png/ega/manifest.json'
manifest = json.loads(manifest_path.read_text())
archive = next(r for r in manifest['archives'] if r['archive'] == 'KOLMPICS')
assert archive['source_payload_count'] == 83 and archive['png_count'] == 71
needed = {0} | {c+3 for c in WEAPON + GUARD}
assets = []
for index in sorted(needed):
    row = next(r for r in archive['images'] if r['unit_index'] == index)
    p = ASSETS / f'kolmpics-{index:03d}.png'
    assert sha(p) == row['png_sha256']
    assert index < archive['source_payload_count']
    im = read_png(p)
    assert (im.width, im.height) == ((320, 200) if index == 0 else (24, 20))
    assets.append(row)

original_inputs = json.loads((OUT/'gui11/inputs.json').read_text())
assert original_inputs['seed'] == 1 and original_inputs['sources_unchanged']
assert not original_inputs['direct_entry'] and not original_inputs['state_injection']
assert original_inputs['mouse_event'] == dict(api='dosgolem.Oracle.MoveMouse', x=620, y=4, buttons=0)
normal = json.loads((OUT/'attack-row-fixed-normal-receipt.json').read_text())
assert normal['inputs_each'] == 438 and normal['states_each'] == 439
assert normal['equal_language_states'] and not normal['direct_entry'] and not normal['state_injection']
binary = OUT/'kol-attack-frame-linux'
assert normal['binary_sha256'] == sha(binary)
logs = (OUT/'attack-frame-tests.log').read_text()
assert all('ok  \tgithub.com/wicanr2/kol-remake-cht/'+p in logs for p in ['internal/ui','cmd/kol'])
for lang in ['en','zh-Hant']:
    before = json.loads((OUT/('attack-row-'+lang)/'trace.json').read_text())['steps']
    after = json.loads((OUT/('attack-row-fixed-'+lang)/'trace.json').read_text())['steps']
    assert before == after and len(after) == 439

checks, copies = [], []
for n in range(6, 23):
    codes = WEAPON if n < 14 else GUARD
    selected = (n-6)%6 if n < 14 else (n-14)%8
    source = OUT/'gui11'/f'plan-{n:02d}.png'
    meta = json.loads(source.with_suffix('.png.state.json').read_text())
    assert meta['icon_selected'] == selected and meta['icon_slots'] == len(codes)
    want = expected(codes, selected)
    assert len(want) == 296*32*3
    original_diff, _ = difference(source, want, 1)
    assert original_diff == 0, (n, '原版', original_diff)
    copies.append((source, f'attack-original-plan-{n:02d}.png'))
    check = dict(original_step=n, remake_step=n+416, codes=codes, selected=selected,
        original_diff=original_diff, expected_rgb_sha256=hashlib.sha256(want).hexdigest(),
        original_sha256=sha(source), original_state_private_sha256=sha(source.with_suffix('.png.state.json')))
    for lang, scale in [('en',1),('zh-Hant',3)]:
        before = OUT/('attack-row-'+lang)/f'step-{n+416}.png'
        after = OUT/('attack-row-fixed-'+lang)/f'step-{n+416}.png'
        before_diff, _ = difference(before, want, scale)
        after_diff, _ = difference(after, want, scale)
        assert before_diff == (0 if lang == 'en' else 520*len(codes))
        assert after_diff == 0, (n, lang, after_diff)
        a, b = read_png(before), read_png(after)
        assert (a.width,a.height) == (b.width,b.height)
        av,bv = a.rgb(),b.rgb()
        changed = 0
        for i in range(0,len(av),3):
            if av[i:i+3] == bv[i:i+3]:
                continue
            x,y = (i//3)%a.width,(i//3)//a.width
            # 只允許原列各框內緣的等比線條，不把整個戰場當例外。
            allowed = any(32*scale+slot*32*scale <= x < (56+slot*32)*scale and
                          170*scale <= y < 192*scale and
                          (x < (33+slot*32)*scale or x >= (55+slot*32)*scale or
                           y < 171*scale or y >= 191*scale) for slot in range(len(codes)))
            assert lang == 'zh-Hant' and allowed, (n,lang,x,y)
            changed += 1
        assert changed == before_diff
        check[lang] = dict(before_sha256=sha(before),after_sha256=sha(after),
            before_diff=before_diff,after_diff=after_diff,whole_frame_changed=changed,
            changed_outside_icon_border=0)
        copies.append((after, f'attack-{lang}-step-{n+416}.png'))
    bad = bytearray(want)
    bad[0] ^= 1
    assert difference(source, bytes(bad), 1)[0] == 1
    assert difference(source, expected(codes,(selected+1)%len(codes)),1)[0] > 0
    assert difference(source, expected(codes[:-1],selected),1)[0] > 0
    checks.append(check)

zip_path = ROOT/'Knights-of-legend.zip'
assert sha(zip_path) == manifest['source_zip']['sha256']
with zipfile.ZipFile(zip_path) as z:
    archive_entries = [dict(entry=name,sha256=hashlib.sha256(z.read(name)).hexdigest())
        for name in z.namelist() if name.startswith(archive['archive_root']) and not name.endswith('/')]
assert archive_entries
life = json.loads((dest/'lifecycle-receipt.json').read_text())
receipt = dict(schema='kol-attack-row-v1',date='2026-10-01',status='CONFORMED，限完整武器／防禦圖示列與列出的選取狀態',
    binary=dict(path='workplace/battle-cursor-20261001/kol-attack-frame-linux',sha256=sha(binary)),
    before_binary=base['binary'],before_source_snapshot=dict(path='workplace/battle-cursor-20261001/battle_cjk-before-attack-frame.go',sha256=sha(before_source)),
    source_zip=dict(path='Knights-of-legend.zip',sha256=sha(zip_path)),source_archive_entries=archive_entries,
    original_tools=life['original_tools'],executable_sources=life['executable_sources'],
    dosgolem_revision=original_inputs['dosgolem_revision'],original_seed=1,remake_seed=1,
    mouse_event=original_inputs['mouse_event'],original_keys=original_inputs['planning_keys'],
    original_inputs_private=dict(path='workplace/battle-cursor-20261001/gui11/inputs.json',sha256=sha(OUT/'gui11/inputs.json')),
    assets_manifest=dict(path=str(manifest_path.relative_to(ROOT)),sha256=sha(manifest_path)),
    archive_count=83,validated_png_count=71,assets=assets,
    product_inputs=[dict(path=r['path'],sha256=sha(ROOT/r['path'])) for r in base['inputs']],
    rectangle=[8,164,296,32],logical_pixels_each=9472,checks=checks,
    summary=dict(original_frames=17,normal_frames=34,normal_runs=2,normal_inputs_each=438,
        normal_states_each=439,different_pixels=0,changed_outside_icon_border=0,
        english_whole_frames_unchanged=17,weapon_selections=6,guard_selections=8,
        right_wrap_verified=True,one_bit_wrong_selection_missing_icon_rejected=True,
        target_marker_scaled_regression_boxes=104),
    tests=dict(path='attack-frame-tests.log',sha256=sha(OUT/'attack-frame-tests.log'),
        command='KOL_MARKER_ORACLE=1 go test ./internal/ui ./cmd/kol -count=1；go build -o /out/kol-attack-frame-linux ./cmd/kol'),
    limits=['圖示列狀態及倍率契約驗收；兩側隊伍、遭遇與目標格不同，不稱整場戰鬥同狀態',
            '原版沒有繁中畫面；繁中驗既定三倍呈現，文字不冒稱原版oracle',
            '其他等待層、法術／戰利品／物品頁與正式平台交付另驗'],
    rights='已驗PNG僅依private GitHub既有例外；原版與payload仍僅本機，不授權公開。')

if '--export' in sys.argv:
    copies += [(OUT/name,name) for name in ['attack-kol_cursor_gui.sh','attack-kol_cursor_gui_prepare.py',
              'attack-kol_cursor_gui_finish.py','attack_normal_replay.py','attack_after_replay.py',
              'attack-frame-tests.log','attack-row-fixed-normal-receipt.json']]
    attachments = []
    for p,name in copies:
        q=dest/name
        assert p.stat().st_uid==os.getuid()
        assert not q.exists(), q
        shutil.copyfile(p,q)
        attachments.append(dict(file=name,sha256=sha(q),uid=q.stat().st_uid,gid=q.stat().st_gid))
    receipt['attachments']=attachments
    receipt['verifiers']=[dict(file=p.name,sha256=sha(p)) for p in [Path(__file__),dest/'attack_rows_pixels.py']]
    q=dest/'attack-row-receipt.json'
    assert not q.exists()
    q.write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
else:
    saved=json.loads((dest/'attack-row-receipt.json').read_text())
    assert saved['checks']==checks
    for row in saved['attachments']:
        p=dest/row['file']
        assert sha(p)==row['sha256'] and p.stat().st_uid==row['uid']==os.getuid()
        assert p.stat().st_gid==row['gid']==os.getgid()
    for row in saved['verifiers']:
        assert sha(dest/row['file'])==row['sha256']
print(json.dumps(receipt['summary'],ensure_ascii=False))
