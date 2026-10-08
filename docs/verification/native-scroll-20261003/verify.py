"""正常原版卷軸購買、實際信物槽、繁中顯示與全新保存鏈的有限驗證。"""
from pathlib import Path
import argparse, ast, base64, copy, csv, hashlib, json, os, runpy

ROOT=Path('/src')
PRIVATE=ROOT/'workplace/spell-native-20261003'
HERE=ROOT/'docs/verification/native-scroll-20261003'
TOKEN='a132a500d014b88201ae000002007000'
SPELL='85958e0b1ee00c814000000006000a22'
def read(p): return json.loads(p.read_text())
def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def canonical(v): return json.loads(json.dumps(v))
def semantic(rows):
    rows=copy.deepcopy(rows)
    for row in rows:
        for k in ['exec_text','encounter_text']: row.get('battle',{}).pop(k,None)
    return rows
def steps(p): return read(p/'trace.json')['steps']
def keys(p,flag='-flow-script'):
    cmd=read(p/'command.json'); values=cmd[cmd.index(flag)+1].split()
    assert not any(k.startswith(('CELL:','ACTOR:')) for k in values)
    return values
def party_good(m,oracle):
    assert m['worn'][7]==m['token']
    assert all(m['worn'][7][k]==v for k,v in oracle.items())
    assert base64.b64decode(m['carried'][14]).hex()==TOKEN
