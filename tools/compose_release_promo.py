#!/usr/bin/env python3
"""由 v1 唯一正常抽樣的新畫格製作本機推廣片；入口見本版交付收據。"""
from pathlib import Path
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
from png_io import write_png_rgb

ROOT, OUT = Path('/src'), Path('/out')
P = OUT / 'promo'
W, H, FPS = 1920, 1080, 25
VERSION = json.loads((OUT / 'build-inputs.json').read_text())['version']
assert re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}', VERSION)
FONT = json.loads((ROOT / 'assets/ui-font/manifest.json').read_text())['fonts'][0]
GLYPHS = {g['character']: g for g in FONT['glyphs']}


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def run(cmd, log=None):
    result = subprocess.run(cmd, check=True, env=dict(os.environ, MAGICK_THREAD_LIMIT='2', OMP_NUM_THREADS='2'), stdout=subprocess.PIPE,
                            stderr=subprocess.PIPE, timeout=300)
    if log:
        log.write_bytes(result.stdout + result.stderr)
    return result.stdout


def label(text, scale, color, name):
    advances = [8 if c == ' ' else GLYPHS[c]['advance'] for c in text]
    width, height = sum(advances) * scale, 18 * scale
    assert width <= W - 160 and height <= 180, (text, width, height)
    pixels = bytearray(bytes((32, 20, 20)) * width * height)
    x = 0
    for c, advance in zip(text, advances):
        if c != ' ':
            raw = bytes.fromhex(GLYPHS[c]['rows_hex'])
            for y in range(18):
                row = int.from_bytes(raw[y*2:y*2+2], 'big')
                for gx in range(min(advance, 16)):
                    if row & (1 << (15 - gx)):
                        for dy in range(scale):
                            for dx in range(scale):
                                i = ((y*scale+dy)*width+x+gx*scale+dx)*3
                                pixels[i:i+3] = bytes(color)
        x += advance * scale
    path = P / (name + '.png')
    write_png_rgb(path, width, height, bytes(pixels))
    return path


