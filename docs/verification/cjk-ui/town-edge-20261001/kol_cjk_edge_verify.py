import hashlib
import json
import pathlib
import shutil
import subprocess
import sys
sys.path.insert(0, '/src/tools')
from png_io import read_png

root = pathlib.Path('workplace/cjk-edge-20261001')

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def mismatch(path, source, scale, x0, y0):
    screen = read_png(path)
    pic = read_png(source)
    a, b = screen.rgb(), pic.rgb()
    errors = 0
    for y in range(pic.height * scale):
        for x in range(pic.width * scale):
            pos = ((y+y0)*screen.width+x+x0)*3
            src = ((y//scale)*pic.width+x//scale)*3
            errors += a[pos:pos+3] != b[src:src+3]
    return errors

cases = {}
for case, directory in [('edge', root/'after'), ('building', root/'after/building'), ('leave', root/'after/leave')]:
    pairs = [json.loads((directory/lang/'trace.json').read_text()) for lang in ('en','zh-Hant')]
    assert pairs[0] == pairs[1], f'{case} 英中軌跡不同'
    cases[case] = {'step_count': len(pairs[0]['steps']), 'identical_states': True,
                   'command': {lang: json.loads((directory/lang/'run.json').read_text()) for lang in ('en','zh-Hant')}}
    if case == 'edge':
        for lang in ('en','zh-Hant'):
            before = json.loads((root/'before'/lang/'trace.json').read_text())
            assert before == pairs[0], '修補改變了規則狀態'
        assert pairs[0]['steps'][30]['town']['at_edge']
        assert pairs[0]['steps'][31]['town']['at_edge']
        assert not pairs[0]['steps'][32]['town'].get('at_edge', False)
    if case == 'leave':
        last = pairs[0]['steps'][-1]
        # flow 的內層名稱仍是 town；正式切換以 left_town 與 world 狀態表示。
        assert last['town']['left_town'] and last['world']['mode'] == 0, '沒有正常出城'

source = pathlib.Path('assets/original-derived-png/ega/MISCPICS/miscpics-000.png')
pixels = {}
for step in (30,31):
    errors = mismatch(root/'after/zh-Hant'/f'step-{step:03}.png',source,3,312,174)
    assert errors == 0, f'城門第 {step} 步林地圖像素不符：{errors}'
    pixels[f'gate-{step}'] = errors
before_errors = mismatch(root/'before/zh-Hant/step-030.png',source,3,312,174)
assert before_errors > 0, '修補前反面對照意外通過'
pixels['before_gate_mismatches'] = before_errors
door_source = pathlib.Path('assets/original-derived-png/ega/TWN7PICS/twn7pics-013.png')
pixels['building'] = mismatch(root/'after/building/zh-Hant/step-021.png',door_source,3,312,174)
assert pixels['building'] == 0
pixels['english_unchanged'] = all(sha(p) == sha(root/'after/en'/p.name) for p in (root/'before/en').glob('step-*.png'))
assert pixels['english_unchanged']
pixels['return_map_unchanged'] = sha(root/'before/zh-Hant/step-032.png') == sha(root/'after/zh-Hant/step-032.png')
assert pixels['return_map_unchanged']

free = json.loads((root/'free-tile-audit.json').read_text())
assert sum(len(t['matching_tiles']) for t in free['towns']) == 43
assert all(v == 0 for t in free['towns'] for v in t['map_references'].values())

target = pathlib.Path('docs/verification/cjk-ui/town-edge-20261001')
if target.exists():
    raise RuntimeError('正式收據目錄已存在，拒絕覆寫')
assert target.parent.stat().st_uid == 1000
target.mkdir()
files = {}
def copy(src, name):
    shutil.copyfile(src,target/name)
    files[name] = sha(target/name)

for case, directory in [('edge', root/'after'), ('building', root/'after/building'), ('leave', root/'after/leave')]:
    for lang in ('en','zh-Hant'):
        copy(directory/lang/'trace.json',f'{case}-{lang}-trace.json')
copy(root/'before/zh-Hant/trace.json','before-zh-Hant-trace.json')
for src, name in [(root/'after/zh-Hant/step-030.png','gate-zh-Hant.png'),
                  (root/'after/zh-Hant/step-031.png','gate-uturn-zh-Hant.png'),
                  (root/'after/zh-Hant/step-032.png','return-map-zh-Hant.png'),
                  (root/'after/building/zh-Hant/step-021.png','building-zh-Hant.png'),
                  (root/'after/leave/zh-Hant/step-031.png','leave-world-zh-Hant.png'),
                  (root/'before/zh-Hant/step-030.png','before-gate-zh-Hant.png'),
                  (root/'go-test.log','go-test.log'),
                  (root/'english-regression.log','english-regression.log'),
                  (root/'free-tile-audit.json','free-tile-audit.json')]:
    copy(src,name)
for r in free['references']:
    copy(root/r['file'],r['file'])
for name in ['kol_cjk_edge_run.py','kol_cjk_edge_control.sh','kol_cjk_edge_build.sh','kol_free_tile_audit.py','kol_cjk_edge_verify.py']:
    copy(pathlib.Path('/control')/name, name)
record = {'schema':'kol-cjk-town-edge-verification-v1','date':'2026-10-01',
          'scope':'城門、建築門面、返回地圖與出城的繁中顯示；三條從標題開始的正常玩家按鍵路徑',
          'status':'CONFORMED（SPEC-016 城門繁中繪圖子契約）',
          'limitations':'原版沿用 RE-065 的 IDA 證據與 DOSBox-X 輔助基準；沒有新 dosgolem 對拍，不宣稱全遊戲繁中或完整原版 parity。FREE 結論限所列 43 個已目視字形及固定地圖，不含其他相似字形。',
          'git_head':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
          'binary':{'path':str(root/'kol-edge-fixed'),'sha256':sha(root/'kol-edge-fixed')},
          'toolchain':{'image':'eob-remake-go:1.26.7-ebiten2.9.9-audio',
                       'image_id':'sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa',
                       'go':'1.26.7','ebitengine':'v2.9.9'},
          'source_zip':{'file':'Knights-of-legend.zip','sha256':sha(pathlib.Path('Knights-of-legend.zip'))},
          'source_pixels':{str(source):sha(source),str(door_source):sha(door_source)},
          'production_sources':{str(p):sha(p) for base in ('cmd','internal') for p in sorted(pathlib.Path(base).rglob('*.go')) if not p.name.endswith('_test.go')},
          'cases':cases,'pixel_checks':pixels,
          'visual_review':{'status':'passed','method':'已目視修補前後城門、門面、返回地圖、出城及三種 FREE 來源圖塊'},
          'rights':'PNG 僅依專案已核准的 private GitHub 例外；不得公開發行原作衍生素材',
          'files':files}
(target/'receipt.json').write_text(json.dumps(record,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'cases':{k:v['step_count'] for k,v in cases.items()},'pixels':pixels,
                  'receipt':str(target/'receipt.json'),'files':len(files),'binary':record['binary']},ensure_ascii=False))
