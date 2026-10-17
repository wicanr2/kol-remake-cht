from pathlib import Path
import subprocess,time,signal,json,os

out=Path('/out/smoke/hd-startup-diagnosis');out.mkdir(exist_ok=True)
name=os.environ['KOL_PROBE_IMAGE']
dest=out/name;dest.mkdir()
cmd=['/out/full/kol','-scene','cover','-lang','zh-Hant','-graphics','hd','-shot',str(dest/'cover.png')]
with (dest/'run.log').open('w') as log:
    proc=subprocess.Popen(cmd,cwd='/out/full',stdout=log,stderr=subprocess.STDOUT)
    start=time.monotonic()
    while proc.poll() is None and time.monotonic()-start<35:time.sleep(.1)
    completed=proc.poll() is not None
    if not completed:
        proc.send_signal(signal.SIGQUIT)
    try:proc.wait(timeout=15)
    except subprocess.TimeoutExpired:proc.kill();proc.wait(timeout=5)
record={'command':cmd,'elapsed':time.monotonic()-start,'completed':completed,'returncode':proc.returncode,'scope':'bounded stack diagnosis; not a passing sample'}
(dest/'probe.json').write_text(json.dumps(record,indent=2)+'\n')
text=(dest/'run.log').read_text()
idx=text.find('goroutine 1 ')
print(json.dumps(record))
print(text[idx:idx+6500] if idx>=0 else text[-3500:])
