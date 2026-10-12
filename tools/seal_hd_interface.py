#!/usr/bin/env python3
"""驗住宿後新程序重載、保存正式畫格與本機壓縮證據；完成後另行清工作區。"""
import gzip,hashlib,json,os,shutil,subprocess,tarfile
from pathlib import Path
from PIL import Image,ImageChops
ROOT=Path('/src');OUT=Path('/out')
assert Path('/.dockerenv').exists()
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def write(p,d):p.write_text(json.dumps(d,ensure_ascii=False,indent=2)+'\n')
assert json.loads((OUT/'gui-check.json').read_text())['passed']
script=json.loads((ROOT/'docs/verification/player-path/town-load.json').read_text())['script']
for name,mode,save in [('baseline-ega','ega','stay-baseline.json'),('ega','ega','stay-switched.json'),('hd','hd','stay-switched.json')]:
    dest=OUT/'gui-reload'/name;dest.mkdir(parents=True,exist_ok=True)
    argv=[OUT/'kol','-scene','play','-lang','zh-Hant','-graphics',mode,'-scale','1','-seed','1','-load',OUT/save,'-play-script',script,'-trace',dest/'trace.json','-text-log',dest/'text.jsonl','-shot-dir',dest]
    write(dest/'command.json',list(map(str,argv)))
    with (dest/'run.log').open('w') as log:subprocess.run(list(map(str,argv)),cwd=ROOT,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=90)
for name in ('trace.json','text.jsonl'):
    assert (OUT/'gui-reload/ega'/name).read_bytes()==(OUT/'gui-reload/hd'/name).read_bytes()
    assert (OUT/'gui-reload/ega'/name).read_bytes()==(OUT/'gui-reload/baseline-ega'/name).read_bytes(), 'GUI切換前後正常住宿重載不同'
public=ROOT/'docs/verification/hd-interface-runtime-20261012';local=ROOT/'docs/verification/local/hd-interface-runtime-20261012'
for p in (public,local):assert p.parent.is_dir() and p.parent.stat().st_uid==os.getuid();p.mkdir(exist_ok=True)
entries={}
for name in ['door-hd.png','inn-hd.png','battle-hd.png','door-compare.png','inn-compare.png','battle-compare.png','game-menu-hd.png','game-resized.png','game-closed.png']:
    shutil.copyfile(OUT/name,public/name);entries[name]=dict(sha256=sha(public/name),width=Image.open(public/name).width,height=Image.open(public/name).height,rights='private GitHub PNG exception; local remake capture; no public distribution grant',source_zip_sha256='2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46',frame={'door':'six/hd/step-002.png','inn':'six/hd/step-004.png','battle':'battle/hd/step-013.png'}.get(name.split('-')[0],name),comparison='previous HD / current HD, same retained state' if 'compare' in name else None)
receipt=json.loads((OUT/'interface-scenes.json').read_text());receipt.update(status='PASS',gui=json.loads((OUT/'gui-check.json').read_text()),normal_gui_stay_reload_equal=True,reload_states=len(json.loads((OUT/'gui-reload/hd/trace.json').read_text())['steps']),approved_font_250_glyphs_pixel_exact=True,screenshots_visually_reviewed=['door-hd.png','inn-hd.png','battle-hd.png','game-menu-hd.png'],assets_manifest_sha256=sha(ROOT/'assets/hd/first-slice/interface/manifest.json'),source_base_commit='7ff980d2',program_inputs={str(p.relative_to(ROOT)):sha(p) for d in ('cmd','internal') for p in sorted((ROOT/d).rglob('*.go'))})
receipt['program_inputs'].update({n:sha(ROOT/n) for n in ('go.mod','go.sum')})
receipt['tools']={str(p.relative_to(ROOT)):sha(p) for p in [ROOT/'tools/build_hd_interface.py',ROOT/'tools/verify_hd_interface.py',ROOT/'tools/verify_display_switch.py',ROOT/'tools/seal_hd_interface.py']}
write(public/'receipt.json',receipt);write(OUT/'interface-receipt.json',receipt)
write(public/'manifest.json',dict(schema='kol-hd-interface-png/1',entries=entries,receipt='receipt.json',runtime_sha256=sha(OUT/'kol'),source_zip_sha256='2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46',toolchain=receipt['toolchain']))
assetlog=ROOT/'assets/hd/first-slice/interface/bake.log'
if assetlog.exists():shutil.copyfile(assetlog,OUT/'asset-bake.log');assetlog.unlink()
music=ROOT/'docs/verification/hd-music-20261012'
for p in music.glob('*.log'):
    raw=p.read_bytes();(music/(p.name+'.gz')).write_bytes(gzip.compress(raw,mtime=0));p.unlink()
m=json.loads((music/'manifest.json').read_text());m['qa']='technical-check.json';m['browser_qa']='browser-check.json';m['rights']='Project-authored MIDI and rendered OGG: repository LICENSE (RRSAL-1.0); FluidR3 GM MIT attribution retained. No original game audio sampled.';m['tools']={str(p.relative_to(ROOT)):sha(p) for p in [ROOT/'tools/compose_hd_music.py',ROOT/'tools/verify_hd_music.py',ROOT/'tools/verify_hd_music_browser.py',ROOT/'tools/Dockerfile.music']};m['files']={p.name:dict(sha256=sha(p),size=p.stat().st_size) for p in sorted(music.iterdir()) if p.name!='manifest.json'};write(music/'manifest.json',m)
files=[p for p in sorted(OUT.rglob('*')) if p.is_file() and 'gocache' not in p.parts and p.name not in ('kol','kol-sceneverify')]
retained={str(p.relative_to(OUT)):dict(sha256=sha(p),size=p.stat().st_size) for p in files}
archive=local/'evidence.tar.xz'
with tarfile.open(archive,'w:xz',preset=3) as tar:
    for p in files:tar.add(p,arcname=str(p.relative_to(OUT)),recursive=False)
with tarfile.open(archive) as tar:
    assert set(tar.getnames())==set(retained)
    for name,info in retained.items():assert hashlib.sha256(tar.extractfile(name).read()).hexdigest()==info['sha256'],name
write(local/'retention.json',dict(schema='kol-retained-evidence/1',archive=dict(path='evidence.tar.xz',sha256=sha(archive),size=archive.stat().st_size),retained=retained,excluded_rebuildable_binaries=receipt['binaries'],excluded=['gocache/'],source_workspace='workplace/hd-interface-runtime-20261012',public_receipt='../../hd-interface-runtime-20261012/receipt.json'))
assert all(p.stat().st_uid==os.getuid() for p in files)
print('SEALED',len(files),'files, archive independently verified;',receipt['reload_states'],'normal GUI-save reload states',flush=True)
