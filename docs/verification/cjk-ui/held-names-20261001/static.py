from pathlib import Path
import subprocess,json,os
root=Path('workplace/cjk-held-names-20261001')
for path in [root,Path('docs/verification'),Path('docs/verification/cjk-text-fit.json')]:
    assert (path.stat().st_uid,path.stat().st_gid)==(os.getuid(),os.getgid()),path
checks=[]
def check(label,args,expected):
    result=subprocess.run(['python3',*args],capture_output=True,text=True,timeout=40)
    assert result.returncode==expected,(label,result.stdout,result.stderr)
    rec={'label':label,'command':['python3',*args],'exit_code':result.returncode,
         'stdout':result.stdout,'stderr':result.stderr}
    checks.append(rec);print(label,result.stdout,result.stderr,flush=True)
    return result
oldtool=str(root/'before-verify-display-names.py')
oldtable=str(root/'before-display-names.tsv')
check('舊檢查器對舊表通過，證實漏掃', [oldtool,'--table',oldtable],0)
negative=check('新檢查器拒絕舊表', ['tools/verify_display_names.py','--table',oldtable],1)
assert "['Black Choker', 'Maul']" in negative.stderr
assert json.loads(negative.stdout)['problems']==1
check('新檢查器及新表無缺譯或孤兒',['tools/verify_display_names.py'],0)
check('字模涵蓋',['tools/verify_cjk_coverage.py','--report',str(root/'glyph-coverage.json')],0)
check('文字安全矩形',['tools/verify_text_fit.py',
    '--layout','workplace/cjk-layout.json','--font','assets/ui-font/manifest.json',
    '--strings','docs/i18n/newchar-zh-Hant.json',
    '--bank-strings','docs/i18n/ui-zh-Hant.json','--ui-i18n','docs/i18n/ui-zh-Hant.json',
    '--data','assets/original-derived-data/character-creation.json',
    '--town-i18n','docs/i18n/town-zh-Hant.json',
    '--display-names','docs/i18n/display-names-zh-Hant.tsv',
    '--out','docs/verification/cjk-text-fit.json'],0)
old=json.loads((root/'before-text-fit.json').read_text());new=json.loads(Path('docs/verification/cjk-text-fit.json').read_text())
assert new['checked']==old['checked']+10 and new['failed']==0
newnames={'Maul','Black Choker'}
assert [c for c in new['checks'] if c['en'] not in newnames]==old['checks']
assert len([c for c in new['checks'] if c['en'] in newnames])==10
(root/'static-checks.json').write_text(json.dumps({'checks':checks,
    'old_checks_unchanged':True,'additional_checks':10,'checked':new['checked'],
    'failed':0},ensure_ascii=False,indent=2)+'\n')
