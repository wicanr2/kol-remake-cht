#!/usr/bin/env python3
"""沿同一封存輸入驗HD介面；不重抽原版、不改RNG條件。"""
import hashlib,json,subprocess,tarfile
from pathlib import Path
from PIL import Image,ImageChops,ImageDraw
ROOT=Path('/src');OUT=Path('/out')
assert Path('/.dockerenv').exists()
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def write(p,d):p.write_text(json.dumps(d,ensure_ascii=False,indent=2)+'\n')
plans=[dict(name='six',script='L R RET SPACE',fixture='party6.json'),
       dict(name='mirror',script=json.loads((ROOT/'docs/verification/player-path/town-mirror.json').read_text())['script']),
       dict(name='custom',script='L R',fixture='custom.json'),
       dict(name='reload-v3',script='SPACE',fixture='old-v3.json'),
       dict(name='stay',script=json.loads((ROOT/'docs/verification/player-path/town-rest.json').read_text())['script'],save=True),
       dict(name='stay-reload',script=json.loads((ROOT/'docs/verification/player-path/town-load.json').read_text())['script'],reload=True),
       dict(name='battle',script=(OUT/'battle-script.txt').read_text().strip(),fixture='final.json',controlled=True)]
write(OUT/'interface-predeclared.json',dict(spec='docs/spec/039-hd-art-ui.md#操作介面與人物頭像候選',seed=1,plans=plans,original_parity_claimed=False,baseline='docs/verification/local/hd-runtime-20261011/evidence.tar.xz',input_sha256={p.name:sha(p) for p in (OUT/'inputs').iterdir()}))
oldroot=ROOT/'docs/verification/local/hd-runtime-20261011';idx=json.loads((oldroot/'retention.json').read_text());archive=oldroot/'evidence.tar.xz';assert sha(archive)==idx['archive']['sha256']
checks=[]
with tarfile.open(archive) as old:
    for plan in plans:
        for mode in ('ega','hd'):
            dest=OUT/plan['name']/mode;dest.mkdir(parents=True,exist_ok=True)
            binary=OUT/('kol-sceneverify' if plan.get('controlled') else 'kol')
            argv=[binary,'-lang','zh-Hant','-graphics',mode,'-scale','1','-seed','1','-scene','play','-play-script',plan['script'],'-shot-dir',dest,'-trace',dest/'trace.json','-text-log',dest/'text.jsonl']
            if plan.get('fixture'):argv+=['-load',OUT/'inputs'/plan['fixture']]
            if plan.get('reload'):argv+=['-load',OUT/'stay'/mode/'save.json']
            if (plan.get('fixture') or plan.get('save')) and not plan.get('controlled'):argv+=['-save',dest/'save.json']
            if plan.get('controlled'):argv+=['-verify-scenes','-verify-lock-hp','-verify-attack','20','-verify-world-start','48,9','-quest-sites',OUT/'inputs/sites-final.json']
            write(dest/'command.json',list(map(str,argv)))
            with (dest/'run.log').open('w') as log:subprocess.run(list(map(str,argv)),cwd=ROOT,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=120)
            print('CAPTURE',plan['name'],mode,flush=True)
        left=OUT/plan['name']/'ega';right=OUT/plan['name']/'hd'
        assert (left/'trace.json').read_bytes()==(right/'trace.json').read_bytes(),plan['name']+' trace changed'
        assert (left/'text.jsonl').read_bytes()==(right/'text.jsonl').read_bytes(),plan['name']+' text changed'
        frames=sorted(left.glob('step-*.png'));baseline_count=0
        for p in frames:
            member=plan['name']+'/ega/'+p.name
            try:oldbytes=old.extractfile(member).read()
            except KeyError:continue
            assert hashlib.sha256(oldbytes).hexdigest()==idx['retained'][member]['sha256']
            assert p.read_bytes()==oldbytes,member+' EGA changed'
            baseline_count+=1
        if (left/'save.json').exists():
            saves=[json.loads((d/'save.json').read_text()) for d in (left,right)]
            for s in saves:s.pop('saved_at',None)
            assert saves[0]==saves[1],plan['name']+' save changed'
        checks.append(dict(path=plan['name'],states=len(json.loads((left/'trace.json').read_text())['steps']),ega_frames_identical=baseline_count,state_and_text_equal=True))
    # 對比的兩邊沿同一HD人物、同一場景；左邊為先前首批HD，右邊為正式新介面。
    for name,member in [('door','six/hd/step-002.png'),('inn','six/hd/step-004.png'),('battle','battle/hd/step-013.png')]:
        image=Image.open(OUT/member).convert('RGB');image.save(OUT/(name+'-hd.png'))
        import io
        previous=Image.open(io.BytesIO(old.extractfile(member).read())).convert('RGB');assert previous.size==image.size==(1920,1200)
        pair=Image.new('RGB',(3840,1200));pair.paste(previous,(0,0));pair.paste(image,(1920,0));pair.save(OUT/(name+'-compare.png'))
receipt=dict(schema='kol-hd-interface-runtime/1',status='SCENES_PASS_GUI_PENDING',seed=1,checks=checks,original_parity_claimed=False,binaries={n:sha(OUT/n) for n in ('kol','kol-sceneverify')},toolchain='kol-package:1.26.7-ebiten2.9.9-20261011-r1',limits=['Controlled battle same retained seed 1, lock HP, attack 20; not original parity.','Original injury body, map, nonapproved command icons and complex subpages retained.','Linux Docker software rendering only.'])
write(OUT/'interface-scenes.json',receipt);print('SCENES PASS',len(checks),sum(c['states'] for c in checks),'states',flush=True)
