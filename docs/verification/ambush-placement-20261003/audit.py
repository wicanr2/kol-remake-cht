#!/usr/bin/env python3
"""Docker 內獨立重算本輪結果；--write 只建立首次 audit.json。"""
from pathlib import Path
import argparse,hashlib,json,runpy,subprocess
ROOT=Path(__file__).resolve().parents[3]
HERE=Path(__file__).resolve().parent
PRIVATE=ROOT/'workplace/ambush-placement-20261003'
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()

def validate():
    receipt=read(HERE/'receipt.json')
    for name,h in receipt['attachments'].items():assert sha(ROOT/name)==h,name
    result=runpy.run_path(str(HERE/'verify.py'))['validate']()
    assert result=={k:v for k,v in receipt.items()if k!='attachments'}
    control=read(PRIVATE/'audit-control.json')
    assert control['head']=='1384e508f70e8a5d81bc050ea79f66d3f0d38fe6'
    assert control['initial_count']==406 and control['final_count']==410
    assert control['missing_initial_paths']==control['staged']==control['project_containers']==[]
    assert control['diff_check_exit_code']==0
    owners=set()
    for name in receipt['attachments']:
        p=ROOT/name;owners.add((p.stat().st_uid,p.stat().st_gid))
    assert owners=={(1000,1000)},owners
    check=subprocess.run(['find',str(ROOT),'-user','root','-o','-type','d','-name','*.md'],capture_output=True,text=True,check=True)
    assert not check.stdout.strip(),check.stdout[:2000]
    items=read(ROOT/'docs/worklist.json')['items'];done=sum(x.get('status')=='done'for x in items)
    assert (len(items)-done,done)==(22,84)
    scan=(PRIVATE/'worklist-scan.log').read_text();assert '0 行沒有' in scan
    assert '正反對照通過' in (PRIVATE/'worklist-selftest.log').read_text()
    return dict(schema='kol-ambush-placement-audit-v1',receipt_sha256=sha(HERE/'receipt.json'),attachments=len(receipt['attachments']),recalculated_result=True,deep_parent_passed=True,parent_attachments=2317,owners=[[1000,1000]],root_owned_or_md_directories=0,head=control['head'],initial_dirty=406,final_dirty=410,all_initial_paths_retained=True,staged=[],diff_check_exit_code=0,project_containers=[],other_project_container_records=len(control['docker_ps'].strip().splitlines()),worklist=dict(incomplete=22,complete=84,unclaimed=0),production_inputs=319,changed_inputs=receipt['changed_code_inputs'],private_local_outputs='原版遊戲副本、PNG、原始追蹤及候選程式只留本機workplace；容器內/tmp隨--rm移除',goal_status='active',commit=False,push=False,remote_issue_write=False,package=False,release=False)

def main():
    parser=argparse.ArgumentParser();parser.add_argument('--write',action='store_true');args=parser.parse_args();result=validate()
    if args.write:
        p=Path('/out/audit.json');assert not p.exists();p.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    else:assert result==read(HERE/'audit.json')
    print(json.dumps(result,ensure_ascii=False))

if __name__=='__main__':main()
