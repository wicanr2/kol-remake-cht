"""RE-195正式成本接續、正常保存及歷史身分的唯讀重驗。"""
import argparse,base64,copy,hashlib,json,os,runpy
from pathlib import Path
ROOT=Path('/src')
PRIVATE=ROOT/'workplace/world-cost-resume-20261003'
HERE=ROOT/'docs/verification/world-cost-resume-20261003'
PARENT=ROOT/'docs/verification/world-return-cost-20261002'
CHANGED={'internal/world/session.go','cmd/kol/main.go','cmd/kol/battle_world_time_test.go'}
ADDED={'internal/world/return_cost.go','internal/world/return_cost_test.go'}
def read(p):return json.loads(p.read_text())
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def validate():
    preservation=read(PRIVATE/'parent-preservation.json');saved=preservation['sources']
    assert len(saved)==1358 and preservation['schema']=='kol-world-cost-resume-parent-v1'
    for name,digest in saved.items():assert sha(PRIVATE/'parent-inputs'/name)==digest,name
    parent=read(PARENT/'receipt.json');assert sha(PARENT/'receipt.json')==preservation['receipt_sha256']
    for name,digest in parent['attachments'].items():assert sha(PRIVATE/'parent-inputs'/name if name in saved else ROOT/name)==digest,name
    class ArchivedRoot:
        def __truediv__(self,name):return PRIVATE/'parent-inputs'/name if str(name) in saved else ROOT/name
    prior=runpy.run_path(str(PARENT/'verify_scratch.py'))['validate'];prior.__globals__['ROOT']=ArchivedRoot()
    assert prior()=={k:v for k,v in parent.items() if k!='attachments'}
    code={n for n in saved if n.startswith(('cmd/','internal/','tools/'))};assert len(code)==316
    assert {n for n in code if sha(ROOT/n)!=saved[n]}==CHANGED
    built=read(PRIVATE/'production-inputs.json');assert set(built['code_inputs'])==code|ADDED and len(built['code_inputs'])==318
    for name,digest in built['code_inputs'].items():assert sha(ROOT/name)==digest,name
    product=PRIVATE/'kol-world-cost-linux';probe=PRIVATE/'kol-settlement-probe2-linux';inn=PRIVATE/'kol-inn-probe2-linux'
    assert sha(product)==built['production_sha256']=='f4c4f70b82159b943a555bd855b4a65ca568c5dd25471544df17293032cfe960'
    assert sha(probe)=='24afa7803ea5b900b98adfbc0beed2899e007d53fc07a1e76826a3fe6352496d'
    assert sha(inn)=='523cff5c964e53ed99edb052706da778188a30b34e34d22cd3ced0c738859b51'
    assert built['image_id']=='sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa'
    assert product.stat().st_mode&0o777==0o755
    manifest=read(PRIVATE/'probe2-source.json')
    assert manifest['parent_sha256']==sha(ROOT/'cmd/kol/main.go') and manifest['probe_sha256']==sha(PRIVATE/'probe2-main-private.go')
    for logname in ['rules.log','return-cost-tests.log']:
        log=(PRIVATE/logname).read_text();assert 'FAIL' not in log and 'cmd/kol' in log
    focused=(PRIVATE/'return-cost-tests.log').read_text()
    assert '正反對照通過' in (PRIVATE/'worklist-selftest.log').read_text()
    assert '0 行沒有' in (PRIVATE/'worklist-scan.log').read_text()
    assert '未完成 22 條' in (PRIVATE/'worklist-verify.log').read_text()
    assert 'Read-only file system' in (PRIVATE/'worklist-render.log.attempt1').read_text()
    assert '有 4 行自承沒有條目認領' in (PRIVATE/'worklist-verify.log.attempt1').read_text()
    assert '新增六個精確訊號正對照及無缺口反對照通過' in (PRIVATE/'worklist-pattern-check.log').read_text()
    for test in ['TestReturnCostOriginalSlices','TestReturnCostReadsAfterWeatherAndClearsPendingIndex','TestUnresolvedReturnCostRemainsExplicit','TestBattleExitReadsActualScratchOnlyForItsPendingWalk']:
        assert f'--- PASS: {test}' in focused and f'--- SKIP: {test}' not in focused
    native=read(ROOT/'workplace/return-cost-gate-20261002/spell-alias-cases-private.json')
    assert len(native['rows'])==10
    for r in native['rows']:assert r['Cost']==r['Before']-r['After'] and r['Calls']==0 and r['Seed']==r['FinalSeed']
    assert 'field and method with the same name State' in (PRIVATE/'probe2-build.log.attempt1').read_text()
    normals=[];saves=[]
    for lang in ['en','zh-Hant']:
        folder=PRIVATE/('gui-cost-'+lang);trace=read(folder/'trace.json');rows=trace['steps']
        before=ROOT/'workplace/world-return-cost-20261002'/('gui-scratch-'+lang)
        plain=copy.deepcopy(trace)
        for row in plain['steps']:
            if row['step']>=480:row['world']['ticks']+=60
        assert plain==read(before/'trace.json') and len(rows)==493
        battle=[r for r in rows if 'battle' in r];assert len(battle)==362 and all(r['world']['ticks']==3360 for r in battle)
        returned=rows[480]['world'];assert returned['ticks']==3260 and returned['probe_return_scratch']==16
        assert returned['probe_world_seed']==3077853194 and returned['probe_world_roll_count']==1491
        assert returned['probe_world_last_roll']==dict(N=128,Before=4284445587,After=3077853194,Round=61,Stage=5)
        assert rows[-1]['world']['ticks']==3040
        for image in folder.glob('*.png'):assert sha(image)==sha(before/image.name),image.name
        assert len(list(folder.glob('*.png')))==493
        oldreload=PRIVATE/('production-reload-'+lang)
        baseline=ROOT/'workplace/battle-loot-20261002/fix2'/('production-reload2-'+lang)
        assert read(oldreload/'trace.json')==read(baseline/'trace.json')
        for image in oldreload.glob('*.png'):assert sha(image)==sha(baseline/image.name)
        lodging=PRIVATE/('fresh-inn-'+lang);keys=read(lodging/'driver-keys.json');innrows=read(lodging/'trace.json')['steps']
        prefix=read(folder/'command.json');prefix=prefix[prefix.index('-flow-script')+1].split()
        assert keys[:492]==prefix and len(innrows)==len(keys)+1 and len(keys)<1200
        assert innrows[:493]==rows
        savedfile=lodging/'complete.json';data=read(savedfile)
        assert data['schema']=='kol-save-v3' and len(data['characters'])==16 and data['party']==[0]
        assert data['settings']['faithful_inn_stay'] and data['location']==dict(kind='town',town=7,entry_point=3,x=57,y=8)
        byname={m['name']:m for m in data['characters'] if m is not None}
        for m in rows[480]['town']['probe_full_party']:
            s=byname[m['name']]
            assert s['carried']==[base64.b64decode(x).hex() for x in m['carried']],(lang,m['name'],'34槽')
            assert s['wounds']==m['wounds'] and s['gold']==m.get('gold',0)-(56 if m['name']=='Olear' else 0)
        observed=PRIVATE/('fresh-observation-reload-'+lang);formal=PRIVATE/('fresh-production-reload-'+lang)
        reloadrows=read(observed/'trace.json')['steps'];formalrows=read(formal/'trace.json')['steps']
        assert len(reloadrows)==len(formalrows)==16
        final=reloadrows[-1]['town']['probe_full_party'];assert [m['name'] for m in final]==['Theodoric','Olear']
        for m in final:
            s=byname[m['name']];assert [base64.b64decode(x).hex() for x in m['carried']]==s['carried']
            for key in ['id','wounds','place','town_code','building_id','sheet','quests','progress']:assert m[key]==s[key],(lang,m['name'],key)
            assert m.get('gold',0)==s['gold']
            for key in ['in_hand','backpack','worn']:
                simplify=lambda xs:[(x.get('name',''),x.get('record',''),x.get('condition',0)) if x else ('','',0) for x in xs]
                expected=s[key] if isinstance(s[key],list) else [s[key]]
                actual=m.get(key) if isinstance(m.get(key),list) else [m.get(key)]
                assert simplify(expected)==simplify(actual),(lang,m['name'],key)
        for row,other in zip(reloadrows,formalrows):
            plain=copy.deepcopy(row);plain['town'].pop('probe_full_party');assert plain==other
            name=f"step-{row['step']:03d}.png";assert sha(observed/name)==sha(formal/name)
        assert len(list(observed.glob('*.png')))==len(list(formal.glob('*.png')))==16
        comparable=copy.deepcopy(data);comparable.pop('saved_at');saves.append(comparable)
        normals.append(dict(language=lang,keys=492,states=493,images=493,battle_states=362,return_ticks=3260,
            final_ticks=3040,other_state_unchanged=True,images_unchanged=True,fresh_inn_keys=len(keys),fresh_inn_states=len(innrows),
            save_sha256=sha(savedfile),saved_characters=len(byname),saved_roster_slots=16,restored_inventory_slots=34,formal_reload_images=16))
    assert saves[0]==saves[1]
    return dict(schema='kol-world-cost-resume-receipt-v1',historical_shared_inputs=1358,parent_attachments=len(parent['attachments']),
        parent_results_recomputed=True,changed_code_inputs=sorted(CHANGED),added_code_inputs=sorted(ADDED),production_code_inputs=318,
        production_sha256=sha(product),observation_sha256=sha(probe),inn_observation_sha256=sha(inn),native_reader_cases=10,
        internal_and_cmd_tests='passed',normal=normals,fresh_lodging_save=True,complete_indexed_memory_implemented=False,
        normal_out_of_range_return_verified=False,full_entry_rng_verified=False,status='已證實成本讀取及正常敗北返回／新保存有限CONFORMED；完整表外讀取與父閘門保持開放。')
