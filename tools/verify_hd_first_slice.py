#!/usr/bin/env python3
"""Docker內重播首批HD，核對EGA不變、狀態相同與替換區域；SPEC-039/040。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tarfile
import time
from PIL import Image, ImageChops, ImageDraw

ROOT, OUT = Path('/src'), Path('/out')
assert Path('/.dockerenv').exists()
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()


def write(p, data):
    p.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')


def run(argv, log):
    with log.open('w') as f:
        subprocess.run(list(map(str, argv)), cwd=ROOT, stdout=f, stderr=subprocess.STDOUT,
                       check=True, timeout=120)


def prepare():
    (OUT/'inputs').mkdir(exist_ok=True)
    argv=['python3','-B',ROOT/'tools/import_original_characters.py','--zip',ROOT/'Knights-of-legend.zip',
          '--repo',ROOT,'--names','Theodoric','Poontang','Olear','Moraine','Fighting Bull','Kenneth',
          '--location','7,3,57,8','--saved-at','2026-10-11T00:00:00Z','--out',OUT/'inputs/party6.json']
    run(argv,OUT/'import-six.log')
    retained=ROOT/'docs/verification/local/controlled-scenes-20261011'
    index=json.loads((retained/'retention.json').read_text())
    assert sha(retained/'evidence.tar.xz')==index['archive']['sha256']
    with tarfile.open(retained/'evidence.tar.xz') as archive:
        for name in ('inputs/final.json','inputs/sites-final.json','battle-script.txt'):
            data=archive.extractfile(name).read()
            assert hashlib.sha256(data).hexdigest()==index['retained'][name]['sha256']
            (OUT/name).write_bytes(data)
    source=ROOT/'dist-all/v.1.0.0-20261011/smoke/save.dat'
    shutil.copy2(source,OUT/'inputs/old-v3.json')
    # v1匯入會重建原版完整角色；自訂圖需放在實際可保存picture的v3欄位。
    custom=json.loads(source.read_text())
    custom['characters'][custom['party'][0]]['picture']='01'+'00'*759
    write(OUT/'inputs/custom.json',custom)
    write(OUT/'predeclared.json',dict(schema='kol-hd-runtime/1',baseline='8a6b52da',
        seed=1,graphics='hd',resolution=[1920,1200],spec='docs/spec/039-hd-art-ui.md#首批接入契約',
        accepted_prototype='docs/verification/hd-first-slice-20261011/manifest.json',
        imported_command=list(map(str,argv)),old_v3_source=str(source.relative_to(ROOT)),old_v3_sha256=sha(source),
        substituted_custom='現行交付v3第一位characters[party[0]].picture改為合法760-byte單點圖；確認不誤套HD',
        original_parity_claimed=False,inputs={p.name:sha(p) for p in (OUT/'inputs').iterdir()}))
    print('PREPARED six actual party members, custom fallback, old v3 and retained battle',flush=True)


def captures():
    plans=[dict(name='full',scene='flow',script=json.loads((ROOT/'docs/verification/player-path/full.json').read_text())['script']),
           dict(name='six',scene='play',script='L R RET SPACE',fixture='party6.json'),
           dict(name='mirror',scene='play',script=json.loads((ROOT/'docs/verification/player-path/town-mirror.json').read_text())['script']),
           dict(name='custom',scene='play',script='L R',fixture='custom.json'),
           dict(name='reload-v3',scene='play',script='SPACE',fixture='old-v3.json'),
           dict(name='stay',scene='play',script=json.loads((ROOT/'docs/verification/player-path/town-rest.json').read_text())['script'],save=True),
           dict(name='stay-reload',scene='play',script=json.loads((ROOT/'docs/verification/player-path/town-load.json').read_text())['script'],reload=True),
           dict(name='battle',scene='play',script=(OUT/'battle-script.txt').read_text().strip(),fixture='final.json',controlled=True)]
    write(OUT/'plans.json',plans)
    for plan in plans:
        if args.only and plan['name']!=args.only:continue
        modes=('ega','hd') if plan.get('controlled') else ('baseline','ega','hd')
        for mode in modes:
            dest=OUT/plan['name']/mode
            dest.mkdir(parents=True,exist_ok=True)
            binary='kol-sceneverify' if plan.get('controlled') else 'kol-baseline' if mode=='baseline' else 'kol'
            argv=[OUT/binary,'-lang','zh-Hant','-scale','1','-seed','1','-scene',plan['scene'],
                  '-flow-script' if plan['scene']=='flow' else '-play-script',plan['script'],
                  '-shot-dir',dest,'-trace',dest/'trace.json','-text-log',dest/'text.jsonl']
            if mode!='baseline':argv+=['-graphics',mode]
            if plan.get('fixture'):argv+=['-load',OUT/'inputs'/plan['fixture']]
            if plan.get('reload'):argv+=['-load',OUT/'stay'/mode/'save.json']
            if (plan.get('fixture') or plan.get('save')) and not plan.get('controlled'):
                argv+=['-save',dest/'save.json']
            if plan.get('controlled'):
                argv+=['-verify-scenes','-verify-lock-hp','-verify-attack','20','-verify-world-start','48,9',
                       '-quest-sites',OUT/'inputs/sites-final.json']
            run(argv,OUT/(plan['name']+'-'+mode+'.log'))
            steps=json.loads((dest/'trace.json').read_text())['steps']
            keep={f'step-{row["step"]:03d}.png' for row in steps}
            for extra in dest.glob('*.png'):
                if extra.name not in keep:
                    retained=OUT/'input-error'
                    retained.mkdir(exist_ok=True)
                    extra.rename(retained/(plan['name']+'-'+mode+'-earlier-'+extra.name))
            write(dest/'command.json',list(map(str,argv)))
            print('CAPTURE',plan['name'],mode,len(list(dest.glob('*.png'))),flush=True)


def gui():
    argv=[str(OUT/'kol'),'-scene','game','-lang','zh-Hant','-graphics','hd','-scale','1',
          '-load',str(OUT/'six/hd/save.json'),'-save',str(OUT/'gui-save.json'),'-seed','1','-log-input']
    commands=[]
    def xdo(*args):
        command=['xdotool',*map(str,args)]
        commands.append(command)
        return subprocess.check_output(command,text=True,timeout=10).strip()
    log=OUT/'gui-input.log'
    (OUT/'gui-check.json').unlink(missing_ok=True)
    with log.open('w') as f:
        proc=subprocess.Popen(argv,cwd=ROOT,stdout=f,stderr=subprocess.STDOUT)
        try:
            window=None
            for _ in range(80):
                time.sleep(.1)
                found=subprocess.run(['xdotool','search','--name','武士傳說'],capture_output=True,text=True)
                if found.returncode==0:
                    window=found.stdout.splitlines()[-1]
                    break
                assert proc.poll() is None,'interactive startup failed'
            assert window,'window not found'
            xdo('windowfocus','--sync',window)
            time.sleep(.8)
            xdo('key','--window',window,'Left','Right')
            time.sleep(.4)
            xdo('mousemove','--window',window,264,1086)
            xdo('click','1')
            time.sleep(.4)
            xdo('key','--window',window,'space')
            time.sleep(.3)
            xdo('mousemove','--window',window,500,150)
            xdo('click','1')
            time.sleep(.4)
            run(['ffmpeg','-y','-loglevel','error','-f','x11grab','-video_size','1920x1200',
                 '-i',':99','-frames:v','1','-update','1',OUT/'gui.png'],OUT/'gui-capture.log')
            xdo('windowsize',window,960,600)
            time.sleep(.5)
            xdo('mousemove','--window',window,120,75)
            xdo('click','1')
            time.sleep(.4)
            # 保持修飾鍵跨過至少一個更新幀，避免12ms合鍵在讀取前已放開。
            xdo('keydown','--window',window,'ctrl')
            time.sleep(.1)
            xdo('keydown','--window',window,'q')
            proc.wait(timeout=10)
            # 退出已關窗；在X server釋放按鍵，不再向舊window送事件。
            xdo('keyup','q','ctrl')
            assert proc.returncode==0,proc.returncode
        finally:
            if proc.poll() is None:
                proc.terminate()
                proc.wait(timeout=5)
    text=log.read_text()
    for marker in ('游標=(132,543) 可點=2 token=[RET]',
                   '游標=(250,75) 可點=7 token=[2]',
                   '游標=(120,75) 可點=11 token=[1]'):
        assert marker in text,(marker,text)
    assert not (OUT/'gui-save.json').exists(),'clicks must not save'
    write(OUT/'gui-check.json',dict(passed=True,command=argv,events=commands,
          expected_logical_positions=[[132,543],[250,75],[120,75]],
          window_sizes=[[1920,1200],[960,600]],screenshot_sha256=sha(OUT/'gui.png')))
    print('GUI PASS native and resized window clicks',flush=True)


def seal():
    checks=[]
    plans=json.loads((OUT/'plans.json').read_text())
    selected=[]
    for plan in plans:
        name=plan['name']
        ega,hd=OUT/name/'ega',OUT/name/'hd'
        e=json.loads((ega/'trace.json').read_text())
        h=json.loads((hd/'trace.json').read_text())
        assert e==h,(name,'state changed')
        assert (ega/'text.jsonl').read_bytes()==(hd/'text.jsonl').read_bytes(),(name,'text changed')
        before=OUT/name/'baseline'
        if before.exists():
            assert json.loads((before/'trace.json').read_text())==e,(name,'baseline state changed')
        changed=[]
        for p in sorted(ega.glob('*.png')):
            if before.exists():assert sha(before/p.name)==sha(p),(name,p.name,'EGA changed')
            expected=Image.open(p).convert('RGBA').resize((1920,1200),Image.Resampling.NEAREST)
            actual=Image.open(hd/p.name).convert('RGBA')
            assert actual.size==(1920,1200)
            delta=ImageChops.difference(expected,actual)
            # alpha相同時RGBA.getbbox會忽略RGB差異，分通道合併。
            diff=ImageChops.lighter(ImageChops.lighter(delta.getchannel(0),delta.getchannel(1)),delta.getchannel(2))
            mask=Image.new('L',(1920,1200),0)
            draw=ImageDraw.Draw(mask)
            for i in range(6):draw.rectangle((88+i*288,10,335+i*288,241),fill=255)
            draw.rectangle((624,348,1775,923),fill=255)
            draw.rectangle((100,80,815,453),fill=255)
            assert not ImageChops.subtract(diff,mask).getbbox(),(name,p.name,'outside approved areas')
            if diff.getbbox():changed.append(p.name)
        if name in ('six','full','battle'):assert changed,(name,'HD assets not displayed')
        if name=='six':
            old=Image.open(ega/'step-002.png').convert('RGB').resize((1920,1200),Image.Resampling.NEAREST)
            new=Image.open(hd/'step-002.png').convert('RGB')
            for i in range(6):
                rect=(88+i*288,10,336+i*288,242)
                assert ImageChops.difference(old.crop(rect),new.crop(rect)).getbbox(),('missing HD member',i)
            assert ImageChops.difference(old.crop((624,348,1776,924)),new.crop((624,348,1776,924))).getbbox(),'missing HD inn'
            selected.append(('inn-door',hd/'step-002.png'))
        if name=='custom':
            for p in sorted(ega.glob('*.png')):
                a=Image.open(p).convert('RGB').resize((1920,1200),Image.Resampling.NEAREST)
                b=Image.open(hd/p.name).convert('RGB')
                assert not ImageChops.difference(a.crop((84,6,342,300)),b.crop((84,6,342,300))).getbbox(),'custom portrait overwritten'
        checks.append(dict(name=name,states=len(e['steps']),state_equal=True,text_equal=True,
                           baseline_ega_unchanged=before.exists(),frames=len(list(ega.glob('*.png'))),changed_frames=changed))
        if changed:selected.append((name,hd/changed[-1]))
    declared=json.loads((OUT/'predeclared.json').read_text())
    for name,h in declared['inputs'].items():assert sha(OUT/'inputs'/name)==h,('input changed',name)
    assert not list((OUT/'inputs').glob('*.v2.json')),'scene viewing wrote a save'
    saves={}
    for mode in ('baseline','ega','hd'):
        save=json.loads((OUT/'stay'/mode/'save.json').read_text())
        assert save['schema']=='kol-save-v3'
        save.pop('saved_at',None)
        saves[mode]=save
    assert saves['baseline']==saves['ega']==saves['hd'],'HD changed saved state'
    public=ROOT/'docs/verification/hd-runtime-20261011'
    images=[]
    for name,p in selected:
        dest=public/(name+'.png')
        shutil.copy2(p,dest)
        images.append(dict(file=dest.name,sha256=sha(dest),frame=str(p.relative_to(OUT)),
                           source_entry='SPEC-039 runtime，核准素材＋既有EGA畫面',
                           source_sha256=json.loads((ROOT/'assets/hd/first-slice/manifest.json').read_text())['source_zip_sha256'],
                           rights='private GitHub 可保存；無公開散布權',evidence='受控場景驗收'))
    receipt=dict(schema='kol-hd-runtime/1',date='2026-10-11',spec='docs/spec/039-hd-art-ui.md#首批接入契約',
        scope='首批1920×1200呈現，非全804張HD完成',original_parity_claimed=False,
        predeclared=declared,checks=checks,reviewed_pngs=images,
        program_inputs={str(p.relative_to(ROOT)):sha(p) for d in ('cmd','internal') for p in sorted((ROOT/d).rglob('*.go'))},
        files={str(p.relative_to(ROOT)):sha(p) for p in sorted((ROOT/'assets/hd/first-slice').iterdir())},
        binaries={p:sha(OUT/p) for p in ('kol','kol-baseline','kol-sceneverify')},
        tools={p:sha(ROOT/p) for p in ('tools/build_hd_first_slice.py','tools/verify_hd_first_slice.py')})
    receipt['program_inputs'].update({p:sha(ROOT/p) for p in ('go.mod','go.sum')})
    receipt['toolchain']='kol-package:1.26.7-ebiten2.9.9-20261011-r1 / Go1.26.7 / Ebitengine2.9.9 / Xvfb / xdotool'
    receipt['normal_stay_and_reload_save_equal_except_saved_at']=True
    receipt['gui']=json.loads((OUT/'gui-check.json').read_text())
    assert receipt['gui']['passed']
    write(public/'receipt.json',receipt)
    write(public/'manifest.json',dict(schema='kol-hd-runtime-png/1',entries=images,
          toolchain=receipt['toolchain'],runtime_sha256=receipt['binaries']['kol'],
          receipt='receipt.json',approved_assets='../../../assets/hd/first-slice/manifest.json'))
    receipt['attachments']={str(p.relative_to(OUT)):dict(size=p.stat().st_size,sha256=sha(p))
                            for p in sorted(OUT.rglob('*')) if p.is_file() and p.name!='receipt.json'}
    write(OUT/'receipt.json',receipt)
    print('SEALED',len(checks),'paths',sum(c['frames'] for c in checks),'HD frames',flush=True)


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('phase',choices=['prepare','captures','gui','seal'])
    parser.add_argument('--only',choices=['full','six','mirror','custom','reload-v3','stay','stay-reload','battle'])
    args=parser.parse_args()
    assert not args.only or args.phase=='captures','--only僅限重跑已修正的單一捕圖輸入'
    globals()[args.phase]()
