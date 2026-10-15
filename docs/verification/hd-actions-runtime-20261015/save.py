from pathlib import Path
import hashlib,json,os,subprocess,time
r=Path('/src');w=Path('/out');out=w/'formal-action-save';out.mkdir()
env=dict(os.environ,DISPLAY=':99',LIBGL_ALWAYS_SOFTWARE='1')
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
log=(out/'xvfb.log').open('w');server=subprocess.Popen(['Xvfb',':99','-screen','0','2048x1536x24','-nolisten','tcp'],stdout=log,stderr=log)
def run(name,binary,mode,args):
 d=out/name;d.mkdir();cmd=[str(w/'bin'/binary),'-lang','zh-Hant','-graphics',mode,'-seed','1','-scale','1','-music','off','-scene','play',*args,'-shot-dir',str(d/'shots'),'-trace',str(d/'trace.json'),'-text-log',str(d/'text.jsonl')]
 (d/'command.json').write_text(json.dumps(cmd,ensure_ascii=False,indent=2)+'\n')
 with (d/'run.log').open('w') as log:subprocess.run(cmd,cwd=r,env=env,stdout=log,stderr=log,timeout=120,check=True)
 return json.loads((d/'trace.json').read_text()),(d/'text.jsonl').read_text()
try:
 for _ in range(50):
  if subprocess.run(['xdotool','getdisplaygeometry'],env=env,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL).returncode==0:break
  time.sleep(.1)
 else:raise RuntimeError('Xvfb not ready')
 rest=json.loads((r/'docs/verification/player-path/town-rest.json').read_text())['script']
 results=[]
 for label,binary,mode in [('base','baseline','ega'),('formal','formal-actions','hd')]:
  save=out/(label+'-save.dat');trace,text=run('rest-'+label,binary,mode,['-play-script',rest,'-save',str(save)])
  data=json.loads(save.read_text());assert data['schema']=='kol-save-v3';data.pop('saved_at',None)
  before=sha(save);reload,rt=run('reload-'+label,binary,mode,['-load',str(save),'-play-script','SPACE 1 > RET'])
  assert sha(save)==before
  results.append(dict(trace=trace,text=text,save=data,reload=reload,reload_text=rt))
 assert results[0]==results[1]
 old=out/'base-save.dat';before=sha(old)
 oldtrace,oldtext=run('old-v3-formal','formal-actions','hd',['-load',str(old),'-play-script','SPACE 1 > RET'])
 assert sha(old)==before and oldtrace==results[0]['reload'] and oldtext==results[0]['reload_text']
 result=dict(schema='kol-hd-action-save-check/1',rest_script=rest,rest_states=len(results[1]['trace']['steps']),reload_states=len(results[1]['reload']['steps']),normal_save_equal_except_saved_at=True,reload_file_unchanged=True,old_v3_from_baseline=dict(sha256=before,trace_equal=True,text_equal=True,file_unchanged=True),limits='Normal play rest/reload only; no new completion or original parity claim.')
 (out/'results.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
 print('Normal rest/save/reload and old v3 passed',result['rest_states'],result['reload_states'],flush=True)
finally:
 server.terminate()
 try:server.wait(timeout=5)
 except subprocess.TimeoutExpired:server.kill();server.wait()
 log.close()
