from pathlib import Path
import json,hashlib,os
out=Path('/out')/os.environ.get('KOL_CURSOR_ATTEMPT','gui1');p=out/'inputs.json';d=json.loads(p.read_text())
d['sources_unchanged']=all(hashlib.sha256(Path(r['path']).read_bytes()).hexdigest()==r['sha256'] for r in d['sources'])
d['outputs']=[{'path':str(f.relative_to(out)),'sha256':hashlib.sha256(f.read_bytes()).hexdigest(),'uid':f.stat().st_uid,'gid':f.stat().st_gid} for f in sorted(out.rglob('*')) if f.is_file() and f!=p]
assert d['sources_unchanged']
assert all(r['uid']==os.getuid() and r['gid']==os.getgid() for r in d['outputs'])
p.write_text(json.dumps(d,ensure_ascii=False,indent=2)+'\n')
print('原版正常路徑完成；來源與輸出擁有權已核對。')
