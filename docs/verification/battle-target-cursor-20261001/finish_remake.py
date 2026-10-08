"""Docker內彙整RE-171窄驗收；/src為專案、/out為本機驗證根。"""
from pathlib import Path
import json, hashlib, os, shutil, sys
sys.path.insert(0, '/src/tools')
from png_io import read_png

src, out = Path('/src'), Path('/out')
dest = src / 'docs/verification/battle-target-cursor-20261001'
assert dest.stat().st_uid == out.stat().st_uid == os.getuid()
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
life = json.loads((dest / 'lifecycle-receipt.json').read_text())
gpu = json.loads((out / 'remake-pixel-checks.json').read_text())
normal = json.loads((out / 'normal3-player-receipt.json').read_text())
assert len(life['checks']) == 52 and all(x['different_pixels'] == 0 for x in life['checks'])
assert len(gpu['checks']) == 104 and all(x['different_pixels'] == 0 for x in gpu['checks'])
assert normal['keys'] == 455 and normal['covered_prompts'] == [0, 1, 2, 3, 4, 5]
assert normal['same_states'] and not normal['direct_entry'] and not normal['state_injection']
logs = (out / 'product-tests-final.log').read_text()
assert all('ok  \tgithub.com/wicanr2/kol-remake-cht/' + p in logs
           for p in ['internal/battle', 'internal/ui', 'cmd/kol'])
annotated = json.loads((out / 'cursor-final14-annotated-private.json').read_text())
expected = set(json.loads((src / 'docs/re/target-cursor-address-index.json').read_text())['entries'])
found = {i['original_ea'] for r in annotated['records'] for i in r['instructions']
         if any(a['index'].endswith('target-cursor-address-index.json') for a in i['annotations'])}
assert found == expected and len(found) == 14
attachments, changes = [], []
def copy(p, name):
    assert p.stat().st_uid == os.getuid()
    q = dest / name
    if q.exists(): assert q.stat().st_uid == os.getuid()
    shutil.copyfile(p, q)
    attachments.append(dict(file=name, sha256=sha(q), uid=q.stat().st_uid, gid=q.stat().st_gid))
    return q

for c in gpu['checks']:
    p = out / f"remake-{c['frame']}-x{c['scale']}.png"
    im = read_png(p)
    assert (im.width, im.height) == (16*c['scale'], 16*c['scale'])
    copy(p, p.name)
for lang, scale in [('en', 1), ('zh-Hant', 3)]:
    for q in sorted((out / ('normal3-after-' + lang)).glob('*.png')):
        before = out / ('normal3-before-' + lang) / q.name
        a, b = read_png(q), read_png(before)
        assert (a.width, a.height) == (b.width, b.height)
        av, bv = a.rgb(), b.rgb()
        inside = outside = 0
        for y in range(a.height):
            for x in range(a.width):
                i = (y*a.width + x)*3
                if av[i:i+3] != bv[i:i+3]:
                    if 152*scale <= x < 296*scale and 8*scale <= y < 152*scale: inside += 1
                    else: outside += 1
        assert outside == 0, (lang, q.name, outside)
        # 初次與換人主列：目前人物尚無指令，不能沿用前人的值多畫圓環。
        if q.name in ['step-417.png', 'step-418.png']:
            assert inside == 0, (lang, q.name, '初次／換人多畫圓環', inside)
        copied = copy(q, f'normal-final-{lang}-{q.name}')
        changes.append(dict(lang=lang, file=copied.name, sha256=sha(copied),
                            before_sha256=sha(before), changed_in_field=inside, changed_outside_field=outside))
assert len(changes) == 34
for name in ['remake-pixel-checks.json', 'normal3-player-receipt.json', 'product-tests-final.log']:
    copy(out / name, name)
copy(out / 'normal_replay.py', 'normal_final_replay.py')
for name in ['kol_cursor_gui.sh', 'kol_cursor_gui_prepare.py', 'kol_cursor_gui_finish.py']:
    q = copy(out / name, 'lifecycle-' + name)
    text = q.read_text()
    for tool in ['kol_cursor_gui.sh', 'kol_cursor_gui_prepare.py', 'kol_cursor_gui_finish.py']:
        text = text.replace('/scratch/' + tool, '/scratch/lifecycle-' + tool)
    q.write_text(text)
    attachments[-1]['sha256'] = sha(q)
sources = []
for folder in ['cmd', 'internal']:
    for p in sorted((src / folder).rglob('*.go')):
        sources.append(dict(path=str(p.relative_to(src)), sha256=sha(p)))
