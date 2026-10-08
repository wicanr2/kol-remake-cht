"""正常勝利、領取戰利品及新住宿保存的有限唯讀驗收。"""
from pathlib import Path
import argparse,base64,copy,hashlib,json,os,runpy
ROOT=Path('/src')
PRIVATE=ROOT/'workplace/normal-victory-20261003'
HERE=ROOT/'docs/verification/normal-victory-20261003'
def read(p):return json.loads(p.read_text())
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def canonical_item(value):
    if value is None:return None
    if isinstance(value,list):return [canonical_item(v)for v in value]
    value=copy.deepcopy(value)
    # town.Good只省略這兩個合法零值；非零值與其他欄位均逐欄比較。
    value.setdefault('picture',0);value.setdefault('movement',0)
    return value
def validate():
    saved=read(PRIVATE/'parent-preservation.json')['sources'];assert len(saved)==1367
    for name,digest in saved.items():assert sha(PRIVATE/'parent-inputs'/name)==digest,name
    parenthere=ROOT/'docs/verification/world-cost-resume-20261003'
    parent=read(PRIVATE/'parent-inputs/docs/verification/world-cost-resume-20261003/receipt2.json')
    for name,digest in parent['attachments'].items():assert sha(PRIVATE/'parent-inputs'/name if name in saved else ROOT/name)==digest,name
    class ArchivedRoot:
        def __truediv__(self,name):return PRIVATE/'parent-inputs'/name if str(name) in saved else ROOT/name
    prior=runpy.run_path(str(parenthere/'verify_final.py'))['validate'];prior.__globals__['ROOT']=ArchivedRoot()
    assert prior()=={k:v for k,v in parent.items()if k!='attachments'}
    priorbuilt=read(ROOT/'workplace/world-cost-resume-20261003/production-inputs.json')
    built=read(PRIVATE/'save-production-inputs.json')
    assert built['changed_inputs']==['cmd/kol/complete_save_test.go','internal/save/complete.go','internal/save/complete_test.go']
    assert len(built['code_inputs'])==len(priorbuilt['code_inputs'])==318
    for name,digest in built['code_inputs'].items():assert sha(ROOT/name)==digest,name
    assert [n for n in built['code_inputs']if built['code_inputs'][n]!=priorbuilt['code_inputs'][n]]==built['changed_inputs']
    assert sha(ROOT/'workplace/world-cost-resume-20261003/kol-world-cost-linux')==priorbuilt['production_sha256']
    assert sha(PRIVATE/'kol-victory-save-linux')==built['production_sha256']
    assert sha(PRIVATE/'kol-victory-inn10-linux')==built['observer_sha256']
    clocks=read(ROOT/'docs/verification/world-clock-probe-20260927/world-clock-probe.json')['rows']
    delayed=next(r for r in clocks if r['key']=='Left'and r['step']==13)
    assert (delayed['time_before'],delayed['time_after'],delayed['word_2DF42'])==(3420,3360,0)
    assert '合法延後時段不可被拒絕' in(PRIVATE/'old-save-negative.log').read_text()
    assert '正常住宿提交失敗：世界時段與剩餘時刻不一致' in(PRIVATE/'fresh-inn9-en/run.log').read_text()
    assert 'FAIL' not in(PRIVATE/'save-fix-rules.log').read_text()
    assert 'TestCompleteSaveKeepsPendingWorldPart' in(PRIVATE/'save-fix-targeted.log').read_text()
    native=read(ROOT/'workplace/battle-victory-20261002/source4/complete-source-private.json')
    assert native['normal_startup'] and not any(native[k]for k in ['direct_entry','state_injection','wait_stub'])
    assert native['input_sha256']=='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
    assert native['seed']==1 and native['rounds']==4 and len(native['inputs'])==147
    actor_diffs=runpy.run_path(str(ROOT/'docs/verification/battle-settlement-20261002/verify.py'))['actor_diffs']
    oldkeys=read(ROOT/'workplace/world-cost-resume-20261003/gui-cost-en/command.json')
    prefix=oldkeys[oldkeys.index('-flow-script')+1].split()[:119]
    offsets=[0x30]+list(range(0xa0,0x1a0,16))
    outputs=[];traces=[];saves=[]
    for lang in ['en','zh-Hant']:
        path=PRIVATE/('normal3-'+lang);rows=read(path/'trace.json')['steps'];traces.append(rows)
        assert len(rows)==270 and len(list(path.glob('*.png')))==270
        command=read(path/'command.json');keys=command[command.index('-flow-script')+1].split();assert len(keys)==269 and keys[:119]==prefix
        assert not any(k.startswith(('CELL:','ACTOR:'))for k in keys)
        previous=read(ROOT/('workplace/world-cost-resume-20261003/gui-cost-'+lang+'/trace.json'))['steps']
        assert rows[:120]==previous[:120]
        for r in rows[:120]:assert (path/(r['name']+'.png')).read_bytes()==(ROOT/('workplace/world-cost-resume-20261003/gui-cost-'+lang)/(r['name']+'.png')).read_bytes()
        entries=[]
        for boundary in native['boundaries']:
            if not boundary['at'].startswith('執行段入口'):continue
            matching=[r for r in rows if r.get('battle',{}).get('prompt')==8 and r['battle']['round']==boundary['round']]
            assert len(matching)==1;row=matching[0]
            assert row['battle']['probe_seed']==boundary['state']['seed']
            assert actor_diffs(boundary['state'],row['battle']['probe_actors'])==[]
            entries.append(row['step'])
        end=next(r for r in rows if r.get('battle',{}).get('stage')==5)
        b=end['battle'];assert(b['round'],b['verdict'],b['probe_return_scratch'])==(4,2,1)
        diffs=actor_diffs(native['battle_exit_state'],b['probe_actors'])
        assert diffs==[(6,'LastHurt',25,23)],diffs
        assert len(b['probe_rolls'])==99
        assert [r['Before']for r in b['probe_rolls']]==[r['seed_before']for r in native['rolls'][:-1]]
        roll_diffs=[(i,source['n'],observed['N'])for i,(source,observed)in enumerate(zip(native['rolls'][:-1],b['probe_rolls']))if source['n']!=observed['N']]
        assert roll_diffs==[(73,100,3),(74,2,100)],roll_diffs
        world=rows[-1]['world'];assert[world[k]for k in ['x','y','sub_x','sub_y','ticks']]==[49,23,4,4,3300]
        assert world['probe_return_scratch']==1 and world['probe_world_seed']==1084083701 and world['probe_world_roll_count']==100
        assert world['probe_world_last_roll']['Before']==1502690714
        assert all(r['world']['ticks']==3360 for r in rows if r.get('battle'))
        party=rows[-1]['town']['probe_full_party'];rawchars=native['post_battle'][4]['state']['characters']
        for member,rawhex in zip(party,rawchars):
            raw=bytes.fromhex(rawhex)
            assert [base64.b64decode(x).hex()for x in member['carried']]==[raw[o:o+16].hex()for o in offsets]
            assert member['wounds']==raw[27] and member['gold']==int.from_bytes(raw[42:45],'big')
            assert member['sheet']['adventure_points']==int.from_bytes(raw[38:40],'big')
        assert party[0]['backpack'][-1]['name']=='War Hammer'
        assert [m['gold']for m in party]==[1835,1067] and [m['sheet']['adventure_points']for m in party]==[2891,749]
        cancellation=read(PRIVATE/('normal2-'+lang)/'trace.json')['steps'][-1]['town']['probe_full_party']
        assert cancellation[0]['carried']!=party[0]['carried']
        lodging=PRIVATE/('fresh-inn10-'+lang);lodgingrows=read(lodging/'trace.json')['steps'];save=read(lodging/'complete.json');saves.append(save)
        assert len(lodgingrows)==len(read(lodging/'driver-keys.json'))+1
        assert lodgingrows[:270]==rows
        assert save['schema']=='kol-save-v3' and len(save['characters'])==16
        assert save['settings']['faithful_inn_stay'] and save['location']==dict(kind='town',town=7,entry_point=3,x=57,y=8)
        assert save['clock']['ticks']==3360 and save['clock']['part']==0
        byname={m['name']:m for m in save['characters']if m}
        beforestay=lodgingrows[494]['town']['probe_full_party']
        assert [m['name']for m in beforestay]==['Olear','Theodoric']
        route_diffs=[(m['name'],i)for m,s in zip(party,beforestay)for i,(a,b)in enumerate(zip(m['carried'],s['carried']))if a!=b]
        assert route_diffs==[('Theodoric',0)],route_diffs
        assert base64.b64decode(beforestay[1]['carried'][0])==bytes(16)
        assert [m['gold']for m in beforestay]==[1835,1067]
        assert [m['wounds']for m in beforestay]==[0,0]
        assert [m['sheet']['adventure_points']for m in beforestay]==[2891,749]
        for m in beforestay:
            s=byname[m['name']]
            assert s['carried']==[base64.b64decode(x).hex()for x in m['carried']]
            assert s['wounds']==m['wounds'] and s['gold']==m['gold']-(56 if m['name']=='Olear'else 0)
            assert s['sheet']['adventure_points']==m['sheet']['adventure_points']
        observed=PRIVATE/('observed-reload-'+lang);formal=PRIVATE/('formal-reload-'+lang)
        reloadrows=read(observed/'trace.json')['steps'];formalrows=read(formal/'trace.json')['steps']
        assert len(reloadrows)==len(formalrows)==16
        plain=copy.deepcopy(reloadrows)
        for row in plain:row.get('town',{}).pop('probe_full_party',None)
        assert plain==formalrows
        for image in observed.glob('*.png'):assert image.read_bytes()==(formal/image.name).read_bytes()
        assert len(list(formal.glob('*.png')))==16
        final=reloadrows[-1]['town']['probe_full_party'];assert[m['name']for m in final]==['Theodoric','Olear']
        for m in final:
            s=byname[m['name']]
            assert[base64.b64decode(x).hex()for x in m['carried']]==s['carried']
            for key in ['id','wounds','gold','place','town_code','building_id','sheet','quests','progress']:assert m.get(key,0)==s.get(key,0),(lang,m['name'],key)
            for key in ['in_hand','backpack','worn']:
                actual=m.get(key);expected=s.get(key)
                assert canonical_item(actual)==canonical_item(expected),(lang,m['name'],key)
        outputs.append(dict(language=lang,normal_keys=269,states=270,images=270,execution_entries=entries,
            rolls=100,roll_seed_before_matches=100,roll_face_differences=roll_diffs,world=[49,23,4,4,3300],carried_slots=34,gold=[1835,1067],adventure_points=[2891,749],
            final_temporary_difference=diffs,return_route_carried_changes=route_diffs,lodging_keys=len(lodgingrows)-1,lodging_images=len(list(lodging.glob('*.png'))),formal_reload_states=16,formal_reload_images=16))
    for a,b in zip(*traces):
        for k in ['probe_actors','probe_party','probe_rolls','probe_seed','probe_return_scratch']:assert a.get('battle',{}).get(k)==b.get('battle',{}).get(k)
        assert a.get('town',{}).get('probe_full_party')==b.get('town',{}).get('probe_full_party')
        assert a.get('world')==b.get('world')
    assert {k:v for k,v in saves[0].items()if k!='saved_at'}=={k:v for k,v in saves[1].items()if k!='saved_at'}
    for filename,signal in [('worklist-selftest.log','正反對照通過'),('worklist-scan.log','0 行沒有'),('worklist-verify.log','未完成 22 條')]:assert signal in(PRIVATE/filename).read_text()
    return dict(schema='kol-normal-victory-receipt-v1',historical_shared_inputs=1367,parent_attachments=len(parent['attachments']),parent_results_recomputed=True,
        production_sha256=built['production_sha256'],production_unchanged=False,changed_code_inputs=built['changed_inputs'],normal=outputs,
        cancelled_loot_negative_rejected=True,full_entry_rng_verified=False,full_battle_rng_parity=False,full_actor_record_parity=False,
        status='本場正常勝利返回、報酬、34槽與新住宿重載有限CONFORMED；敵方倒地後LastHurt差異及其他玩家分支仍開放。')
