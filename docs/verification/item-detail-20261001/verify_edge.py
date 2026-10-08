"""英文快取還原：兩條正常原版完整畫面、雙語狀態與保存／重載核對。"""
from pathlib import Path
import argparse,hashlib,importlib.util,json,shutil,sys
ROOT,OUT=Path('/src'),Path('/out')
sys.path.insert(0,str(ROOT/'tools'))
from png_io import Image,read_png
AFTER='fb185a0a1a9217feb25b66b2458cde1f6c15ae328f87ed79a9b6d291b57f5fba'
BEFORE='821ee273ce3d78f658d4efa83b2716ec89a12e4002fa0365b7c8bf17fc86b78a'
STAGES=[15,28,42,56,72,88]
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def load(p):return json.loads(p.read_text())
def project_path(p):
    if p.is_relative_to(OUT):return str(Path('workplace/item-detail-20261001')/p.relative_to(OUT))
    return str(p.relative_to(ROOT))
def main():
    ap=argparse.ArgumentParser();ap.add_argument('--export',action='store_true');args=ap.parse_args()
    proof=Path('/proof') if args.export else ROOT/'docs/verification/item-detail-20261001'
    assert sha(OUT/'kol-items-edge-final-linux')==AFTER and sha(OUT/'kol-items-mark-linux')==BEFORE
    assert sha(proof/'mark-receipt.json')=='75616fa5be90046faff4e213412b57e4bee0a1a3514fed7e50fcda0f997b78e8'
    old=load(proof/'mark-receipt.json');source=load(proof/'edge-source-receipt.json')
    for f in source['source_inputs']:
        # 原版與證據不變；此來源收據保留實作前的產品雜湊，不重新換入新程式。
        if f['path'].startswith(('internal/','cmd/')):continue
        assert sha(ROOT/f['path'])==f['sha256'],f['path']
    historical=[]
    for name in ['edge-source-receipt.json','mark-receipt.json','cjk-figure-receipt.json','mirror-receipt.json','actions-receipt.json','all-slots-receipt.json','receipt.json']:
        d=load(proof/name)
        for f in d['attachments']:assert sha(proof/f['path'])==f['sha256']
        historical.append(dict(path=name,sha256=sha(proof/name),attachments=len(d['attachments'])))
    spec=importlib.util.spec_from_file_location('historical_actions',proof/'verify_actions.py')
    legacy=importlib.util.module_from_spec(spec);spec.loader.exec_module(legacy)
    mapping={'.':'>','Return':'RET','Down':'D','Up':'U'}
    checks=[];changed=[];copied=[]
    for prefix,count,oracle,keys_file,previous in [
        ('normal-edge-final-',104,'gui-edge2','actions-keys.json','normal-mark-'),
        ('normal-edge-final-slots-',61,'gui3','all-keys.json','normal-all-')]:
        original_keys=load(OUT/keys_file)
        assert len(original_keys)==count
        expected=['SPACE','SPACE','6','RET','SPACE']+[mapping[k] for k in original_keys]
        traces=[]
        for lang in ['en','zh-Hant']:
            path=OUT/(prefix+lang);cmd=load(path/'command.json')
            assert cmd[cmd.index('-flow-script')+1].split()==expected
            assert cmd[cmd.index('-seed')+1]=='1'
            trace=load(path/'trace.json')['steps'];assert len(trace)==count+6
            assert trace==load(OUT/(previous+lang)/'trace.json')['steps']
            traces.append(trace)
        assert traces[0]==traces[1]
        equal=[]
        for frame in range(count):
            original=read_png(OUT/oracle/f'item-{frame:02d}.png')
            current=read_png(OUT/(prefix+'en')/f'step-{frame+6:03d}.png')
            assert (original.width,original.height)==(current.width,current.height)==(320,200)
            points,mouse=legacy.diff(original,current)
            assert not points and mouse==67,(prefix,frame,points[:10],mouse)
            equal.append(frame)
        checks.append(dict(run=prefix,original_frames=count,full_native_frames_equal=equal,
                           original_keys=original_keys,remake_keys=len(expected),language_states_equal=count+6))
    for step in range(110):
        current=read_png(OUT/'normal-edge-final-en'/f'step-{step:03d}.png')
        previous=read_png(OUT/'normal-mark-en'/f'step-{step:03d}.png')
        cr,pr=current.rgb(),previous.rgb()
        points=[[n%320,n//320] for n in range(64000) if cr[n*3:n*3+3]!=pr[n*3:n*3+3]]
        assert all(x==15 for x,y in points)
        if points:changed.append(dict(step=step,pixels=points))
        current=read_png(OUT/'normal-edge-final-zh-Hant'/f'step-{step:03d}.png')
        previous=read_png(OUT/'normal-mark-zh-Hant'/f'step-{step:03d}.png')
        assert current.rgb()==previous.rgb(),('繁中不應改變',step)
    assert len(changed)==70
    # 真正的首版失敗：漏圖示底板覆蓋，不是依同一段實作製造預期值。
    original=read_png(OUT/'gui-edge2/item-100.png')
    first=read_png(OUT/'normal-edge-en/step-106.png')
    points,_=legacy.diff(original,first);assert points==[[15,164],[15,165]]
    # 舊每幀重畫的實際畫格在70頁被拒絕，包含換列與完整返回兩種生命週期。
    assert [c['step']-6 for c in changed]==[c['original_frame'] for c in old['unverified_list_frames']]
    for c in old['unverified_list_frames']:
        points,_=legacy.diff(read_png(OUT/'gui-edge2'/f"item-{c['original_frame']:02d}.png"),
                            read_png(OUT/'normal-mark-en'/f"step-{c['remake_step']:03d}.png"))
        assert points==c['pixels']
    # 原版以外任一框線像素損壞也不應被完整比較吞掉。
    good=read_png(OUT/'normal-edge-final-en/step-106.png').rgb()
    bad=bytearray(good);n=(2*320+15)*3;bad[n:n+3]=bytes(3)
    points,_=legacy.diff(original,Image(320,200,3,bytes(bad)))
    assert points==[[15,2]],'損壞框線反例未經正式完整比較拒絕'
    persistence=[];old_saves=[];reloads=[]
    for stage in STAGES:
        for mode in ['save','reload']:
            languages=[]
            for lang in ['en','zh-Hant']:
                current=OUT/f'{mode}-edge-final-{stage:02d}-{lang}/save.json'
                previous=OUT/f'{mode}-mark-{stage:02d}-{lang}/save.json'
                a,b=load(current),load(previous);a.pop('saved_at');b.pop('saved_at')
                assert a==b,(mode,stage,lang,'持久資料改變')
                languages.append(a)
                persistence.append(dict(path=project_path(current),sha256=sha(current),stage=stage,language=lang,mode=mode))
                old_saves.append(dict(path=project_path(previous),sha256=sha(previous)))
            assert languages[0]==languages[1]
        for lang in ['en','zh-Hant']:
            current=OUT/f'reload-edge-final-{stage:02d}-{lang}'
            previous=OUT/f'reload-mark-{stage:02d}-{lang}'
            cmd=load(current/'command.json');step=int(cmd[cmd.index('-shot-steps')+1].split(',')[0])
            frame=f'step-{step:03d}.png'
            assert read_png(current/frame).rgb()==read_png(previous/frame).rgb()
            reloads.append(dict(stage=stage,language=lang,step=step))
    assert len(persistence)==24 and len(reloads)==12
    assert load(OUT/'edge-final-persistence-comparison.json')['languages_equal']
    origins=load(OUT/'edge-final-script-origins.json')
    for f in origins:
        assert sha(OUT/f['source'])==f['source_sha256'] and sha(OUT/f['path'])==f['sha256']
    current_inputs=sorted(set(f['path'] for f in old['current_inputs'])|{
        'internal/ui/itempage_edge.go','docs/re/item-edge-address-index.json',
        'docs/verification/item-detail-20261001/verify_edge.py'})
    current_inputs=[dict(path=n,sha256=sha(ROOT/n)) for n in current_inputs]
    source_names=['gui-edge2/item-calls-private.json','gui-edge2/inputs.json','gui3/item-calls-private.json','gui3/inputs.json',
                  'edge-final-script-origins.json','tests-edge-final.log','edge-final-persistence-comparison.json',
                  'kol-items-edge-linux','item-edge-annotated-private.json']
    for prefix,count,oracle,_,_ in [('normal-edge-final-',104,'gui-edge2',None,None),('normal-edge-final-slots-',61,'gui3',None,None)]:
        source_names +=[f'{oracle}/item-{i:02d}.png' for i in range(count)]
        for lang in ['en','zh-Hant']:
            source_names +=[f'{prefix}{lang}/step-{i:03d}.png' for i in range(count+6)]
            source_names +=[f'{prefix}{lang}/command.json',f'{prefix}{lang}/trace.json']
    source_inputs=[dict(path=project_path(OUT/n),sha256=sha(OUT/n)) for n in sorted(set(source_names))]
    for frame in [3,4,15,100,101]:
        for role,folder,filename in [('original','gui-edge2',f'item-{frame:02d}.png'),
                ('remake-en','normal-edge-final-en',f'step-{frame+6:03d}.png'),
                ('remake-zh-Hant','normal-edge-final-zh-Hant',f'step-{frame+6:03d}.png')]:
            source_png=OUT/folder/filename;name=f'edge-final-{role}-{frame:03d}.png'
            if args.export:
                assert not (proof/name).exists();shutil.copyfile(source_png,proof/name)
            assert sha(proof/name)==sha(source_png)
            copied.append(dict(path=name,sha256=sha(source_png),role=role,frame=frame,
                source_entry='KOLTOWN.EXE正常書本／旅店／物品頁' if role=='original' else 'kol-items-edge-final-linux正常輸入',
                source_sha256='c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221' if role=='original' else AFTER,
                tool='dosgolem 968ca774705853276ae55c8d1163c369481ac71f' if role=='original' else 'Go 1.26.7／Ebitengine 2.9.9',
                evidence_status='英文完整正常畫面已對拍；繁中與前版一致',rights='private GitHub可保存；不得公開散布'))
    result=dict(schema='kol-item-edge-v1',date='2026-10-01',status='英文快取還原正常窄切片通過',
        binary=dict(path='workplace/item-detail-20261001/kol-items-edge-final-linux',sha256=AFTER),before_binary_sha256=BEFORE,
        original_revision=source['original_revision'],original_seed=1,remake_seed=1,
        seed_setup='原版正常旅店後DS:1600=1、DS:1602=0；重製版-seed 1',normal_startup=True,state_injection=False,direct_entry=False,
        checks=checks,changed_english_frames=changed,chinese_frames_unchanged=110,original_full_frames_equal=165,
        persistence_checks=persistence,previous_persistence_locked=old_saves,reload_detail_frames_unchanged=reloads,
        current_inputs=current_inputs,source_inputs=source_inputs,runner_scripts=origins,historical=historical,attachments=copied,
        negative_controls=['實際前版70個清單失敗頁拒絕','實際首版漏圖示底板的兩像素拒絕','完整框線單像素損壞拒絕'],
        source_receipt=dict(path='edge-source-receipt.json',sha256=sha(proof/'edge-source-receipt.json')),
        verifier_sha256=sha(Path(__file__)),
        limits='Theodoric兩條正常物品序列；繁中只驗前後一致。其他人物／野外／全部物品動作仍未全驗，原版完整保存／時間／金幣未對拍。')
    target=proof/'edge-receipt.json'
    if args.export:
        assert not target.exists();target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    else:assert load(target)==result,'來源或產品收據重驗不一致'
    print('165個完整英文原版畫格零差；110繁中畫格不變；24持久資料及12重載明細不變。')
if __name__=='__main__':main()
