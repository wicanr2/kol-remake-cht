import hashlib,json,os,pathlib,re,subprocess,time,zipfile

out=pathlib.Path('/out'); work=pathlib.Path('/work')
env=dict(os.environ,DISPLAY=':99',WINEPREFIX='/tmp/kol-hotfix-wine',WINEDEBUG='-all',WINEDLLOVERRIDES='mscoree,mshtml=',PULSE_SERVER='unix:/work/smoke/wine-pulse/native',PULSE_SINK='kol_wine')
pathlib.Path(env['WINEPREFIX']).mkdir()
x=subprocess.Popen(['Xvfb',':99','-screen','0','1920x1200x24','-nolisten','tcp'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
runs=[]
def xdo(*args): return subprocess.run(['xdotool',*map(str,args)],env=env,check=True,capture_output=True,text=True,timeout=10).stdout.strip()
def wait(fn,p,log,seconds=45):
    until=time.monotonic()+seconds
    while time.monotonic()<until:
        assert p.poll() is None,log.read_text(errors='replace')
        if fn(): return
        time.sleep(.1)
    raise AssertionError('timeout '+log.read_text(errors='replace'))
def key(w,k):
    xdo('keydown','--window',w,k); time.sleep(.8); xdo('keyup','--window',w,k); time.sleep(.8)
def inputkey(w,k,scene):
    marker='input scene='+scene+' '
    count=log.read_text(errors='replace').count(marker)
    xdo('keydown','--window',w,k)
    try: wait(lambda:log.read_text(errors='replace').count(marker)>count,p,log,30)
    finally: xdo('keyup','--window',w,k)
    time.sleep(1)
def rss(p):
    s=pathlib.Path(f'/proc/{p.pid}/status').read_text()
    return int(re.search(r'^VmRSS:\s+(\d+)',s,re.M)[1])*1024
try:
    time.sleep(.5)
    for i in range(200):
        if (work/'smoke/wine-audio-ready').exists(): break
        time.sleep(.1)
    assert (work/'smoke/wine-pulse/native').is_socket()
    subprocess.run(['/usr/lib/wine/wine64','reg','add',r'HKCU\Software\Wine\Drivers','/v','Audio','/t','REG_SZ','/d','pulse','/f'],env=env,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL,check=True)
    for theme,bundle in [('ega','full'),('ega-hd','full'),('ega','release'),('ega-hd','release')]:
        zpath=next((out/bundle).glob(f'*-{bundle}-{theme}-windows-amd64.zip'))
        dest=pathlib.Path('/tmp')/('game-'+bundle+'-'+theme)
        with zipfile.ZipFile(zpath) as z: z.extractall(dest)
        root=next(dest.iterdir()); exe=root/'kol.exe'
        if bundle=='release':
            subprocess.run(['python3',str(root/'tools/import_game_data.py'),'--source','/src','--target',str(root)],check=True,timeout=120)
        assert (root/'assets/hd/first-slice/manifest.json').exists()==(theme=='ega-hd')
        version=subprocess.check_output(['/usr/lib/wine/wine64',str(exe),'-version'],cwd=root,env=env,text=True,timeout=20).strip()
        assert version=='v.1.0.4-20261017',version
        name=bundle+'-'+theme
        log=work/(name+'-gui.log'); memory=[]; timings=[]
        with log.open('w') as stream:
            p=subprocess.Popen(['/usr/lib/wine/wine64',str(exe),'-seed','1','-log-input','-save',f'Z:\\tmp\\save-{theme}.json'],cwd=root,env=env,stdout=stream,stderr=subprocess.STDOUT)
            text=lambda:log.read_text(errors='replace')
            win=[]
            def find():
                result=subprocess.run(['xdotool','search','--onlyvisible','--name','^武士傳說$'],env=env,capture_output=True,text=True)
                if result.returncode==0: win[:]=result.stdout.splitlines()[-1:]
                return bool(win)
            wait(find,p,log); w=win[0]; xdo('windowfocus','--sync',w)
            wait(lambda:'music cue=title' in text(),p,log)
            memory.append(dict(stage='title',rss=rss(p)))
            inputkey(w,'space','title')
            inputkey(w,'Return','cover')
            inputkey(w,'5','contents')
            key(w,'F10'); wait(lambda:'display-frame hd=false open=true' in text(),p,log)
            key(w,'Down'); start=time.monotonic(); key(w,'Return')
            if theme=='ega':
                wait(lambda:'display action=hd-unavailable hd=false' in text(),p,log)
                assert 'display-frame hd=true' not in text()
            else:
                wait(lambda:'display-frame hd=true open=true cursor=1' in text(),p,log,60)
                timings.append(time.monotonic()-start)
                memory.append(dict(stage='first-hd',rss=rss(p)))
                for round in range(3 if bundle=='full' else 1):
                    n=text().count('display-frame hd=false open=true cursor=0')
                    key(w,'Up'); key(w,'Return')
                    wait(lambda:text().count('display-frame hd=false open=true cursor=0')>n,p,log)
                    memory.append(dict(stage=f'ega-{round}',rss=rss(p)))
                    n=text().count('display-frame hd=true open=true cursor=1')
                    key(w,'Down'); start=time.monotonic(); key(w,'Return')
                    wait(lambda:text().count('display-frame hd=true open=true cursor=1')>n,p,log)
                    timings.append(time.monotonic()-start)
                    memory.append(dict(stage=f'hd-{round}',rss=rss(p)))
            states=re.findall(r'display action=\S+ hd=\S+ state=([0-9a-f]+)',text())
            assert len(set(states))==1,'F10 changed formal game state'
            key(w,'Escape')
            inputkey(w,'Down','menu'); inputkey(w,'Return','menu')
            inputkey(w,'Escape','newchar'); inputkey(w,'Escape','menu')
            inputkey(w,'6','contents'); inputkey(w,'Return','select')
            key(w,'F9'); key(w,'F9')
            wait(lambda:'display action=music-true' in text(),p,log)
            memory.append(dict(stage='party-inn',rss=rss(p)))
            xdo('keydown','--window',w,'ctrl'); xdo('keydown','--window',w,'q')
            code=p.wait(timeout=15); xdo('keyup','q','ctrl'); assert code==0,(code,text())
            assert not pathlib.Path(f'/tmp/save-{theme}.json').exists(),'display controls wrote save'
            assert 'HD畫面載入失敗' not in text() and 'panic:' not in text(),text()
        # Deterministic same flow checkpoints, using the actual shipped PE.
        shots=work/(name+'-flow'); shots.mkdir()
        win=lambda path:'Z:'+str(path).replace('/','\\')
        with (work/(name+'-flow.log')).open('w') as stream:
            subprocess.run(['/usr/lib/wine/wine64',str(exe),'-scene','flow','-lang','zh-Hant','-graphics','hd' if theme=='ega-hd' else 'ega','-seed','1','-flow-script','RET RET 5 D RET','-shot-dir',win(shots),'-trace',win(work/(name+'-trace.json')),'-text-log',win(work/(name+'-text.jsonl'))],cwd=root,env=env,stdout=stream,stderr=subprocess.STDOUT,check=True,timeout=90)
        runs.append(dict(theme=theme,bundle=bundle,package=zpath.name,sha256=hashlib.file_digest(zpath.open('rb'),'sha256').hexdigest(),version=version,exit_code=code,rss_samples=memory,hd_seconds=timings,no_save_written=True,normal_intro=True,hd_disabled=theme=='ega',f10_state_unchanged=True))
    # No diagnostic flag: exercise the double-click logging path too.
    log=work/'default-stdout.log'
    with log.open('w') as stream:
        p=subprocess.Popen(['/usr/lib/wine/wine64',str(exe)],cwd=root,env=env,stdout=stream,stderr=subprocess.STDOUT)
        win=[]; wait(find,p,log); w=win[0]; xdo('windowfocus','--sync',w)
        for k in ['space','Return','5']: key(w,k)
        xdo('keydown','--window',w,'ctrl'); xdo('keydown','--window',w,'q')
        assert p.wait(timeout=15)==0
        xdo('keyup','q','ctrl')
    logs=list(pathlib.Path(env['WINEPREFIX']).rglob('kol.log')); assert len(logs)==1
    default_log=logs[0].read_text(errors='replace'); assert 'v.1.0.4-20261017' in default_log and 'panic:' not in default_log
    (work/'default-kol.log').write_text(default_log)
    # Corrupt only an extracted test copy, after the complete ZIP passed.
    art=root/'assets/hd/first-slice/pages/book-open-painted.png'; original=art.read_bytes()
    art.write_bytes(original[:-1]+bytes([original[-1]^1]))
    log=work/'corrupt-hd-gui.log'
    with log.open('w') as stream:
        p=subprocess.Popen(['/usr/lib/wine/wine64',str(exe),'-log-input','-text-log','Z:\\work\\corrupt-hd-text.jsonl'],cwd=root,env=env,stdout=stream,stderr=subprocess.STDOUT)
        win=[]; wait(find,p,log); w=win[0]; xdo('windowfocus','--sync',w)
        wait(lambda:'music cue=title' in log.read_text(errors='replace'),p,log)
        inputkey(w,'space','title'); inputkey(w,'Return','cover'); inputkey(w,'5','contents')
        key(w,'F10'); key(w,'Down'); key(w,'Return')
        wait(lambda:'HD畫面載入失敗，返回EGA' in log.read_text(errors='replace'),p,log)
        wait(lambda:'手繪圖片讀取失敗，已回到經典畫面。' in (work/'corrupt-hd-text.jsonl').read_text(),p,log)
        assert p.poll() is None
        count=log.read_text(errors='replace').count('display action=toggle hd=false')
        key(w,'F10')
        wait(lambda:log.read_text(errors='replace').count('display action=toggle hd=false')>count,p,log)
        xdo('keydown','--window',w,'ctrl'); xdo('keydown','--window',w,'q')
        assert p.wait(timeout=15)==0; xdo('keyup','q','ctrl')
    art.write_bytes(original)
    (work/'wine-runtime.json').write_text(json.dumps(dict(status='PASS',runs=runs,memory_events=pathlib.Path('/sys/fs/cgroup/memory.events').read_text(),memory_peak=int(pathlib.Path('/sys/fs/cgroup/memory.peak').read_text()),default_logging=True,corrupt_hd_returns_to_ega=True,real_windows_machine=False),ensure_ascii=False,indent=2)+'\n')
finally:
    subprocess.run(['wineserver','-k'],env=env,check=False)
    x.terminate(); x.wait()
    (work/'smoke/wine-audio-done').touch()
