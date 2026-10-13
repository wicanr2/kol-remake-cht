#!/usr/bin/env python3
"""Docker限定：實際Linux封包的旅店、F9／F10、縮放與音訊。"""
import hashlib,json,re,shutil,time
from pathlib import Path
import numpy as np
import verify_music_runtime as player
ROOT=Path('/src');OUT=Path('/out')
def check(bundle,command,fixture,base):
    dest=OUT/'smoke'/('linux-'+bundle+'-controls');dest.mkdir();(dest/'inputs').mkdir()
    shutil.copyfile(fixture,dest/'inputs/save.json')
    wrapper=dest/'kol';wrapper.write_text(command);wrapper.chmod(0o755)
    player.OUT=dest
    def body(proc,w,text,markers):
        def action(fn,needle):
            before=len(text());fn();player.waitfor(lambda:needle in text()[before:],20)
        player.key(w,'Left','Right','Return','space')
        player.waitfor(lambda:'music cue=inn enabled=true' in text(),20);time.sleep(2)
        player.shot('inn-hd.png')
        action(lambda:player.key(w,'F10'),'display-frame hd=true open=true')
        action(lambda:player.key(w,'Up','Return'),'display-frame hd=false open=true');player.shot('menu-ega.png')
        action(lambda:player.key(w,'Down','Return'),'display-frame hd=true open=true');player.shot('menu-hd.png')
        action(lambda:player.key(w,'Escape'),'display-frame hd=true open=false')
        action(lambda:player.key(w,'F9'),'display action=music-false');time.sleep(3.5)
        action(lambda:player.key(w,'F9'),'display action=music-true');time.sleep(2)
        player.xdo('windowsize',w,960,600);time.sleep(.5)
        action(lambda:player.click(w,792,588),'display action=music-false');time.sleep(.8)
        action(lambda:player.click(w,792,588),'display action=music-true');time.sleep(2)
        player.shot('inn-small.png',960,600)
        states=re.findall(r'display action=\S+ hd=\S+ state=([0-9a-f]+)',text());assert len(states)==8 and len(set(states))==1
        action(lambda:player.key(w,'F9'),'display action=music-false');time.sleep(3.5)
        markers.append(dict(default_music_on=True,normal_inn=True,f9=True,f10=True,resized_mouse=True,state_sha256=states[0],original_parity_claimed=False))
    report,samples,log=player.record('package-'+bundle,body,dest/'inputs/save.json')
    assert np.max(np.abs(samples.astype(np.int32)))>100 and np.max(np.abs(samples[-44100:].astype(np.int32)))==0
    report.update(status='PASS',bundle=bundle,version=json.loads((OUT/'build-inputs.json').read_text())['version'])
    (dest/'controls.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    return report

def main():
    assert Path('/.dockerenv').exists()
    fixture=OUT/'smoke/linux-gui/inputs/save.json';assert fixture.is_file()
    version=json.loads((OUT/'build-inputs.json').read_text())['version']
    full='#!/bin/sh\nexec /out/full/kol-'+version+'-full-linux-x86_64.AppImage --appimage-extract-and-run "$@"\n'
    release='#!/bin/sh\ncd /tmp/imported || exit 1\nexec /tmp/release-extract/squashfs-root/usr/bin/kol "$@"\n'
    runs=[check('full',full,fixture,OUT/'full'),check('release',release,fixture,OUT/'release')]
    (OUT/'smoke/linux-controls.json').write_text(json.dumps(dict(status='PASS',version=version,runs=runs,host_fuse=False,device='PulseAudio null sink monitor; actual package Ebitengine/Oto output'),ensure_ascii=False,indent=2)+'\n')
    print('LINUX PACKAGE CONTROLS PASS: full/release, normal inn, real audio, F9/F10, resized mouse, unchanged state',flush=True)
if __name__=='__main__':main()
