"""RE-183封存後補充稽核；不更新已封存的主收據與其附件。"""
from pathlib import Path
import argparse, hashlib, json, os, runpy, stat

ROOT = Path('/src')
HERE = ROOT/'docs/verification/battle-world-20261002'

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def audit():
    verifier = runpy.run_path(str(HERE/'verify_world.py'))
    receipt = json.loads((HERE/'receipt.json').read_text())
    assert verifier['validate']() == {k:v for k,v in receipt.items() if k!='attachments'}
    for name,digest in receipt['attachments'].items():
        p=ROOT/name
        assert sha(p)==digest,name
        assert p.stat().st_uid==p.stat().st_gid==1000,name
    root_owned=[]
    md_dirs=[]
    for base,dirs,files in os.walk(ROOT):
        for name in dirs+files:
            p=Path(base)/name
            if p.lstat().st_uid==0:root_owned.append(str(p.relative_to(ROOT)))
            if p.is_dir() and p.name.endswith('.md'):md_dirs.append(str(p.relative_to(ROOT)))
    assert not root_owned and not md_dirs
    binary=ROOT/'workplace/battle-loot-20261002/fix2/kol-battle-loot-linux'
    assert stat.S_IMODE(binary.stat().st_mode)==0o755
    assert sha(binary)=='7bc7d34a76270fe355c268393ac68308cead4d9c7f59ed960c6681cc311495d4'
    control=json.loads((ROOT/'workplace/battle-world-20261002/final-control.json').read_text())
    assert control['git_diff_check_exit']==0 and not control['staged_files'] and not control['related_containers']
    data=json.loads((ROOT/'docs/worklist.json').read_text())
    done=sum(x.get('status')=='done' for x in data['items'])
    assert done==84 and len(data['items'])-done==22
    return dict(schema='kol-world-entry-audit-v1',date='2026-10-02',
        receipt_sha256=sha(HERE/'receipt.json'),verifier_sha256=sha(HERE/'verify_world.py'),
        auditor_sha256=sha(HERE/'audit_world.py'),attachments=len(receipt['attachments']),
        uid=1000,gid=1000,root_owned=root_owned,md_same_name_directories=md_dirs,
        production_sha256=sha(binary),production_mode='0755',production_unchanged=True,
        worklist=dict(open=22,done=84),git_diff_check_exit=0,staged_files=[],related_containers=[],
        cleanup='一次性容器--rm；GUI Xvfb由trap及wait清理；主機控制面收據確認專案容器0。',
        remaining='世界時刻3360／3320、次格返回與其他战鬥分支仍開放。')

def main():
    parser=argparse.ArgumentParser()
    parser.add_argument('--export',action='store_true')
    args=parser.parse_args()
    result=audit()
    if args.export:
        target=Path('/out/audit.json')
        assert not target.exists()
        target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    else:
        assert json.loads((HERE/'audit.json').read_text())==result
    print(json.dumps(result,ensure_ascii=False))

if __name__=='__main__':
    main()