def attachments():
    names={str(p.relative_to(ROOT))for p in PRIVATE.rglob('*')if p.is_file()}
    names.update(['CONTEXT.md','WORKLIST.md','WORKLOG.md','docs/worklist.json','docs/re/RESEARCH-LOG.md','docs/re/20261003-normal-victory.md','docs/spec/011-wilderness-ambush.md','docs/spec/023-complete-save.md','docs/spec/027-battle-exit-wounds.md','docs/verification/world-clock-probe-20260927/world-clock-probe.json','docs/verification/normal-victory-20261003/verify.py'])
    return {n:sha(ROOT/n)for n in sorted(names)}
if __name__=='__main__':
    ap=argparse.ArgumentParser();ap.add_argument('--seal',action='store_true');ap.add_argument('--audit',action='store_true');args=ap.parse_args()
    assert not(args.seal and args.audit);result=validate()
    if args.seal:
        target=Path('/out/receipt.json');assert not target.exists();result['attachments']=attachments();target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    else:
        saved=read(HERE/'receipt.json')
        for name,digest in saved['attachments'].items():assert sha(ROOT/name)==digest,name
        assert result=={k:v for k,v in saved.items()if k!='attachments'}
        if args.audit:
            control=read(PRIVATE/'host-control.json')
            assert control['git_diff_check']=='passed'and control['git_staged']==control['project_containers']==[]
            assert control['head']=='1384e508f70e8a5d81bc050ea79f66d3f0d38fe6'
            bad=[];md=[]
            for directory,dirs,files in os.walk(ROOT):
                for name in dirs+files:
                    p=Path(directory)/name
                    if p.lstat().st_uid==0:bad.append(str(p))
                    if p.is_dir()and p.name.endswith('.md'):md.append(str(p))
            assert not bad and not md
            names=set(saved['attachments'])|set(read(PRIVATE/'save-production-inputs.json')['code_inputs'])
            for name in names:
                s=(ROOT/name).stat();assert(s.st_uid,s.st_gid)==(1000,1000),name
            assert '20261003-normal-victory.md'in(ROOT/'docs/re/RESEARCH-LOG.md').read_text()
            print(json.dumps(dict(schema='kol-normal-victory-audit-v1',receipt_sha256=sha(HERE/'receipt.json'),attachments=len(saved['attachments']),
                code_inputs=318,historical_attachments_preserved=3961,full_readonly_verification='passed',host_control=control,
                root_owned_files=bad,md_directories=md,artifact_uid_gid=[1000,1000]),ensure_ascii=False))
    print(json.dumps({k:v for k,v in result.items()if k!='attachments'},ensure_ascii=False))