for name in ['go.mod', 'go.sum', 'assets/original-derived-data/character-roster.json',
             'assets/original-derived-png/ega/KOLPICS-MARKERS/kolpics-markers.png',
             'assets/original-derived-png/ega/BATTLE/battle-tiles-08.png',
             'assets/ui-font/manifest.json', 'assets/ui-font/zh-hant-16.png',
             'docs/i18n/display-names-zh-Hant.tsv', 'docs/i18n/ui-zh-Hant.json']:
    sources.append(dict(path=name, sha256=sha(src / name)))
receipt = dict(schema='kol-target-marker-remake-v1', date='2026-10-01',
    status='CONFORMED，限SPEC-010圓環來源／合成／列出的等待層、首次遠程初值與主列重載目前人物指令；父規格未結案',
    image='eob-remake-go:1.26.7-ebiten2.9.9-audio',
    image_id='sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa',
    binary=dict(path='workplace/battle-cursor-20261001/kol-ring-linux-final', sha256=sha(out/'kol-ring-linux-final')),
    original_base=dict(path='receipt.json', sha256=sha(dest/'receipt.json')),
    original_lifecycle=dict(path='lifecycle-receipt.json', sha256=sha(dest/'lifecycle-receipt.json')),
    original_replay=[dict(run=r, keys=json.loads((out/r/'inputs.json').read_text())['planning_keys'],
                         roster_down=json.loads((out/r/'inputs.json').read_text())['initial_roster_down'])
                     for r in ['gui6','gui7','gui9','gui10']],
    summary=dict(original_lifecycle_boxes=52, remake_scaled_boxes=104, different_pixels=0,
                 normal_runs=4, normal_inputs_each=455, normal_states_each=456,
                 normal_full_frames=34, different_pixels_outside_field=0,
                 annotation_entries=14, malformed_atlas_rejected=True,
                 presentation_rng_calls=0, phase_injected_for_pixel_fixture=True),
    build='既有鎖定映像、UID/GID1000、離線module cache、Xvfb：go test ./internal/battle ./internal/ui ./cmd/kol；go build -o /out/kol-ring-linux-final ./cmd/kol',
    normal_changes=changes, inputs=sources, attachments=attachments,
    resolution_backlinks=[dict(module='DOS/KOLTOWN-unpacked.EXE', address_space='IDA linear',
        address='0x249cb', confidence='已證實',
        input_sha256='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2',
        index='docs/re/target-cursor-address-index.json', evidence='docs/re/20261001-target-cursor.md',
        older_spec='docs/re/RESEARCH-LOG.md', required_correction='2026-10-01 RE-171-source 勘誤：主列重建重載目前人物指令')],
    provenance=dict(ida_private_output='workplace/battle-cursor-20261001/cursor-final14-annotated-private.json',
                    ida_private_sha256=sha(out/'cursor-final14-annotated-private.json')),
    corrections=['首輪正常重播第一位是Olear，不能把按鍵說明當作Poontang近戰分支；第二輪先正常YES，再完整走四種子選單',
                 '首輪遠程面板差異來自已證實的起點0,0及空格清除；不放寬第二輪非戰場區零差異判準',
                 'Ebitengine RunGame結束後不能再建圖片，GPU測試改在隔離子行程；修正後同命令乾淨重跑通過',
                 '前版5bf08653沿用其他人物／AI的全域指令；RE-171-source補證sub_249AE前置重載後，先訂正READY規格再修BeginTurn，第三輪以新程式重驗'],
    limits=['正常軌跡不含游標座標，逐框同相位證據由GPU fixture補充；不稱整場同狀態戰鬥',
            '等待速度為BIOS等待契約的現代近似，不稱逐週期／精確wall-clock',
            '回合確認／執行／戰利品其他層、完整攻擊圖示列與保存鏈仍另驗；#33／#64不結案',
            '此為本機中間建置，未重建dist-all、未發布、未變更正式版號'],
    rights='經驗證PNG僅依private GitHub既有例外；不授權公開；原始輸入與反組譯payload僅本機',
    verifier=dict(path='finish_remake.py', sha256=sha(Path(__file__))))
old = dest/'remake-receipt.json'
archive = dest/'remake-receipt-render-only.json'
if old.exists() and not archive.exists():
    shutil.copyfile(old, archive)
receipt['superseded'] = dict(path=archive.name, sha256=sha(archive),
    reason='前版5bf08653未追sub_249AE前置重載，初次／換人摘要來源仍錯；保留其原始PNG、來源與測試，不視為來源已驗收')
(dest/'remake-receipt.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+'\n')
print(json.dumps(receipt['summary'], ensure_ascii=False))
