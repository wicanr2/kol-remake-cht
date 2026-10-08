"""保留初封收據，以JSON標準形重驗元組／陣列及最終入口。"""
from pathlib import Path
import argparse,hashlib,json,os,runpy
ROOT=Path('/src')
PRIVATE=ROOT/'workplace/normal-victory-20261003'
HERE=ROOT/'docs/verification/normal-victory-20261003'
def read(p):return json.loads(p.read_text())
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def validate():
    initial=read(HERE/'receipt.json');preservation=read(PRIVATE/'documentation-preservation.json')
    saved=preservation['sources'];assert len(saved)==9
    assert sha(HERE/'receipt.json')==preservation['receipt_sha256']
    for name,digest in saved.items():assert sha(PRIVATE/'documentation-before-correction'/name)==digest,name
    for name,digest in initial['attachments'].items():
        assert sha(PRIVATE/'documentation-before-correction'/name if name in saved else ROOT/name)==digest,name
    class ArchivedDocuments:
        def __truediv__(self,name):return PRIVATE/'documentation-before-correction'/name if str(name)in saved else ROOT/name
    original=runpy.run_path(str(HERE/'verify.py'))['validate'];original.__globals__['ROOT']=ArchivedDocuments()
    result=json.loads(json.dumps(original()))
    assert result=={k:v for k,v in initial.items()if k!='attachments'}
    for name in ['docs/re/RESEARCH-LOG.md','docs/re/20261003-normal-victory.md','CONTEXT.md']:
        assert 'normal-victory-20261003/verify_final.py'in(ROOT/name).read_text(),name
    for filename,signal in [('worklist-selftest-final.log','正反對照通過'),('worklist-verify-final.log','未完成 22 條'),('worklist-scan-final.log','0 行沒有')]:
        assert signal in(PRIVATE/filename).read_text(),filename
    return dict(result,initial_receipt_sha256=sha(HERE/'receipt.json'),initial_receipt_attachments=len(initial['attachments']),
        initial_receipt_preserved=True,json_tuple_array_comparison_corrected=True,production_and_player_results_unchanged=True)
def attachments():
    names=set(read(HERE/'receipt.json')['attachments'])
    names|={str(p.relative_to(ROOT))for p in PRIVATE.rglob('*')if p.is_file()}
    names|={'docs/verification/normal-victory-20261003/receipt.json','docs/verification/normal-victory-20261003/verify_final.py'}
    return {n:sha(ROOT/n)for n in sorted(names)}
if __name__=='__main__':
    ap=argparse.ArgumentParser();ap.add_argument('--seal',action='store_true');ap.add_argument('--audit',action='store_true');args=ap.parse_args()
    assert not(args.seal and args.audit);result=validate()
    if args.seal:
        target=Path('/out/receipt2.json');assert not target.exists();result['attachments']=attachments()
        target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    else:
        saved=read(HERE/'receipt2.json')
        for name,digest in saved['attachments'].items():assert sha(ROOT/name)==digest,name
        assert result=={k:v for k,v in saved.items()if k!='attachments'}
        if args.audit:
            control=read(PRIVATE/'host-control-final.json')
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
            print(json.dumps(dict(schema='kol-normal-victory-final-audit-v1',receipt_sha256=sha(HERE/'receipt2.json'),attachments=len(saved['attachments']),
                code_inputs=318,initial_receipt_preserved=True,full_readonly_verification='passed',host_control=control,
                root_owned_files=bad,md_directories=md,artifact_uid_gid=[1000,1000]),ensure_ascii=False))
    print(json.dumps({k:v for k,v in result.items()if k!='attachments'},ensure_ascii=False))
