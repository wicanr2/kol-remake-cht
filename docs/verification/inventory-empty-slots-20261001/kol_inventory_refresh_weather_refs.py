from pathlib import Path
import copy,hashlib,json,os,shutil,subprocess,sys
sys.path.insert(0,'tools')
from png_io import read_png
from build_asset_contact_sheets import read_image
root=Path('workplace/inventory-empty-slots-20261001')
target=Path('docs/verification/inventory-empty-slots-20261001')
sha=lambda p:hashlib.sha256(Path(p).read_bytes()).hexdigest()
mapping={
 'docs/verification/remake-parity/shots/world/leave-brettle.png':('town-leave',12,'weather-leave.png'),
 'docs/verification/remake-parity/shots/world-clock/step14.png':('world-clock',26,'weather-clock-14.png'),
 'docs/verification/remake-parity/shots/world-clock/step15.png':('world-clock',27,'weather-clock-15.png'),
 'docs/verification/remake-parity/shots/world-quest/intro.png':('world-quest',89,'weather-quest-intro.png'),
 'docs/verification/remake-parity/shots/world/road-stand.png':('world-road',24,'weather-road-stand.png'),
 'docs/verification/remake-parity/shots/world/road-follow7.png':('world-road',31,'weather-road-follow7.png'),
 'docs/verification/remake-parity/shots/world/road-signpost.png':('world-road',33,'weather-road-signpost.png'),
}
sources=[next(Path('assets/original-derived-png/ega/WEATPICS').glob(f'weatpics-{i:03d}-*.png')) for i in range(18)]
raw=[read_image(p).rgb() for p in sources]
def composite(group,phase):
    if group in (3,4):return raw[group*3+phase%3]
    base,overlay,mask=raw[group*3:group*3+3];out=bytearray(base)
    for y in range(31):
        for x in range(64):
            src=(y*128+(x-phase+128)%128)*3
            if mask[src:src+3]==b'\0\0\0':
                i=(y*64+x)*3;out[i:i+3]=overlay[src:src+3]
    return bytes(out)
archive=root/'weather-reference-before';assert not archive.exists();archive.mkdir()
images=[]
for old,(name,step,filename) in mapping.items():
    current=root/'all-before'/name/f'step-{step:03d}.png';dest=target/filename
    assert not dest.exists()
    a,b=read_png(old),read_png(current);assert(a.width,a.height)==(b.width,b.height)==(320,200)
    x,y=a.rgb(),b.rgb();changed=[i//3 for i in range(0,len(x),3) if x[i:i+3]!=y[i:i+3]]
    assert changed and all(16<=i%320<80 and 76<=i//320<138 for i in changed),(old,'天候框外差異')
    candidates=[{'group':g,'phase':p} for g in range(6) for p in range(64) if composite(g,p)==b.crop(16,76,64,62).rgb()]
    assert candidates
    saved=archive/filename;shutil.copy2(old,saved);shutil.copy2(current,dest)
    images.append({'old':old,'old_sha256':sha(old),'new':str(dest),'new_sha256':sha(dest),
                    'source_capture':str(current),'step':step,'weather_pixel_differences':len(changed),
                    'outside_weather_differences':0,'current_weather_candidates':candidates})
replacements={r['old']:r['new'] for r in images}
region_results=[]
for name in ('world','world-road','world-clock','world-quest'):
    source=Path('docs/verification/remake-parity')/('plan-'+name+'.json')
    plan=json.loads(source.read_text());shutil.copy2(source,archive/source.name)
    for pair in plan['pairs']:
        if pair['remake'] in replacements:pair['remake']=replacements[pair['remake']]
    plan['metadata']={'scope':'RE-164修正後的正常截圖，原版圖、比較矩形、游標例外與必須相同區均沿用原計畫；不宣稱天候框的原版GUI同狀態對拍',
                      'previous_plan':str(source),'previous_plan_sha256':sha(source),
                      'weather_contract':'docs/spec/026-weather-animation.md','rights':'衍生PNG只適用private GitHub特例'}
    p=target/('region-plan-'+name+'.json');assert not p.exists();p.write_text(json.dumps(plan,ensure_ascii=False,indent=2)+'\n')
    receipt=target/('region-receipt-'+name+'.json')
    result=subprocess.run(['python3','tools/verify_remake_parity.py','--plan',str(p),'--out',str(receipt)],capture_output=True,text=True)
    assert result.returncode==0,result.stdout+result.stderr
    data=json.loads(receipt.read_text());assert data['failure_count']==0
    originals={r['original']:sha(r['original']) for r in plan['pairs']}
    region_results.append({'name':name,'pairs':data['pair_count'],'failures':0,'plan':str(p),'plan_sha256':sha(p),
                           'receipt':str(receipt),'receipt_sha256':sha(receipt),'original_images_sha256':originals})
updates=[]
for name in ('town-leave','town-round-trip','world-clock','world-quest','world-road'):
    path=Path('docs/verification/player-path')/(name+'.json');plan=json.loads(path.read_text());changed=[]
    for checkpoint in plan.get('checkpoints',[]):
        original=checkpoint.get('identical_to')
        if original in replacements:
            checkpoint['identical_to']=replacements[original];changed.append(checkpoint['step'])
    assert changed
    plan['reference_correction']='RE-164／SPEC-026天候修正後的正常截圖；原版已驗區域另見inventory-empty-slots-20261001/region-receipt-*.json，舊圖與原始比較條件保留。'
    path.write_text(json.dumps(plan,ensure_ascii=False,indent=2)+'\n');updates.append({'name':name,'steps':changed,'plan_sha256':sha(path)})
summary={'schema':'kol-weather-reference-correction-v1','images':images,'original_region_rechecks':region_results,
         'player_plan_updates':updates,'contract':'RE-164／SPEC-026：正常進野外組0，先旋轉一格再畫',
         'weather_sources_sha256':{str(p):sha(p) for p in sources},
         'rights':'原作衍生PNG只限既有private GitHub特例；舊截圖不移動、不覆寫',
         'limitations':'原版比较仍仅为原计划的矩形区域；天候相位只沿用既有53次常式对拍，不新增原版GUI／硬體時序收據。'}
summary['limitations']='原版比較仍僅為原計畫的矩形區域；天候相位只沿用既有53次常式對拍，不新增原版GUI／硬體時序收據。'
(root/'weather-reference-correction.json').write_text(json.dumps(summary,ensure_ascii=False,indent=2)+'\n')
print('七張正常天候引用已修正，原比較區域重新核實',sum(r['pairs'] for r in region_results),'組通過；舊圖未改。')
