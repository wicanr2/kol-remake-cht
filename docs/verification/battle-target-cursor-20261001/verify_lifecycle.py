"""在 Docker 內核對原版生命週期；/out 為本機探針根，/src 為專案。"""
import hashlib
import json
from pathlib import Path
import shutil
import sys

sys.path.insert(0, '/src/tools')
from png_io import read_png

root = Path('/out')
dest = Path('/src/docs/verification/battle-target-cursor-20261001')
assert dest.stat().st_uid == root.stat().st_uid
assets = Path('/src/assets/original-derived-png/ega')
marker = read_png(assets / 'KOLPICS-MARKERS/kolpics-markers.png')
atlas = read_png(assets / 'BATTLE/battle-tiles-08.png')
assert (marker.width, marker.height) == (16, 160)
assert (atlas.width, atlas.height) == (160, 128)

def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()

def tile(i):
    assert 0 <= i < 80
    return atlas.crop(i % 10 * 16, i // 10 * 16, 16, 16).rgb()

def composite(bg, fg):
    result = bytearray(bg)
    bit = lambda x: 1 << (7 - x if x < 8 else 23 - x)
    for y in range(16):
        mask = sum(bit(x) for x in range(16)
                   if fg[(y * 16 + x) * 3:(y * 16 + x) * 3 + 3] != b'\0\0\0')
        wide = (mask | mask << 1 | mask >> 1) & 65535
        for x in range(16):
            if wide & bit(x):
                i = (y * 16 + x) * 3
                result[i:i + 3] = fg[i:i + 3]
    return bytes(result)

# 按實際圖示列與寫回紀錄指定場景，不能用模式5單獨判定。
cases = {
    'gui6': {1: True, 2: True, 3: True, 5: False, 8: True, 10: False, 14: False},
    'gui7': {7: True, 9: True, 10: False, 15: False, 24: False,
             31: True, 34: True, 38: True},
    'gui9': {**{i: True for i in range(2, 12)}, 12: False},
    'gui10': {i: True for i in range(2, 28)},
}
checks, inputs = [], []
for run, frames in cases.items():
    folder = root / run
    calls = json.loads((folder / 'cursor-calls.json').read_text())['calls']
    renders = [c for c in calls if c['ida_ea'] == '0x1CB28' and c['caller_ida'] == '0x114AB']
    steps = json.loads((folder / 'battle-planning.json').read_text())['steps']
    if run == 'gui10':
        # Olear 的 +0x1D 為1；射擊摘要仍動畫，再開啟保留右下角記憶目標。
        s = json.loads((folder / 'plan-25.png.state.json').read_text())
        assert bytes.fromhex(s['actor_records'][0])[0x1D] == 1
        assert s['summary_flag'] == 1
    for n, shown in frames.items():
        frame = f'plan-{n:02d}.png'
        statefile = folder / (frame + '.state.json')
        s = json.loads(statefile.read_text())
        row, col = s['cursor_row'], s['cursor_col']
        assert 0 <= row <= 8 and 0 <= col <= 8
        pos = 9 * row + col
        tid = bytes.fromhex(s['visible_tiles'])[pos]
        sid = bytes.fromhex(s['visible_sprites'])[pos]
        bg = tile(tid)
        if sid:
            bg = composite(bg, tile(sid - 1))
        expected = bg
        c = next(c for c in reversed(renders) if c['step'] <= s['step'])
        if shown:
            assert c['args'][1] == 0x153 + 640 * row + 2 * col
            assert 5 <= c['args'][0] <= 9
            assert c['args'][3] == (100 if sid else tid)
            expected = composite(bg, marker.crop(0, c['args'][0] * 16, 16, 16).rgb())
        actual = read_png(folder / frame).crop(152 + col * 16, 8 + row * 16, 16, 16).rgb()
        diff = sum(actual[i:i + 3] != expected[i:i + 3] for i in range(0, 768, 3))
        assert diff == 0, (run, frame, shown, diff)
        bad = bytearray(expected)
        bad[0] ^= 1
        assert bad != actual
        target = dest / f'{run}-{frame}'
        shutil.copyfile(folder / frame, target)
        checks.append(dict(run=run, frame=target.name, frame_sha256=sha(target),
                           key=steps[n + 1]['key'], icon_slots=steps[n + 1]['icon_slots'],
                           row=row, col=col, tile=tid, sprite_code=sid, shown=shown,
                           unit=c['args'][0] if shown else None,
                           capture_step=s['step'], last_draw_step=c['step'],
                           rect=[152 + col * 16, 8 + row * 16, 16, 16],
                           different_pixels=diff, rgb_sha256=hashlib.sha256(actual).hexdigest(),
                           one_bit_negative_rejected=True))
        inputs.append(dict(path=str(statefile.relative_to(Path('/src'))) if str(statefile).startswith('/src/')
                           else 'workplace/battle-cursor-20261001/' + run + '/' + statefile.name,
                           sha256=sha(statefile)))
    for name in ['inputs.json', 'cursor-calls.json', 'battle-planning.json', 'run.log',
                 'instrumented-book-startup-private.go']:
        p = folder / name
        inputs.append(dict(path='workplace/battle-cursor-20261001/' + run + '/' + name, sha256=sha(p)))
payload = dict(schema='kol-original-target-lifecycle-v1', date='2026-10-01',
               evidence='已證實；限逐框列出的正常原版路徑與像素',
               dosgolem_revision='968ca774705853276ae55c8d1163c369481ac71f', seed=1,
               rng_reroll=False, state_injection=False, direct_entry=False,
               remake_compared=False, checks=checks, private_sources=inputs,
               rights='已驗證PNG僅依private GitHub既有例外；不授權公開；原始快照只留本機',
               verifier=dict(path=Path(__file__).name, sha256=sha(Path(__file__))))
base = json.loads((dest / 'receipt.json').read_text())
payload['original_base'] = dict(path='receipt.json', sha256=sha(dest / 'receipt.json'))
payload['source_zip_sha256'] = base['source_zip_sha256']
payload['original_tools'] = base['original_tools']
payload['executable_sources'] = base['replay'][0]['executable_sources']
payload['assets'] = json.loads((dest / 'original-pixel-checks.json').read_text())['assets']
(dest / 'lifecycle-receipt.json').write_text(json.dumps(payload, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(dict(checks=len(checks), different_pixels=0), ensure_ascii=False))