def main():
    assert Path('/.dockerenv').exists()
    sample = json.loads((OUT / 'smoke/sampling.json').read_text())
    assert sample['status'] == 'PASS' and sample['scenario_count'] == 1
    trace = json.loads((OUT / 'smoke/normal/trace.json').read_text())['steps']
    tmp = Path('/tmp/kol-promo')
    tmp.mkdir()
    # 全部取自這次封包的同一次正常重播；圖像不裁切、不改色，不宣稱原版同狀態。
    scenes = [
        dict(kind='card', title='武士傳說', sub='Knights of Legend',
             caption='繁體中文重製', seconds=6),
        dict(kind='shot', step=3, caption='選擇隊員，開始冒險', seconds=7),
        dict(kind='shot', step=30, caption='組成隊伍，從旅店出發', seconds=7),
        dict(kind='runtime', file='smoke/linux-gui/game-menu-ega.png', caption='遊戲內開啟畫面設定', seconds=5),
        dict(kind='runtime', file='smoke/linux-gui/game-menu-hd.png', caption='即時切換 HD，冒險接著走', seconds=7),
        dict(kind='shot', step=125, caption='探索城鎮與野外', seconds=8),
        dict(kind='shot', step=300, caption='逐人規劃，逐手結算', seconds=9),
        dict(kind='shot', step=600, caption='武器、體力與傷勢，決定每一手', seconds=9),
        dict(kind='shot', step=1045, caption='回城住宿，保存冒險進度', seconds=8),
        dict(kind='card', title='武士傳說', sub=VERSION,
             caption='Linux  Windows  macOS', seconds=10),
    ]
    total = sum(s['seconds'] for s in scenes)
    assert total == 76
    reuse = '--review-existing' in sys.argv
    sources, concat, t = [], [], 0
    for i, scene in enumerate(scenes):
        name = f'scene-{i:02d}'
        canvas = P / (name + '.png')
        command = ['convert', '-size', f'{W}x{H}', 'xc:#201414', '-fill', '#c8992e',
                   '-draw', 'rectangle 64,64 1856,67', '-draw', 'rectangle 64,1013 1856,1016']
        if scene['kind'] == 'card':
            a = label(scene['title'], 7, (224, 180, 80), name + '-title')
            b = label(scene['sub'], 3, (240, 230, 207), name + '-sub')
            c = label(scene['caption'], 3, (240, 230, 207), name + '-caption')
            command += [str(a), '-gravity', 'center', '-geometry', '+0-110', '-composite',
                        str(b), '-geometry', '+0+70', '-composite',
                        str(c), '-geometry', '+0+200', '-composite']
            if i == len(scenes)-1:
                d = label('需自備合法遊戲資料', 2, (240, 230, 207), name + '-data')
                command += [str(d), '-geometry', '+0+310', '-composite']
        else:
            shot = OUT / scene['file'] if scene['kind'] == 'runtime' else OUT / 'smoke/normal' / f'step-{scene["step"]:03d}.png'
            assert shot.exists(), shot
            fitted = tmp / f'fit-{i}.png'
            run(['convert', str(shot), '-filter', 'point', '-resize', '1536x800',
                 '-bordercolor', '#c8992e', '-border', '3', str(fitted)])
            a = label(scene['caption'], 3, (240, 230, 207), name + '-caption')
            command += [str(fitted), '-gravity', 'center', '-geometry', '+0-25', '-composite',
                        str(a), '-gravity', 'south', '-geometry', '+0+100', '-composite']
            scene['source'] = str(shot.relative_to(OUT))
            scene['source_sha256'] = sha(shot)
            scene['trace_state'] = trace[scene['step']]['scene'] if 'step' in scene else '正式封包，遊戲內畫面設定'
        command += ['-type', 'TrueColor', str(canvas)]
        run(command)
        clip = tmp / f'{i:02d}.mp4'
        if not reuse:
            run(['ffmpeg', '-y', '-v', 'error', '-loop', '1', '-framerate', str(FPS), '-i', str(canvas),
                 '-t', str(scene['seconds']), '-vf', 'format=yuv420p', '-threads', '2',
                 '-c:v', 'libx264', '-preset', 'veryfast', '-crf', '20', str(clip)])
        concat.append("file '" + str(clip) + "'")
        scene.update(start=t, end=t+scene['seconds'], center=t+scene['seconds']/2,
                     composition_sha256=sha(canvas))
        sources.append(scene)
        t += scene['seconds']
    listing = tmp / 'concat.txt'
    listing.write_text('\n'.join(concat)+'\n')
    silent = tmp / 'silent.mp4'
    if not reuse:
        run(['ffmpeg', '-y', '-v', 'error', '-f', 'concat', '-safe', '0', '-i', str(listing),
             '-c', 'copy', str(silent)])
    cue = ROOT / 'assets/audio/title-thunder.ogg'
    video = P / f'kol-{VERSION}-promo.mp4'
    if not reuse:
        run(['ffmpeg', '-y', '-v', 'error', '-i', str(silent), '-i', str(cue),
         '-f', 'lavfi', '-t', str(total), '-i', 'anullsrc=r=44100:cl=stereo',
         '-filter_complex', '[1:a]asplit=2[x][y];[x]adelay=900|900,volume=0.7[a];'
         '[y]adelay=33800|33800,volume=0.5[b];[a][b][2:a]amix=inputs=3:duration=longest:normalize=0[s]',
         '-map', '0:v', '-map', '[s]', '-c:v', 'copy', '-c:a', 'aac', '-b:a', '160k',
             '-t', str(total), '-movflags', '+faststart', str(video)])
    probe = json.loads(run(['ffprobe', '-v', 'error', '-show_streams', '-show_format', '-of', 'json', str(video)]))
    (P / 'ffprobe.json').write_text(json.dumps(probe, indent=2)+'\n')
    streams = {s['codec_type']:s for s in probe['streams']}
    assert streams['video']['codec_name'] == 'h264' and streams['audio']['codec_name'] == 'aac'
    assert streams['video']['width'] == W and streams['video']['height'] == H
    assert abs(float(probe['format']['duration']) - total) < 0.15
    for kind, flag, filt in [('volume', '-af', 'volumedetect'),
                             ('silence', '-af', 'silencedetect=n=-45dB:d=1'),
                             ('black', '-vf', 'blackdetect=d=0.3:pix_th=0.02:pic_th=0.99'),
                             ('freeze', '-vf', 'freezedetect=n=-60dB:d=2')]:
        run(['ffmpeg', '-hide_banner', '-v', 'info', '-i', str(video), flag, filt,
             '-f', 'null', '-'], log=P / (kind + '.log'))
    loud = (P / 'volume.log').read_text()
    peak = float(re.search(r'max_volume: ([-\d.]+) dB', loud)[1])
    assert -30 < peak < -0.1, peak
    assert 'black_start:' not in (P / 'black.log').read_text()
    frames = []
    for i, scene in enumerate(sources):
        f = P / f'frame-{i:02d}.png'
        run(['ffmpeg', '-y', '-v', 'error', '-ss', str(scene['center']), '-i', str(video),
             '-frames:v', '1', '-threads', '2', str(f)])
        frames.append(str(f))
    run(['ffmpeg', '-y', '-v', 'error', '-framerate', '1', '-i', str(P / 'frame-%02d.png'),
         '-vf', 'scale=640:360,tile=2x5:nb_frames=10:padding=8:margin=8',
         '-frames:v', '1', '-threads', '2', str(P / 'contact-sheet.png')])
    metadata = dict(schema='kol-promo/2', version=VERSION, seconds=total, fps=FPS,
       dimensions=[W,H], video_sha256=sha(video), sources=sources,
       picture_policy='本版AppImage唯一正常重播及遊戲內切換驗證的新截圖與自製文字卡；沒有舊幀、原版對照或假遊玩。',
       audio_source='assets/audio/title-thunder.ogg', audio_sha256=sha(cue),
       audio_policy='只有使用者已接受的遊戲雷聲。沒有配樂；其餘靜音是刻意留白。',
       font_source='assets/ui-font/manifest.json', font_sha256=sha(ROOT/'assets/ui-font/manifest.json'),
       theme='赤石與遊戲點陣字模',
       freeze_policy='截圖剪輯，各幕5至10秒停留是分鏡；freezedetect逐段保留，不宣稱連續實機錄影。',
       volume_peak_db=peak, black_detect='PASS: no interval >=0.3 seconds at specified thresholds',
       rights='原作衍生畫面，僅本機保存，不得當成已取得公開素材權利。',
       visual_review='pending')
    (P / 'sources.json').write_text(json.dumps(metadata, ensure_ascii=False, indent=2)+'\n')
    shutil.copy2(Path(__file__), P / 'compose_release_promo.py')
    print('PROMO PASS technical:', total, 'seconds', W, H, 'peak', peak, flush=True)


if __name__ == '__main__':
    main()
