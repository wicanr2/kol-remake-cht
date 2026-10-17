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
import math
from PIL import Image
from png_io import write_png_rgb

ROOT, OUT = Path('/src'), Path('/out')
P = OUT / 'promo'
W, H, FPS = 1920, 1080, 25
VERSION = json.loads((OUT / 'build-inputs.json').read_text())['version']
FULL_HD = json.loads((OUT / 'build-inputs.json').read_text()).get('program_matches_full_hd_receipt', False)
assert re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}', VERSION)
FONT = json.loads((ROOT / 'assets/ui-font/manifest.json').read_text())['fonts'][0]
GLYPHS = {g['character']: g for g in FONT['glyphs']}
HD_FONT_DIR = ROOT / 'assets/hd/first-slice/interface'
HD_FONT = json.loads((HD_FONT_DIR / 'font-atlas.json').read_text()) if FULL_HD else None
HD_ATLAS = Image.open(HD_FONT_DIR / 'font-atlas.png').convert('RGBA') if FULL_HD else None


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def run(cmd, log=None):
    result = subprocess.run(cmd, check=True, env=dict(os.environ, MAGICK_THREAD_LIMIT='2', OMP_NUM_THREADS='2'), stdout=subprocess.PIPE,
                            stderr=subprocess.PIPE, timeout=300)
    if log:
        log.write_bytes(result.stdout + result.stderr)
    return result.stdout


