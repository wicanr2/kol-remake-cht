from pathlib import Path
import copy,hashlib,json,re
root=Path('workplace/inventory-empty-slots-20261001')
def sha(p):return hashlib.sha256(Path(p).read_bytes()).hexdigest()
reports=[]
for name in ('full-chain','full-chain-reload'):
    op=Path('docs/verification/player-path')/name/'trace.json'
    np=root/'battle-regression'/name/'trace.json'
    old=json.loads(op.read_text())['steps'];new=json.loads(np.read_text())['steps']
    assert len(old)==len(new)
    changed=[i for i,(a,b) in enumerate(zip(old,new)) if a!=b]
    party=[None,None]
    for i,pair in enumerate(zip(old,new)):
        normalized=[]
        for n,r in enumerate(pair):
            r=copy.deepcopy(r)
            if 'town' in r:
                if 'party' in r['town']:party[n]=r['town']['party']
                r['town'].pop('party',None)
            normalized.append(r)
        assert normalized[0]==normalized[1],(name,i,'持物摘要以外狀態不同')
        if party[0] is not None and party[1] is not None:
            assert [re.sub(r'carried=\d+','carried=<查證中的計數>',s) for s in party[0]]==[
                re.sub(r'carried=\d+','carried=<查證中的計數>',s) for s in party[1]],(name,i,'隊伍其他欄位不同')
    reports.append({'name':name,'states':len(old),'strict_reference_passed':False,
                    'changed_steps':changed,'other_states_identical':True,
                    'party_other_values_identical':True,'old_sha256':sha(op),'new_sha256':sha(np),
                    'scope':'此分欄診斷只解釋差異，不替換嚴格比較、不作驗收忽略區'})
op=Path('workplace/saves/full-chain.json');np=root/'battle-regression/full-chain/save.json'
old=json.loads(op.read_text());new=json.loads(np.read_text())
corrected=copy.deepcopy(old)
fields=[]
for i,(a,b) in enumerate(zip(corrected['characters'],new['characters'])):
    if a is None:
        assert b is None;continue
    assert a['name']==b['name']
    differences=[k for k in a if a[k]!=b[k]]
    if differences:fields.append({'slot':i,'name':a['name'],'fields':differences})
    a['backpack']=[g for g in (a['backpack'] or []) if bytes.fromhex(g['record'])[0]!=0] or None
    if a['in_hand'] and bytes.fromhex(a['in_hand']['record'])[0]==0:a['in_hand']=None
    if a['carried']!=b['carried']:
        worn=a.get('worn',a.get('equipment'))
        assert worn is not None and len(worn)==10
        pack=a['backpack'] or []
        items=[a['in_hand']]+pack+[None]*(6-len(pack))+worn
        a['carried']=[g['record'] if g else '00'*16 for g in items]
    assert a==b,(i,'原空物品與同步以外的角色資料差異')
corrected.pop('saved_at');new.pop('saved_at')
assert corrected==new
result={'schema':'kol-inventory-battle-impact-v1','trace_checks':reports,
        'save_changes':fields,'save_old_sha256':sha(op),'save_new_sha256':sha(np),
        'save_other_fields_identical':True,
        'save_allowed_changes':'移除首 byte 為零的背包／手持物品，及既有戰後 SyncCarried 的同步結果；保存時間戳記另列',
        'baseline_action':'舊收據不覆寫；新初始資料的全量回歸仍待審查與重生'}
(root/'battle-impact.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(result,ensure_ascii=False,indent=2))
