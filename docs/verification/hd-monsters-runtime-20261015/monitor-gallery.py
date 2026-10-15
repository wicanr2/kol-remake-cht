from pathlib import Path
import os,json,subprocess,time
w=Path('/out');peak=0;samples=[];pid=None
with (w/'run-formal-gallery-v3.log').open('w')as log:
    p=subprocess.Popen(['/out/bin/formal-monster-gui-v3','-test.run','^TestMonsterGallery$','-test.v'],stdout=log,stderr=log,env=dict(os.environ))
    start=time.monotonic()
    try:
        while p.poll()is None:
            if time.monotonic()-start>900:raise TimeoutError('gallery bound900s')
            current=Path('/sys/fs/cgroup/memory.current')
            if current.exists():
                n=int(current.read_text());peak=max(peak,n)
                if not samples or time.monotonic()-samples[-1][0]>=10:samples.append([round(time.monotonic()-start,2),n])
            time.sleep(1)
    finally:
        if p.poll()is None:p.terminate();p.wait(timeout=5)
result=dict(exit_code=p.returncode,peak_cgroup_bytes=peak,samples=samples,GOMEMLIMIT=os.environ.get('GOMEMLIMIT'),capture_models=1,reused_capture_surface=True,runtime_modified=False)
(w/'gallery-v3-memory.json').write_text(json.dumps(result,indent=2)+'\n')
print(result['exit_code'],'gallery peak bytes',peak,flush=True)
assert p.returncode==0
