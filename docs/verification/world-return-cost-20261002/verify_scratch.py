"""RE-194正式暫存寫入與正常路徑的封存／唯讀重驗。"""
import argparse,copy,hashlib,json,os,runpy
from pathlib import Path
ROOT=Path('/src')
PRIVATE=ROOT/'workplace/world-return-cost-20261002'
HERE=ROOT/'docs/verification/world-return-cost-20261002'
PARENT=ROOT/'docs/verification/return-indexed-read-20261002'
CHANGED={f'internal/battle/{n}.go' for n in ['state','fatigue','execution','planning','items','flee','loot']}
ADDED={'internal/battle/return_scratch.go','internal/battle/return_scratch_test.go'}
def read(p):return json.loads(p.read_text())
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def validate():
    preserve=read(PRIVATE/'parent-preservation.json');saved=preserve['sources']
    assert preserve['schema']=='kol-world-return-cost-parent-v1' and len(saved)==1352
    for name,digest in saved.items():assert sha(PRIVATE/'parent-inputs'/name)==digest,name
    parent=read(PARENT/'receipt.json');assert sha(PARENT/'receipt.json')==preserve['receipt_sha256']
    for name,digest in parent['attachments'].items():
        assert sha(PRIVATE/'parent-inputs'/name if name in saved else ROOT/name)==digest,name
    class ArchivedRoot:
        def __truediv__(self,name):return PRIVATE/'parent-inputs'/name if str(name) in saved else ROOT/name
    prior=runpy.run_path(str(PARENT/'verify_window.py'))['validate'];prior.__globals__['ROOT']=ArchivedRoot()
    assert prior()=={k:v for k,v in parent.items() if k!='attachments'}
    code=[n for n in saved if n.startswith(('cmd/','internal/','tools/'))];assert len(code)==314
    assert {n for n in code if sha(ROOT/n)!=saved[n]}==CHANGED
    assert all((ROOT/n).is_file() and n not in saved for n in ADDED)
    built=read(PRIVATE/'production-inputs.json')
    assert set(built['code_inputs'])==set(code)|ADDED and len(built['code_inputs'])==316
    for name,digest in built['code_inputs'].items():assert sha(ROOT/name)==digest,name
    assert built['image_id']=='sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa'
    for n in code:
        if n not in CHANGED:assert sha(ROOT/n)==saved[n],n
    old=ROOT/'workplace/world-clock-20261002/kol-world-time-linux'
    assert sha(old)=='dea837bc052786c3fa90b2672dd462e81ca768d5c44202f6b2333d12164c12b9'
    product=PRIVATE/'kol-return-scratch-linux';probe=PRIVATE/'kol-settlement-probe2-linux'
    assert sha(product)=='8ce2afbe2e7791bd42a0f113c341f3d0c5c47e01469989a0fe40a62263cebf32'
    assert built['production_sha256']==sha(product)
    assert sha(probe)=='b46b9aaf70ce810363ff4d6dc8639267d18e1b68476b4cd2774b892e822d5f72'
    assert product.stat().st_mode&0o777==0o755
    m=read(PRIVATE/'observation-source.json')
    fixture=ROOT/'workplace/return-tracker-phases-20261002/phase-cases-private.json'
    assert sha(PRIVATE/'phase-cases-private.json')==sha(fixture)==m['fixture_sha256']
    assert len(read(fixture)['cases'])==903
    for name,row in m['files'].items():
        source=ROOT/'internal/battle'/name;observed=PRIVATE/('observed-'+name)
        assert sha(source)==row['production_sha256'] and sha(observed)==row['observed_sha256']
        text=source.read_text()
        if name=='state.go':expected=text.replace('type State struct {','type State struct {\n ScratchWrites []uint16\n')
        else:expected=text.replace('s.ReturnScratch = uint16(value)','s.ReturnScratch = uint16(value)\n s.ScratchWrites=append(s.ScratchWrites,s.ReturnScratch)')
        assert observed.read_text()==expected
    for filename,needle in [('observations.log','903個正常原版階段觀察相符'),('rules.log','internal/battle'),('rules.log','cmd/kol')]:
        log=(PRIVATE/filename).read_text();assert needle in log and 'FAIL' not in log
    assert '正反對照通過' in (PRIVATE/'worklist-selftest.log').read_text()
    assert '0 行沒有' in (PRIVATE/'worklist-scan.log').read_text()
    assert '未完成 22 條' in (PRIVATE/'worklist-verify.log').read_text()
    assert '有 1 行自承沒有條目認領' in (PRIVATE/'worklist-verify.log.attempt1').read_text()
    manifest=read(PRIVATE/'probe2-source.json')
    assert manifest['parent_sha256']==sha(ROOT/'cmd/kol/main.go')
    assert manifest['probe_sha256']==sha(PRIVATE/'probe2-main-private.go')
    normal=[]
    for lang in ('en','zh-Hant'):
        p=PRIVATE/('gui-scratch-'+lang);trace=read(p/'trace.json')
        before=ROOT/'workplace/world-clock-20261002'/('gui-time-'+lang)
        plain=copy.deepcopy(trace)
        for row in plain['steps']:
            for section in ('world','battle'):row.get(section,{}).pop('probe_return_scratch',None)
        assert plain==read(before/'trace.json')
        rows=trace['steps'];assert len(rows)==493
        battle=[r for r in rows if 'battle' in r];assert len(battle)==362
        assert all(r['world']['ticks']==3360 for r in battle)
        world=rows[480]['world'];assert world['probe_return_scratch']==16 and world['ticks']==3320
        assert world['probe_world_seed']==3077853194 and world['probe_world_roll_count']==1491
        assert world['probe_world_last_roll']==dict(N=128,Before=4284445587,After=3077853194,Round=61,Stage=5)
        cmd=read(p/'command.json');assert len(cmd[cmd.index('-flow-script')+1].split())==492
        images=list(p.glob('*.png'));assert len(images)==493
        for image in images:assert sha(image)==sha(before/image.name),image.name
        c=read(p/'comparison.json');assert c['old_states_unchanged'] and c['old_images_unchanged'] and not c['cost_implemented']
        assert c['production_sha256']==sha(product) and c['observation_sha256']==sha(probe)
        reload=PRIVATE/('production-reload-'+lang)
        oldreload=ROOT/'workplace/battle-loot-20261002/fix2'/('production-reload2-'+lang)
        assert read(reload/'trace.json')==read(oldreload/'trace.json')
        images=list(reload.glob('*.png'));assert len(images)==16
        for image in images:assert sha(image)==sha(oldreload/image.name)
        command=read(reload/'command.json');assert command[command.index('-load')+1].startswith('/src/workplace/battle-loot-20261002/fix2/')
        binary=read(reload/'binary.json');assert binary['sha256']==sha(product)
        savedfile=ROOT/Path(command[command.index('-load')+1]).relative_to('/src')
        assert binary['save_sha256']==sha(savedfile)
        normal.append(dict(language=lang,keys=492,states=493,images=493,
            battle_states=362,return_scratch=16,old_results_unchanged=True,reload_states=16,reload_images=16))
    return dict(schema='kol-production-return-scratch-receipt-v1',historical_shared_inputs=1352,
        parent_attachments=len(parent['attachments']),parent_results_recomputed=True,
        changed_code_inputs=sorted(CHANGED),added_code_inputs=sorted(ADDED),production_code_inputs=316,
        production_sha256=sha(product),observation_sha256=sha(probe),native_observations=903,
        internal_and_cmd_tests='passed',normal=normal,world_cost_product=40,world_cost_original=100,
        cost_implemented=False,fresh_lodging_save=False,
        status='正式暫存寫入及正常敗北末值16通過；世界成本與完整返回保存仍待驗。')
def attachments():
    names=set(read(PRIVATE/'parent-preservation.json')['sources'])|ADDED
    names|={str(p.relative_to(ROOT)) for p in PRIVATE.rglob('*') if p.is_file()}
    names|={'docs/re/20261002-production-return-scratch.md','docs/verification/world-return-cost-20261002/verify_scratch.py'}
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
                q=(ROOT/name).stat();assert (q.st_uid,q.st_gid)==(1000,1000),name
            print(json.dumps(dict(schema='kol-production-return-scratch-audit-v1',receipt_sha256=sha(HERE/'receipt.json'),
                attachments=len(receipt['attachments']),full_readonly_verification='passed',host_control=control,
                root_owned_files=bad,md_directories=md,artifact_uid_gid=[1000,1000]),ensure_ascii=False))
    print(json.dumps({k:v for k,v in result.items() if k!='attachments'},ensure_ascii=False))