def label(text, scale, color, name):
    if FULL_HD:
        ratio = {7: 116, 3: 42, 2: 30}[scale] / HD_FONT['size']
        missing = set(text) - set(HD_FONT['glyphs'])
        assert not missing, ('HD caption glyphs missing', ''.join(sorted(missing)))
        width = math.ceil(sum(HD_FONT['glyphs'][c]['advance'] for c in text) * ratio) + 24
        height = math.ceil(HD_FONT['cell'] * ratio) + 24
        canvas = Image.new('RGBA', (width, height))
        pen = 12.0
        for c in text:
            g = HD_FONT['glyphs'][c]
            cell = HD_FONT['cell']
            alpha = HD_ATLAS.crop((g['x'], g['y'], g['x']+cell, g['y']+cell)).getchannel('A')
            extent = round(cell * ratio)
            alpha = alpha.resize((extent, extent), Image.Resampling.LANCZOS)
            glyph = Image.new('RGBA', (extent, extent), (*color, 255))
            glyph.putalpha(alpha)
            canvas.alpha_composite(glyph, (round(pen-HD_FONT['xpad']*ratio), 12))
            pen += g['advance'] * ratio
        bounds = canvas.getbbox()
        assert bounds, text
        canvas = canvas.crop(bounds)
        assert canvas.width <= W-160 and canvas.height <= 180, (text, canvas.size)
        result = Image.new('RGB', canvas.size, (32, 20, 20))
        result.paste(canvas, mask=canvas.getchannel('A'))
        path = P / (name + '.png')
        result.save(path)
        return path
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
        dict(kind='runtime', file='smoke/linux-full-controls/inn-hd.png', caption='HD介面與人物，從旅店出發', seconds=7),
        dict(kind='runtime', file='smoke/linux-full-controls/menu-ega.png', caption='配樂獨立開關，畫面即時切換', seconds=5),
        dict(kind='runtime', file='smoke/linux-full-controls/menu-hd.png', caption='四首原創室內樂，陪伴冒險', seconds=7),
        dict(kind='shot', step=125, caption='探索城鎮與野外', seconds=8),
        dict(kind='shot', step=300, caption='逐人規劃，逐手結算', seconds=9),
        dict(kind='shot', step=600, caption='武器、體力與傷勢，決定每一手', seconds=9),
        dict(kind='shot', step=1045, caption='回城住宿，保存冒險進度', seconds=8),
        dict(kind='card', title='武士傳說', sub=VERSION,
             caption='Linux  Windows  macOS', seconds=10),
    ]
    if FULL_HD:
        def hd(name, caption, seconds):
            return dict(kind='runtime', file='smoke/hd-scenes-linux/motion-runtime/'+name+'-first-hd.png',
                        caption=caption, seconds=seconds, controlled=True)
        scenes = [
            dict(kind='card', title='武士傳說', sub='Knights of Legend', caption='繁體中文與手繪HD', seconds=6),
            hd('title-02-SET2PICS-05', '手繪片頭與繁體中文', 7),
            dict(kind='runtime', file='smoke/linux-full-controls/inn-hd.png', caption='從旅店開始六人冒險', seconds=7),
            hd('map-world', '世界探索與原有地圖', 6),
            hd('map-town3', '七座城鎮的手繪風景', 6),
            hd('weather-4-01', '原有天候，手繪呈現', 6),
            hd('battle-08-map-16', '逐人規劃，逐手結算', 7),
            hd('body-1-1', '體力與傷勢，保留原有規則', 5),
            hd('monster-native-encounter', '手繪怪物，保留原有位置', 6),
            hd('event-combat-03', '戰鬥事件與中文訊息', 6),
            dict(kind='split', files=['smoke/linux-full-controls/menu-ega.png','smoke/linux-full-controls/menu-hd.png'],
                 caption='遊戲內 F10 切換 EGA 與 HD', seconds=6),
            dict(kind='card', title='武士傳說', sub=VERSION, caption='Linux  Windows  macOS', seconds=8),
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
        elif scene['kind'] == 'split':
            source_files = [OUT/n for n in scene['files']]
            for n, shot in enumerate(source_files):
                assert shot.is_file(), shot
                fitted = tmp / f'fit-{i}-{n}.png'
                run(['convert', str(shot), '-filter', 'point', '-resize', '824x720',
                     '-bordercolor', '#c8992e', '-border', '3', str(fitted)])
                command += [str(fitted), '-gravity', 'center', '-geometry', ('-440-25' if n == 0 else '+440-25'), '-composite']
            a = label(scene['caption'], 3, (240,230,207), name+'-caption')
            command += [str(a), '-gravity', 'south', '-geometry', '+0+100', '-composite']
            scene['source_sha256'] = {str(p.relative_to(OUT)):sha(p) for p in source_files}
            scene['trace_state'] = 'actual full AppImage F10 settings; same normal inn state'
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
            scene['trace_state'] = (trace[scene['step']]['scene'] if 'step' in scene else
                                    'SPEC-040 controlled scene, same accepted production renderer' if scene.get('controlled') else
                                    '正式封包，遊戲內畫面設定')
        if FULL_HD and scene.get('controlled'):
            note = label('受控場景展示', 2, (240,230,207), name+'-scope')
            command += [str(note), '-gravity', 'south', '-geometry', '+0+37', '-composite']
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
    music = {name:ROOT/'assets/audio/hd'/(name+'-chamber.ogg') for name in ('title','inn','exploration','combat')}
    music_sections=[dict(track='title',start=0,seconds=13),dict(track='inn',start=13,seconds=19),
                    dict(track='exploration',start=32,seconds=8),dict(track='combat',start=40,seconds=18),
                    dict(track='inn',start=58,seconds=8),dict(track='title',start=66,seconds=10)]
    if FULL_HD:
        music_sections = [dict(track='title',start=0,seconds=13),dict(track='inn',start=13,seconds=7),
                          dict(track='exploration',start=20,seconds=18),dict(track='combat',start=38,seconds=24),
                          dict(track='inn',start=62,seconds=6),dict(track='title',start=68,seconds=8)]
    video = P / f'kol-{VERSION}-promo.mp4'
    if not reuse:
        if FULL_HD:
            tracks = {'title':'2:a','inn':'3:a','exploration':'4:a','combat':'5:a'}
            filters = ['[1:a]asplit=2[x][y]', '[x]adelay=900|900,volume=0.7[a]',
                       '[y]adelay=6500|6500,volume=0.5[b]', '[2:a]asplit=2[t0][t1]', '[3:a]asplit=2[i0][i1]']
            labels = ['t0','i0','4:a','5:a','i1','t1']
            mixed = ['a','b']
            for j,(section,source) in enumerate(zip(music_sections,labels)):
                duration,offset=section['seconds'],section['start']
                filters.append(f'[{source}]atrim=duration={duration},asetpts=PTS-STARTPTS,volume=0.8,'
                               f'afade=t=in:d=0.3,afade=t=out:st={duration-.5}:d=0.5,'
                               f'adelay={offset*1000}|{offset*1000}[m{j}]')
                mixed.append('m'+str(j))
            filters.append(''.join('['+x+']' for x in mixed)+'[6:a]amix=inputs=9:duration=longest:normalize=0[s]')
            run(['ffmpeg','-y','-v','error','-i',str(silent),'-i',str(cue),
                 *sum((['-i',str(path)] for path in music.values()),[]),
                 '-f','lavfi','-t',str(total),'-i','anullsrc=r=44100:cl=stereo',
                 '-filter_complex',';'.join(filters),'-map','0:v','-map','[s]',
                 '-c:v','copy','-c:a','aac','-b:a','160k','-t',str(total),'-movflags','+faststart',str(video)])
        else:
            run(['ffmpeg', '-y', '-v', 'error', '-i', str(silent), '-i', str(cue),
         *sum((['-i',str(path)] for path in music.values()),[]),
         '-f', 'lavfi', '-t', str(total), '-i', 'anullsrc=r=44100:cl=stereo',
         '-filter_complex', '[1:a]asplit=2[x][y];[x]adelay=900|900,volume=0.7[a];'
         '[y]adelay=33800|33800,volume=0.5[b];'
         '[2:a]asplit=2[t0][t1];[3:a]asplit=2[i0][i1];'
         '[t0]atrim=duration=13,asetpts=PTS-STARTPTS,volume=0.8,afade=t=in:d=0.3,afade=t=out:st=12.7:d=0.3[ta];'
         '[i0]atrim=duration=19,asetpts=PTS-STARTPTS,volume=0.8,afade=t=in:d=0.3,afade=t=out:st=18.7:d=0.3,adelay=13000|13000[ia];'
         '[4:a]atrim=duration=8,asetpts=PTS-STARTPTS,volume=0.8,afade=t=in:d=0.3,afade=t=out:st=7.7:d=0.3,adelay=32000|32000[ea];'
         '[5:a]atrim=duration=18,asetpts=PTS-STARTPTS,volume=0.8,afade=t=in:d=0.3,afade=t=out:st=17.7:d=0.3,adelay=40000|40000[ca];'
         '[i1]atrim=duration=8,asetpts=PTS-STARTPTS,volume=0.8,afade=t=in:d=0.3,afade=t=out:st=7.7:d=0.3,adelay=58000|58000[ib];'
         '[t1]atrim=duration=10,asetpts=PTS-STARTPTS,volume=0.8,afade=t=in:d=0.3,afade=t=out:st=9.2:d=0.8,adelay=66000|66000[tb];'
         '[a][b][ta][ia][ea][ca][ib][tb][6:a]amix=inputs=9:duration=longest:normalize=0[s]',
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
    assert 'silence_duration:' not in (P/'silence.log').read_text(),'unexpected interval >=1 second below -45dB'
    frames = []
    for i, scene in enumerate(sources):
        f = P / f'frame-{i:02d}.png'
        run(['ffmpeg', '-y', '-v', 'error', '-ss', str(scene['center']), '-i', str(video),
             '-frames:v', '1', '-threads', '2', str(f)])
        frames.append(str(f))
    assert len({sha(Path(f)) for f in frames}) == len(frames), 'duplicate decoded scene centers'
    if FULL_HD:
        freeze_durations = [float(x) for x in re.findall(r'freeze_duration: ([\d.]+)', (P/'freeze.log').read_text())]
        assert all(x <= max(s['seconds'] for s in scenes)+.2 for x in freeze_durations), freeze_durations
    rows = math.ceil(len(frames)/2)
    run(['ffmpeg', '-y', '-v', 'error', '-framerate', '1', '-i', str(P / 'frame-%02d.png'),
         '-vf', f'scale=640:360,tile=2x{rows}:nb_frames={len(frames)}:padding=8:margin=8',
         '-frames:v', '1', '-threads', '2', str(P / 'contact-sheet.png')])
    metadata = dict(schema='kol-promo/2', version=VERSION, seconds=total, fps=FPS,
       dimensions=[W,H], video_sha256=sha(video), sources=sources,
       picture_policy='本版AppImage唯一正常重播及遊戲內切換驗證的新截圖與自製文字卡；沒有舊幀、原版對照或假遊玩。',
       audio_source='assets/audio/title-thunder.ogg', audio_sha256=sha(cue),
       audio_policy='已核准的四首重製版原創室內樂與雷聲，依分鏡節錄並淡入淡出；沒有原版音樂取樣。',
       music_sources={name:dict(path=str(path.relative_to(ROOT)),sha256=sha(path)) for name,path in music.items()},
       music_sections=music_sections,
       font_source='assets/ui-font/manifest.json', font_sha256=sha(ROOT/'assets/ui-font/manifest.json'),
       theme='赤石與遊戲點陣字模',
       freeze_policy='截圖剪輯，各幕5至10秒停留是分鏡；freezedetect逐段保留，不宣稱連續實機錄影。',
       volume_peak_db=peak, black_detect='PASS: no interval >=0.3 seconds at specified thresholds',
       rights='原作衍生畫面，僅本機保存，不得當成已取得公開素材權利。',
       visual_review='pending')
    if FULL_HD:
        metadata.update(picture_policy='本版實際AppImage正常旅店／F10設定及同一正式來源的SPEC-040受控展示器；每幕明示範圍，沒有原版oracle或自然通關聲明。',
                        font_source='assets/hd/first-slice/interface/font-atlas.json', font_sha256=sha(HD_FONT_DIR/'font-atlas.json'),
                        font_atlas_sha256=sha(HD_FONT_DIR/'font-atlas.png'),theme='遊戲赤黑銅框與原生HD中文字',
                        freeze_policy='12幕靜態展示，各幕5至8秒；解碼中心皆不同，freeze_duration不得跨越最長分鏡，非連續實機錄影。',
                        source_acceptance_receipt_sha256=sha(ROOT/'docs/verification/hd-full-runtime-20261017/receipt.json'),
                        differences_from_previous=['完整手繪片頭、地圖、天候與戰場','原生HD襯線字','同一場景EGA／HD並排','受控展示逐幕標示'])
    (P / 'sources.json').write_text(json.dumps(metadata, ensure_ascii=False, indent=2)+'\n')
    shutil.copy2(Path(__file__), P / 'compose_release_promo.py')
    print('PROMO PASS technical:', total, 'seconds', W, H, 'peak', peak, flush=True)


if __name__ == '__main__':
    main()