def attachments():
    names=set(read(PRIVATE/'parent-preservation.json')['sources'])|ADDED
    names|={str(p.relative_to(ROOT)) for p in PRIVATE.rglob('*') if p.is_file()}
    names|={'docs/re/20261003-world-cost-resume.md','docs/verification/world-cost-resume-20261003/verify_cost.py'}
    return {name:sha(ROOT/name) for name in sorted(names)}
if __name__=='__main__':
    ap=argparse.ArgumentParser();ap.add_argument('--seal',action='store_true');ap.add_argument('--audit',action='store_true');args=ap.parse_args()
    assert not(args.seal and args.audit);result=validate()
    if args.seal:
        result['attachments']=attachments();p=Path('/out/receipt.json');assert not p.exists()
        p.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    else:
        receipt=read(HERE/'receipt.json')
        for name,digest in receipt['attachments'].items():assert sha(ROOT/name)==digest,name
        assert result=={k:v for k,v in receipt.items() if k!='attachments'}
        if args.audit:
            control=read(PRIVATE/'host-control.json');assert control['git_diff_check']=='passed' and control['git_staged']==control['project_containers']==[]
            bad=[];md=[]
            for directory,dirs,files in os.walk(ROOT):
                for name in dirs+files:
                    p=Path(directory)/name
                    if p.lstat().st_uid==0:bad.append(str(p))
                    if p.is_dir() and p.name.endswith('.md'):md.append(str(p))
            assert not bad and not md
            for name in receipt['attachments']:
                q=(ROOT/name).stat();assert(q.st_uid,q.st_gid)==(1000,1000),name
            print(json.dumps(dict(schema='kol-world-cost-resume-audit-v1',receipt_sha256=sha(HERE/'receipt.json'),attachments=len(receipt['attachments']),
                full_readonly_verification='passed',host_control=control,root_owned_files=bad,md_directories=md,artifact_uid_gid=[1000,1000]),ensure_ascii=False))
    print(json.dumps({k:v for k,v in result.items() if k!='attachments'},ensure_ascii=False))
