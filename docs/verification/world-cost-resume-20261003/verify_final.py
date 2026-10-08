"""保留初封身分，重驗段落校正後的最終文件與同一程式結果。"""
import argparse,hashlib,json,os,runpy
from pathlib import Path
ROOT=Path('/src')
PRIVATE=ROOT/'workplace/world-cost-resume-20261003'
HERE=ROOT/'docs/verification/world-cost-resume-20261003'
def read(p):return json.loads(p.read_text())
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def validate():
    preservation=read(PRIVATE/'documentation-preservation.json');saved=preservation['sources']
    assert len(saved)==7 and preservation['schema']=='kol-world-cost-documentation-preservation-v1'
    initial=read(HERE/'receipt.json');assert sha(HERE/'receipt.json')==preservation['receipt_sha256']
    for name,digest in saved.items():assert sha(PRIVATE/'documentation-before-correction'/name)==digest,name
    for name,digest in initial['attachments'].items():
        assert sha(PRIVATE/'documentation-before-correction'/name if name in saved else ROOT/name)==digest,name
    class ArchivedDocuments:
        def __truediv__(self,name):return PRIVATE/'documentation-before-correction'/name if str(name) in saved else ROOT/name
    previous=runpy.run_path(str(HERE/'verify_cost.py'))['validate'];previous.__globals__['ROOT']=ArchivedDocuments()
    result=previous();assert result=={k:v for k,v in initial.items() if k!='attachments'}
    index=(ROOT/'docs/re/RESEARCH-LOG.md').read_text()
    assert '20261002-production-return-scratch.md' in index and 'world-cost-resume-20261003/verify_final.py' in index
    spec=(ROOT/'docs/spec/011-wilderness-ambush.md').read_text()
    assert spec.index('### RE-194')<spec.index('驗收：將903筆')<spec.index('### RE-195')<spec.index('## 說明書對照')
    oldspec=(PRIVATE/'documentation-before-correction/docs/spec/011-wilderness-ambush.md').read_text()
    block='驗收：將903筆原版觀察餵給只附加唯讀記錄的正式碼副本，逐筆比對；正式碼另跑\n全部internal／cmd測試及正常雙語戰鬥／保存重播，既有人物、骰序與PNG應保持。\n完整世界扣時刻仍需讀取契約與正常垂直鏈，不能因本契約通過就宣稱成本已完成。\n拾取問句／文字長度／子選單的中途暫存及未覆蓋的表外來源仍DRAFT，不猜補。\n\n'
    assert spec.count(block)==oldspec.count(block)==1
    restored=spec.replace(block,'').replace('[最終唯讀重驗](../verification/world-cost-resume-20261003/verify_final.py)',
        '[唯讀重驗](../verification/world-cost-resume-20261003/verify_cost.py)')
    assert restored==oldspec.replace(block,'')
    for filename,signal in [('worklist-selftest-final.log','正反對照通過'),('worklist-scan-final.log','0 行沒有'),('worklist-verify-final.log','未完成 22 條')]:
        assert signal in (PRIVATE/filename).read_text(),filename
    return dict(result,initial_receipt_sha256=sha(HERE/'receipt.json'),initial_receipt_attachments=len(initial['attachments']),
        initial_receipt_preserved=True,documentation_section_order_corrected=True,production_and_player_results_unchanged=True)
def attachments():
    names=set(read(HERE/'receipt.json')['attachments'])
    names|={str(p.relative_to(ROOT)) for p in PRIVATE.rglob('*') if p.is_file()}
    names|={'WORKLOG.md','docs/verification/world-cost-resume-20261003/receipt.json','docs/verification/world-cost-resume-20261003/verify_final.py'}
    return {name:sha(ROOT/name) for name in sorted(names)}
if __name__=='__main__':
    ap=argparse.ArgumentParser();ap.add_argument('--seal',action='store_true');ap.add_argument('--audit',action='store_true');args=ap.parse_args()
    assert not(args.seal and args.audit);result=validate()
    if args.seal:
        result['attachments']=attachments();target=Path('/out/receipt2.json');assert not target.exists()
        target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    else:
        receipt=read(HERE/'receipt2.json')
        for name,digest in receipt['attachments'].items():assert sha(ROOT/name)==digest,name
        assert result=={k:v for k,v in receipt.items() if k!='attachments'}
        if args.audit:
            control=read(PRIVATE/'host-control-final.json');assert control['git_diff_check']=='passed' and control['git_staged']==control['project_containers']==[]
            bad=[];md=[]
            for directory,dirs,files in os.walk(ROOT):
                for name in dirs+files:
                    p=Path(directory)/name
                    if p.lstat().st_uid==0:bad.append(str(p))
                    if p.is_dir() and p.name.endswith('.md'):md.append(str(p))
            assert not bad and not md
            for name in receipt['attachments']:
                q=(ROOT/name).stat();assert(q.st_uid,q.st_gid)==(1000,1000),name
            print(json.dumps(dict(schema='kol-world-cost-resume-final-audit-v1',receipt_sha256=sha(HERE/'receipt2.json'),
                attachments=len(receipt['attachments']),initial_receipt_preserved=True,full_readonly_verification='passed',
                host_control=control,root_owned_files=bad,md_directories=md,artifact_uid_gid=[1000,1000]),ensure_ascii=False))
    print(json.dumps({k:v for k,v in result.items() if k!='attachments'},ensure_ascii=False))
