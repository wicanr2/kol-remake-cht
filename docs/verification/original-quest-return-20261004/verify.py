from pathlib import Path
import hashlib,json,copy
root=Path(__file__).resolve().parents[3]
receipt=json.loads(Path(__file__).with_name('receipt.json').read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
for group in ['attachments','original_inputs','production_inputs']:
 for path,h in receipt[group].items():assert sha(root/path)==h,path
work=root/'workplace/original-quest-return-20261004'
d=json.loads((work/'native4/quest-private.json').read_text())
def validate(data):
 assert data['seed']==1 and data['state_injection'] is False and data['direct_entry'] is False
 e=data['events']
 q=[x for x in e if x['ida']=='0x2B0FF'];assert len(q)==1
 assert q[0]['state']['camera']==[48,21]
 assert q[0]['state']['quest_scan_index']==5
 ret=[x for x in e if x['ida']=='0x2B3C6'];assert len(ret)==1
 assert ret[0]['state']['camera']==[48,21] and ret[0]['state']['marker']==[4,4]
 assert data['battle_exits']>=1
 pending=None;pairs=0
 for x in e:
  if x['ida']=='0x29800':assert pending is None;pending=x['state']
  elif x['ida']=='0x29981':
   assert pending is not None
   assert (pending['time']-x['state']['time'])&65535==2*(pending['return_cost_byte']+pending['movement'])
   pairs+=1;pending=None
 assert pending is not None and pending['time']==0 and pending['return_index']==65535
 assert pairs==receipt['checks']['cost_pairs']
validate(d)
bad=copy.deepcopy(d)
next(x for x in bad['events'] if x['ida']=='0x2B3C6')['state']['marker'][0]+=1
try:validate(bad)
except AssertionError:pass
else:raise AssertionError('返回座標反例未拒絕')
assert 'PASS' in (work/'production-cost.log').read_text()
assert int((work/'production-cost-negative.exit').read_text())!=0
assert '--- FAIL: TestNormalBookEntry' in (work/'native2/run.log').read_text()
assert receipt['same_state_full_journey'] is False
print('PASS immutable original normal quest return, production cost samples and negative controls')
