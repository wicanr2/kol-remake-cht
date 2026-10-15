from pathlib import Path
import hashlib,json
r=Path('/src');w=Path('/out');tools=r/'docs/verification/hd-equipment-runtime-20261015'
s=(tools/'save.py').read_text().replace("out=w/'equipment-formal-save'","out=w/'npc-formal-save'").replace("'formal-actions'","'baseline'").replace("'equipment-formal'","'formal-npc'")
(w/'save-npc.py').write_text(s)
s=(tools/'normal.py').read_text().replace("'equipment-formal'","'formal-npc'").replace("'formal-actions'","'baseline'")
s=s.replace("'player-pages-v2'","'npc-normal'").replace("'page-prototype-results.json'","'npc-normal-results.json'")
(w/'normal-npc.py').write_text(s)
source=(tools/'gui.py').read_text();prefix=source[:source.index('xlog = ')]
prefix=prefix.replace("DEST = OUT / os.environ.get('KOL_GUI_OUTPUT', 'gui-prototype')","DEST = OUT / ('gui-' + os.environ['KOL_NPC_GUI_CASE'])")
footer='''
xlog = (DEST / 'xvfb.log').open('w')
xserver = subprocess.Popen(['Xvfb', ':99', '-screen', '0', '2048x1536x24', '-nolisten', 'tcp'], stdout=xlog, stderr=subprocess.STDOUT)
try:
    for _ in range(50):
        if subprocess.run(['xdotool','getdisplaygeometry'],env=env,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL).returncode == 0:
            break
        time.sleep(.1)
    else:
        raise RuntimeError('Xvfb not ready')
    process = subprocess.Popen([str(OUT/'bin/formal-npc-gui'), '-test.run', '^TestHDNPCGUI$', '-test.v'], cwd=ROOT, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, bufsize=1)
    reader = threading.Thread(target=read_frames, daemon=True);reader.start()
    before=wait_frame(lambda s:s['hd'] and not settings_open(s),timeout=75)
    result=tool('xdotool','search','--onlyvisible','--pid',str(process.pid),capture_output=True,text=True)
    window=result.stdout.splitlines()[-1];tool('xdotool','windowfocus',window)
    output=[cycle('large',False)]
    old=serial;tool('xdotool','windowsize',window,'960','600');wait_frame(after=old+2)
    output.append(cycle('small',True))
    held=serial;tool('xdotool','keydown','Control_L');wait_frame(after=held+1)
    try:
        tool('xdotool','keydown','q');process.wait(timeout=15)
    finally:
        tool('xdotool','keyup','q','Control_L')
    reader.join(timeout=3)
    assert process.returncode==0 and len(events)==12
    result=dict(status='PASS controlled actual F10 GUI',case=os.environ['KOL_NPC_GUI_CASE'],rounds=output,display_events=events,full_Inn_unchanged=True,exit_code=process.returncode,original_oracle=False,normal_player_path=False)
    (DEST/'results.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\\n')
    print('GUI',result['case'],len(output),'rounds',len(events),'events; complete Inn preserved; HD roundtrip max',max(q['hd_max_channel_delta']for q in output),flush=True)
finally:
    if process is not None and process.poll() is None:
        process.terminate()
        try:process.wait(timeout=5)
        except subprocess.TimeoutExpired:process.kill();process.wait()
    xserver.terminate()
    try:xserver.wait(timeout=5)
    except subprocess.TimeoutExpired:xserver.kill();xserver.wait()
    xlog.close()
'''
(w/'gui-npc.py').write_text(prefix+footer)
(w/'regression-tool-sources.json').write_text(json.dumps({str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest()for p in [tools/'save.py',tools/'normal.py',tools/'gui.py']},indent=2)+'\n')
print('Prepared existing normal/save scripts and controlled F10 GUI driver')
