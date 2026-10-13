#!/usr/bin/env python3
"""Docker內封存四首原創配樂的正式接入收據；/out為本輪工作區。"""
import hashlib,json,shutil,wave
from pathlib import Path
import numpy as np
from PIL import Image,ImageChops,ImageFilter
ROOT=Path('/src');OUT=Path('/out');DEST=ROOT/'docs/verification/hd-music-runtime-20261013'
assert Path('/.dockerenv').exists()
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def write(p,d):p.write_text(json.dumps(d,ensure_ascii=False,indent=2)+'\n')
runtime=json.loads((OUT/'runtime-audio.json').read_text());paths=json.loads((OUT/'player-paths.json').read_text())
assert runtime['status']==paths['status']=='PASS'
assert {r['name'] for r in runtime['runs']}=={'title','inn','missing'}
for row in runtime['runs']+[paths['adventure']]:
    p=OUT/(row['name']+'.wav');assert sha(p)==row['sha256']
    with wave.open(str(p),'rb') as f:
        assert f.getframerate()==44100 and f.getnchannels()==2 and f.getsampwidth()==2
        s=np.frombuffer(f.readframes(f.getnframes()),dtype='<i2').astype(np.int32)
    assert np.max(np.abs(s[-88200:]))==0
    assert int(np.max(np.abs(s)))==row['peak']<32767
    assert row['peak']==0 if row['name']=='missing' else row['peak']>100
program={str(p.relative_to(ROOT)):sha(p) for name in ('cmd','internal') for p in sorted((ROOT/name).rglob('*.go'))}
program.update({n:sha(ROOT/n) for n in ('go.mod','go.sum')})
write(OUT/'source-inputs.json',program)
DEST.mkdir(exist_ok=True)
screens=('inn-controls-hd.png','inn-controls-small.png','world-music-hd.png','combat-music-hd.png','missing-music.png','music-contact-sheet.png')
entries={}
for name in screens:
    source=OUT/name;assert source.stat().st_uid==source.stat().st_gid==1000;shutil.copyfile(source,DEST/name)
    image=Image.open(source);assert image.size in [(1920,1200),(960,600)]
    entries[name]=dict(sha256=sha(source),width=image.width,height=image.height,rights='private GitHub PNG exception; local remake capture; no public distribution grant',source_zip_sha256='2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46',frame=name,extractor='ffmpeg x11grab; tools/verify_music_runtime.py and tools/verify_music_player_paths.py',evidence='remake actual GUI capture; world/combat use prepared world save; not original parity')
# 與A案相同的HD面板及中文標籤；只比較按鈕，不把其他場景差異算入。
box=(1480,1144,1688,1200)
approved=Image.open(ROOT/'docs/verification/hd-music-control-20261013/a-hd.png').convert('RGB').crop(box)
actual=Image.open(OUT/'inn-controls-hd.png').convert('RGB').crop(box)
delta=np.abs(np.asarray(approved).astype(np.int16)-np.asarray(actual).astype(np.int16))
comparison=dict(max_channel_delta=int(delta.max()),different_pixels=int(np.count_nonzero(delta.max(2))))
# 面板四周透明邊保留各自背景；核對內部實際文字與面板。
interior=(8,12,200,56)
diff=ImageChops.difference(approved.crop(interior),actual.crop(interior))
comparison['interior_max_channel_delta']=max(max(pair) for pair in diff.getextrema())
left=np.asarray(approved.crop((20,14,188,50))).mean(2)>100
right=np.asarray(actual.crop((20,14,188,50))).mean(2)>100
dilate=lambda mask:np.asarray(Image.fromarray(mask.astype('uint8')*255).filter(ImageFilter.MaxFilter(3)))>0
comparison['glyph_pixels_outside_one_pixel_neighborhood']=int(np.count_nonzero(left&~dilate(right))+np.count_nonzero(right&~dilate(left)))
assert comparison['glyph_pixels_outside_one_pixel_neighborhood']==0,comparison
comparison['interpretation']='相同字模、標籤、座標與樣式；瀏覽器Canvas和Ebitengine線性取樣有字緣色差，幾何落在一像素內，不宣稱逐像素相同。'
write(DEST/'manifest.json',dict(schema='kol-hd-music-runtime-png/1',entries=entries,source_manifests=['../hd-interface-runtime-20261012/manifest.json','../../../assets/hd/first-slice/manifest.json','../../../assets/original-derived-png/ega/manifest.json'],control_prototype=comparison))
tools={str(p.relative_to(ROOT)):sha(p) for p in [ROOT/'tools'/n for n in ('verify_music_runtime.py','verify_music_player_paths.py','seal_hd_music.py','compose_hd_music.py','verify_adventure_music.py','verify_music_control_prototype.py')]}
receipt=dict(schema='kol-hd-music-runtime/1',status='PASS',date='2026-10-13',spec='docs/spec/004-ogg-audio-catalog.md#遊戲內配樂接入',decision='四首已聆聽核准；A獨立配樂預設開啟，F9與下緣按鈕；EGA/HD切換持續播放；本次遊玩有效',runtime=runtime,player_paths=paths,control_prototype=comparison,binary_sha256=sha(OUT/'kol'),program_inputs=program,tools=tools,catalog_sha256=sha(ROOT/'assets/audio/hd/manifest.json'),go_tests=['go test ./cmd/... ./internal/...','go test -tags sceneverify ./cmd/kol'],toolchains=['kol-package:1.26.7-ebiten2.9.9-20261011-r1','eob-audio-capture:20261009-r3','kol-music:timidity2.14-fluid3.1-r1'],original_parity_claimed=False,real_machine_listening_claimed=False,release_modified=False,limits=['HD無頭軟體繪圖與音訊佇列可能使實際循環等待長於牆鐘預估；收據以實際播放位置與輸出驗證。','HD重載圖逐通道最多1色階且每圖最多1000像素的GPU線性過濾舍入；精確差分另列。','探索與戰鬥由準備存檔開始，未驗自然全旅程；未新增Windows/macOS真機或Release驗收。'])
write(DEST/'receipt.json',receipt)
for n in ('test.log','test-sceneverify.log'):
    assert (OUT/n).is_file() and 'FAIL' not in (OUT/n).read_text()
    shutil.copyfile(OUT/n,DEST/n)
attachments={p.name:dict(sha256=sha(p),size=p.stat().st_size) for p in OUT.iterdir() if p.is_file() and p.name!='receipt.json'}
receipt['attachments']=attachments;write(OUT/'receipt.json',receipt)
print('SEALED PASS: four tracks, real audio, normal v3 save/reload, approved F9 control, six GUI PNGs',flush=True)