def check_pixels():
    load=runpy.run_path(str(ROOT/'tools/png_io.py'))['read_png']
    old=PRIVATE/'formal-neck-zh'; new=PRIVATE/'cjk-neck-zh'
    changed=[i for i in range(184) if (old/f'step-{i:03d}.png').read_bytes()!=(new/f'step-{i:03d}.png').read_bytes()]
    assert changed==[183]
    a=load(old/'step-183.png'); b=load(new/'step-183.png')
    assert (a.width,a.height)==(b.width,b.height)==(960,600)
    aa,bb=a.rgb(),b.rgb()
    points=[(i%b.width,i//b.width) for i in range(b.width*b.height) if aa[3*i:3*i+3]!=bb[3*i:3*i+3]]
    bounds=[min(x for x,y in points),min(y for x,y in points),max(x for x,y in points)+1,max(y for x,y in points)+1]
    assert bounds==[26,222,131,237] and len(points)==431
    glyphs={g['character']:g for g in read(ROOT/'assets/ui-font/manifest.json')['fonts'][0]['glyphs']}
    pixels=bytearray()
    for y in range(18):
        for ch in '藍鍊':
            bits=int(glyphs[ch]['rows_hex'][4*y:4*y+4],16)
            for x in range(16): pixels.extend(b'\xff\xff\xff' if bits&(1<<(15-x)) else b'\0\0\0')
    assert b.crop(24,222,32,18).rgb()==bytes(pixels)
    review=read(PRIVATE/'cjk-pixel-review.json')
    assert review['glyph_pixel_differences']==0 and review['changed_pixels']==431
    assert review['changed_bounds']==bounds and review['name_box']==[24,222,32,18]
    return dict(compared_frames=184,unchanged_frames=183,changed_pixels=431,name_box=[24,222,32,18],glyph_pixel_differences=0)
def save_chain(base,lang,oracle):
    path=base/('fixed-normal-'+lang); rows=steps(path); inputs=keys(path)
    assert len(inputs)==303 and len(rows)==len(list(path.glob('*.png')))==304
    assert inputs[-3:]==['0','7','2']
    purchased=rows[61]['town']['probe_full_party'][0]
    assert purchased['gold']==987 and purchased['order']['name']=='Blue Gem'
    assert purchased['spells']==[dict(name='Kelnalon',effect=6,turn_code=0,power_code=10,guards=[64,0,0,0])]
    party_good(purchased,oracle)
    old=steps(ROOT/'workplace/spell-area-20261003'/('fixed-normal-'+lang))
    # 信物的完整持有物快照已修正；實際戰鬥規則與骰序必須維持。
    for n,(a,b) in enumerate(zip(rows,old)):
        a=copy.deepcopy(a.get('battle',{})); b=copy.deepcopy(b.get('battle',{}))
        a.pop('probe_party',None); b.pop('probe_party',None)
        assert a==b,(lang,n,'battle')
    assert rows[194]['battle']['probe_seed']==1
    before=rows[300]['battle']; after=rows[303]['battle']
    assert before==rows[301]['battle']==rows[302]['battle']
    assert before['prompt']==12 and before['round']==6 and before['probe_seed']==410959018
    assert before['probe_actors'][0]['Negative']
    assert (before['probe_actors'][1]['Chest'],after['probe_actors'][1]['Chest'])==(16,15)
    assert after['probe_seed']==before['probe_seed'] and after['probe_rolls']==before['probe_rolls']
    changed=[(i,k) for i,(a,b) in enumerate(zip(before['probe_actors'],after['probe_actors'])) for k in a if a[k]!=b[k]]
    assert sorted(changed)==[(1,'Chest'),(1,'Damage'),(1,'Hurt')]
    assert (before['probe_actors'][1]['Damage'],after['probe_actors'][1]['Damage'])==(0,1)
    assert (before['probe_actors'][1]['Hurt'],after['probe_actors'][1]['Hurt'])==(0,1)
    text=' '.join(before['exec_text'])
    assert ('選擇要傷害的部位' if lang=='zh-Hant' else 'Select an area to hurt') in text
    lodging=base/('lodging-'+lang); lodgingrows=steps(lodging)
    tokens=read(lodging/'driver-keys.json'); save=read(lodging/'complete.json')
    assert len(tokens)==533 and len(lodgingrows)==534 and len(list(lodging.glob('*.png')))==32
    assert tokens[:303]==inputs and lodgingrows[:304]==rows
    assert save['schema']=='kol-save-v3' and len(save['characters'])==16
    assert save['location']==dict(kind='town',town=7,entry_point=3,x=57,y=8)
    assert save['clock']==dict(ticks=3480,part=0,day=2,month=3,year=0,weather=1)
    byname={m['name']:m for m in save['characters'] if m}
    assert [byname[n]['gold'] for n in ['Olear','Theodoric']]==[931,1009]
    assert byname['Olear']['worn'][7]['record']==TOKEN
    assert byname['Olear']['carried'][14]==TOKEN
    assert byname['Olear']['worn'][7]==byname['Olear']['token']
    assert byname['Olear']['worn'][7]['name']=='Blue Chain'
    assert byname['Olear']['spells']==purchased['spells'] and byname['Olear']['order']==purchased['order']
    observed=base/('observed-reload-'+lang); formal=base/('formal-reload-'+lang)
    reloadrows=steps(observed); formalrows=steps(formal)
    assert len(reloadrows)==len(formalrows)==16
    assert len(list(formal.glob('*.png')))==len(list(observed.glob('*.png')))==16
    plain=copy.deepcopy(reloadrows)
    for x in plain: x.get('town',{}).pop('probe_full_party',None)
    assert plain==formalrows
    for p in observed.glob('*.png'): assert p.read_bytes()==(formal/p.name).read_bytes()
    item=runpy.run_path(str(ROOT/'docs/verification/unarmed-resolution-20261003/verify.py'))['item']
    final=reloadrows[-1]['town']['probe_full_party']
    assert [m['name'] for m in final]==['Theodoric','Olear']
    assert sum(len(m['carried']) for m in final)==34
    for m in final:
        saved=byname[m['name']]
        assert [base64.b64decode(x).hex() for x in m['carried']]==saved['carried']
        for k in ['id','wounds','gold','place','town_code','building_id','sheet','quests','progress']:
            assert m.get(k,0)==saved.get(k,0),(lang,m['name'],k)
        assert (m.get('spells') or [])==(saved.get('spells') or [])
        assert m.get('order')==saved.get('order')
        for k in ['in_hand','backpack','worn','token']: assert item(m.get(k))==item(saved.get(k))
    return dict(language=lang,normal_keys=303,normal_states=304,lodging_keys=533,lodging_states=534,
                lodging_images=32,reload_states=16,reload_images=16,carried_slots=34,neck_record=TOKEN),save,rows
def validate(deep=True):
    p=read(PRIVATE/'parent-preservation.json'); sources=p['sources']
    assert len(sources)==1390 and p['parent_attachments']==5123
    for n,d in sources.items(): assert sha(PRIVATE/'parent-inputs'/n)==d,n
    parent_path=PRIVATE/'parent-inputs/docs/verification/spell-area-20261003/receipt.json'
    parent=read(parent_path); assert sha(parent_path)==p['parent_receipt_sha256']
    for n,d in parent['attachments'].items():
        assert sha(PRIVATE/'parent-inputs'/n if n in sources else ROOT/n)==d,n
    if deep:
        class ArchivedRoot:
            def __truediv__(self,n): return PRIVATE/'parent-inputs'/n if str(n) in sources else ROOT/n
        previous=runpy.run_path(str(PRIVATE/'parent-inputs/docs/verification/spell-area-20261003/verify.py'))['validate']
        previous.__globals__['ROOT']=ArchivedRoot()
        assert canonical(previous())=={k:v for k,v in parent.items() if k!='attachments'}
        assert '最新續核RE-200' in (ROOT/'CONTEXT.md').read_text()
        assert '20261003-native-scroll-token.md' in (ROOT/'docs/re/RESEARCH-LOG.md').read_text()
        for filename,signal in [('worklist-selftest-final.log','正反對照通過'),('worklist-verify-final.log','未完成 22 條'),('worklist-scan-final.log','0 行沒有')]:
            assert signal in (PRIVATE/filename).read_text(),filename
    bindings=read(PRIVATE/'source-bindings.json')
    assert len(bindings['code_inputs'])==319
    assert bindings['changed_code_inputs']==['internal/town/scroll.go','internal/town/scroll_test.go']
    assert bindings['added_code_inputs']==[]
    assert sorted(n for n in bindings['code_inputs'] if bindings['code_inputs'][n]!=sources[n])==bindings['changed_code_inputs']
    for n,d in {**bindings['code_inputs'],**bindings['inputs']}.items(): assert sha(ROOT/n)==d,n
    review=read(PRIVATE/'token-evidence-review.json'); assert review['status']=='READY'
    section=next(ast.literal_eval(n.value) for n in ast.walk(ast.parse((PRIVATE/'begin_evidence.py').read_text()))
                 if isinstance(n,ast.Assign) and any(isinstance(t,ast.Name) and t.id=='section' for t in n.targets))
    draft=(PRIVATE/'parent-inputs/docs/spec/012-combat-spells.md').read_text()+section
    assert hashlib.sha256(draft.encode()).hexdigest()==review['draft_spec_sha256']
    assert '入會後實際頸部槽缺信物' in (PRIVATE/'old-token-negative.log').read_text()
    for name in ['TestScrollJoinsTheOrder','TestScrollJoinKeepsOccupiedNeck','TestScrollJoinRefusesFullInventory','TestScrollJoinInvalidTokenDoesNotCharge']:
        assert '--- PASS: '+name in (PRIVATE/'token-targeted.log').read_text()
    assert 'github.com/wicanr2/kol-remake-cht/cmd/kol' in (PRIVATE/'rules.log').read_text()
    native=read(PRIVATE/'shop1/native-shop-private.json')
    assert sha(PRIVATE/'shop1/native-shop-private.json')==review['native_sha256']
    assert sha(PRIVATE/'shop1/native-driver.go')==review['native_driver_sha256']
    assert native['tool']=='dosgolem' and native['revision']=='968ca774705853276ae55c8d1163c369481ac71f'
    assert native['input_sha256']=='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
    assert native['normal_startup'] and not any(native[k] for k in ['direct_entry','state_injection','wait_stub','stopped_for_battle','stopped_for_healing_question'])
    assert native['reached_purchase'] and native['final_spell_slot_hex']==SPELL and native['seed']==1
    assert len(native['rows'])==164 and len(native['rolls'])==10
    assert native['rows'][:145]==read(PRIVATE/'entry7-1/native-shop-private.json')['rows']
    assert native['rows'][144]['state']['camera']==[62,27]
    assert native['rows'][144]['state']['bounds']==[16,48,63,31]
    assert native['missing']==['region1.dat','region2.dat','region3.dat','region4.dat']
    for run in ['exit1','entry7-1','shop1']:
        d=read(PRIVATE/run/'driver.json')
        for file,key in [('native-driver.go','native_driver_sha256'),('instrumented-book-private.go','startup_sha256'),('keys.txt','keys_sha256')]:
            assert sha(PRIVATE/run/file)==d[key]
        driver=(PRIVATE/run/'native-driver.go').read_text()
        assert driver.count('.WriteU16(')==2
        assert 'o.WriteU16(o.DS(0x1600),1);o.WriteU16(o.DS(0x1602),0)' in driver
        assert d['normal_startup'] and not d['state_injection'] and not d['wait_stub']
    old=steps(PRIVATE/'remake-shop1'); oracle=read(PRIVATE/'token-item-oracle.json')
    parser=runpy.run_path(str(ROOT/'tools/extract_town_buildings.py'))['good']
    assert parser(bytes.fromhex(TOKEN),0)==oracle
    assert sha(PRIVATE/'token-item-oracle.json')==review['token_item_oracle_sha256']
    offsets=[0x30]+list(range(0xa0,0x1a0,16))
    raw=bytes.fromhex(native['rows'][-1]['state']['characters'][0])
    assert raw[0x170:0x180].hex()==TOKEN and raw[0x200:0x210].hex()==SPELL
    assert raw[0x1a0:0x1b0].hex()=='8132a501cad00000b4b1959a5ee4ce80'
    differences=[]
    for n,m in enumerate(old[-1]['town']['probe_full_party']):
        raw=bytes.fromhex(native['rows'][-1]['state']['characters'][n])
        differences += [(n,i,hex(off)) for i,off in enumerate(offsets) if raw[off:off+16]!=base64.b64decode(m['carried'][i])]
    assert differences==[(0,14,'0x170')]
    shops=[]
    for name in ['fixed-shop-en','fixed-shop-zh','cjk-shop-zh']:
        path=PRIVATE/name; rows=steps(path); shops.append(rows)
        assert len(keys(path))==172 and len(rows)==len(list(path.glob('*.png')))==173
        for i in range(123):
            a=native['rows'][21+i]['state']; b=rows[30+i]['world']
            assert a['camera']==[b['x'],b['y']] and a['marker']==[b['sub_x'],b['sub_y']]
            assert (a['time'],a['part'],a['seed'])==(b['ticks'],b['time_index'],b['probe_world_seed'])
        for i in range(143):
            a=native['rows'][21+i]['state']; b=rows[30+i]
            members=b.get('town',{}).get('probe_full_party') or b.get('world',{}).get('probe_full_party')
            if members is None:
                # 世界追蹤只另存位置／時間；持有物在首個及後續城鎮追蹤驗證。
                assert i<123
                continue
            if i>=123: assert len(members)==2 and sum(len(m['carried']) for m in members)==34
            for n,m in enumerate(members):
                raw=bytes.fromhex(a['characters'][n]); assert int.from_bytes(raw[0x2a:0x2d],'big')==m['gold']
                assert all(raw[off:off+16]==base64.b64decode(m['carried'][j]) for j,off in enumerate(offsets))
        party_good(rows[-1]['town']['probe_full_party'][0],oracle)
    assert semantic(shops[0])==semantic(shops[1])==semantic(shops[2])
    for lang in ['en','zh']:
        formal=PRIVATE/('formal-neck-'+lang); rows=steps(formal)
        assert len(keys(formal))==183 and len(rows)==len(list(formal.glob('*.png')))==184
        plain=copy.deepcopy(shops[0 if lang=='en' else 1])
        for x in plain:
            x.get('town',{}).pop('probe_full_party',None)
            for k in ['probe_group','probe_weather','probe_world_seed']: x.get('world',{}).pop(k,None)
        assert rows[:173]==plain
        for i in range(173): assert (formal/f'step-{i:03d}.png').read_bytes()==(PRIVATE/('fixed-shop-'+lang)/f'step-{i:03d}.png').read_bytes()
    assert steps(PRIVATE/'cjk-neck-zh')==steps(PRIVATE/'formal-neck-zh')
    assert len(keys(PRIVATE/'cjk-neck-zh'))==183
    cjk=read(PRIVATE/'cjk-evidence-review.json'); assert cjk['status']=='READY'
    assert sha(PRIVATE/'cjk-spec-draft.md')==cjk['draft_sha256']
    assert cjk['missing']==['Blue Chain','Pearl Amulet','Red Neclace','Stone Amulet']
    assert '缺譯原名' in (PRIVATE/'cjk-old-table-negative.log').read_text()
    counts=json.loads((PRIVATE/'cjk-name-check.log').read_text())
    assert counts==dict(required_unique_names=519,catalog_keys=537,catalog_unique_names=519,problems=0)
    catalog=list(csv.DictReader((ROOT/'docs/i18n/display-names-zh-Hant.tsv').open(),delimiter='\t'))
    tokens=[x for x in catalog if x['key'].startswith('order-token.')]
    assert len(tokens)==6 and all(x['zh-Hant']==cjk['translations'][x['en']] and x['status']=='暫譯' for x in tokens)
    font=read(ROOT/'assets/ui-font/manifest.json')
    assert font['fonts']==read(PRIVATE/'parent-inputs/assets/ui-font/manifest.json')['fonts']
    assert sha(ROOT/'assets/ui-font/zh-hant-16.png')==font['fonts'][0]['png_sha256']
    for e in font['inputs']: assert sha(ROOT/e['path'])==e['sha256']
    assert read(PRIVATE/'cjk-font-review.json')['old_glyphs_unchanged']
    pixel_result=check_pixels()
    fit=read(PRIVATE/'cjk-text-fit.json')
    assert fit['checked']==9675 and fit['failed']==0 and fit['unregistered_kinds']==[]
    assert fit==read(ROOT/'docs/verification/cjk-text-fit.json')
    for n,d in fit['inputs_sha256'].items(): assert sha(ROOT/n)==d,n
    normal=[]; saves=[]; traces=[]
    for lang in ['en','zh-Hant']:
        result,save,rows=save_chain(PRIVATE,lang,oracle); normal.append(result); saves.append(save); traces.append(rows)
    assert semantic(traces[0])==semantic(traces[1])
    assert {k:v for k,v in saves[0].items() if k!='saved_at'}=={k:v for k,v in saves[1].items() if k!='saved_at'}
    cjk_result,cjk_save,cjk_rows=save_chain(PRIVATE/'cjk-save','zh-Hant',oracle)
    assert cjk_rows==traces[1]
    assert {k:v for k,v in cjk_save.items() if k!='saved_at'}=={k:v for k,v in saves[1].items() if k!='saved_at'}
    return dict(schema='kol-native-scroll-token-receipt-v1',historical_shared_inputs=1390,parent_attachments=5123,
                parent_results_recomputed=deep,code_inputs=319,changed_code_inputs=bindings['changed_code_inputs'],added_code_inputs=[],
                native_states=164,normal_world_states_compared=123,normal_shop_states_compared=20,
                carried_slots_per_member=17,carried_slots_per_party=34,normal=normal,cjk_save=cjk_result,
                cjk_names=counts,cjk_pixels=pixel_result,cjk_text_fit=9675,original_normal_shop=True,original_normal_area=False,full_entry_rng_verified=False,
                status='正常空頸槽入會、卷軸購買、繁中信物及新保存鏈有限驗證；正常原版AREA與完整重製仍開放。')
def attachments():
    names={str(p.relative_to(ROOT)) for p in PRIVATE.rglob('*') if p.is_file()}
    names|={'CONTEXT.md','WORKLIST.md','WORKLOG.md','docs/worklist.json','docs/re/RESEARCH-LOG.md',
            'docs/re/20261003-native-scroll-token.md','docs/spec/012-combat-spells.md','docs/spec/033-cjk-display-names.md',
            'docs/verification/native-scroll-20261003/verify.py','docs/verification/cjk-text-fit.json','docs/i18n/README.md','docs/manual/glossary.md'}
    bindings=read(PRIVATE/'source-bindings.json'); names|=set(bindings['inputs'])|set(bindings['code_inputs'])
    return {n:sha(ROOT/n) for n in sorted(names)}
if __name__=='__main__':
    ap=argparse.ArgumentParser(); ap.add_argument('--seal',action='store_true'); ap.add_argument('--data',action='store_true'); ap.add_argument('--audit',action='store_true'); args=ap.parse_args()
    assert sum([args.seal,args.data,args.audit])<=1
    result=validate(not args.data)
    if args.seal:
        target=Path('/out/receipt.json'); assert not target.exists(); result['attachments']=attachments()
        target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    elif not args.data:
        saved=read(HERE/'receipt.json')
        for n,d in saved['attachments'].items(): assert sha(ROOT/n)==d,n
        assert canonical(result)=={k:v for k,v in saved.items() if k!='attachments'}
        if args.audit:
            control=read(PRIVATE/'host-control.json')
            assert control['git_diff_check']=='passed' and control['git_staged']==control['project_containers']==[]
            assert control['head']=='1384e508f70e8a5d81bc050ea79f66d3f0d38fe6'
            bad=[]; md=[]
            for directory,dirs,files in os.walk(ROOT):
                for name in dirs+files:
                    p=Path(directory)/name
                    if p.lstat().st_uid==0: bad.append(str(p))
                    if p.is_dir() and p.name.endswith('.md'): md.append(str(p))
            assert not bad and not md
            for n in saved['attachments']:
                st=(ROOT/n).stat(); assert (st.st_uid,st.st_gid)==(1000,1000),n
            target=Path('/out/audit.json'); assert not target.exists()
            target.write_text(json.dumps(dict(status='PASS',receipt_sha256=sha(HERE/'receipt.json'),attachments_verified=len(saved['attachments']),
                                             parent_results_recomputed=True,root_owned_files=0,markdown_directories=0,control=control),ensure_ascii=False,indent=2)+'\n')
    print(json.dumps({k:v for k,v in result.items() if k!='attachments'},ensure_ascii=False))
