#!/usr/bin/env python3
"""Docker內從核准的Canvas函式重生介面PNG及完整HD字模。"""
import ast, base64, hashlib, json, math, os, shutil, subprocess, time, urllib.request
from pathlib import Path
from PIL import Image, ImageDraw, ImageFont

ROOT=Path('/src'); OUT=Path('/out'); REVIEW=ROOT/'docs/verification/hd-ui-theme-20261012'
assert Path('/.dockerenv').exists()
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def write(p,d):p.write_text(json.dumps(d,ensure_ascii=False,indent=2)+'\n')

def main():
    OUT.mkdir(exist_ok=True)
    # 只載入已保存工具的 WebSocket class，不執行其中的測試／截圖頂層程式。
    source=REVIEW/'verify_browser.py'
    tree=ast.parse(source.read_text())
    cls=next(n for n in tree.body if isinstance(n,ast.ClassDef) and n.name=='CDP')
    import socket, struct
    from urllib.parse import urlparse
    ns=dict(globals(),socket=socket,struct=struct,urlparse=urlparse)
    exec(compile(ast.Module(body=[cls],type_ignores=[]),str(source),'exec'),ns)
    profile=Path('/tmp/kol-interface-export'); profile.mkdir()
    with (OUT/'bake.log').open('w') as log:
        chrome=subprocess.Popen(['chromium','--headless','--no-sandbox','--disable-dev-shm-usage','--disable-gpu','--allow-file-access-from-files','--remote-debugging-port=0','--user-data-dir='+str(profile),'about:blank'],stdout=log,stderr=log)
        try:
            for _ in range(100):
                if (profile/'DevToolsActivePort').exists():break
                assert chrome.poll() is None;time.sleep(.05)
            port=int((profile/'DevToolsActivePort').read_text().splitlines()[0])
            pages=json.load(urllib.request.urlopen('http://127.0.0.1:'+str(port)+'/json'))
            c=ns['CDP'](next(p['webSocketDebuggerUrl'] for p in pages if p['type']=='page'))
            c.call('Page.navigate',{'url':REVIEW.as_uri()+'/review.html?single=1'})
            for _ in range(200):
                if c.evaluate('document.body && document.body.dataset.ready')=='true':break
                time.sleep(.05)
            assert c.evaluate('document.body.dataset.tests')=='PASS'
            def export(name,w,h,code):
                data=c.evaluate('(()=>{const c=document.createElement("canvas");c.width='+str(w)+';c.height='+str(h)+';const ctx=c.getContext("2d");'+code+';return c.toDataURL("image/png").split(",")[1]})()')
                (OUT/name).write_bytes(base64.b64decode(data))
            export('background.png',1920,1200,'background(ctx)')
            sizes=[(256,288),(144,136),(514,372),(1166,118),(1166,406),(400,132),(1164,588),(736,450),(808,456),(878,882),(192,48),(744,712),(664,148),(432,72)]
            for w,h in sizes:
                for active in (0,1):
                    export(f'panel-{w}-{h}-{active}.png',w+8,h+8,'panel(ctx,[4,4,'+str(w)+','+str(h)+'],{selected:'+str(bool(active)).lower()+',red:'+str((w,h)==(744,712)).lower()+'})')
            for key in ['door','back','yes','sword','fist','swap','drop','move','rest','shield','arrow']:
                export('icon-'+key+'.png',100,100,'icon(ctx,"'+key+'",0,0,100)')
            export('crest.png',320,320,'ctx.translate(-130,-502);crest(ctx)')
        finally:
            chrome.terminate()
            try:chrome.wait(timeout=5)
            except subprocess.TimeoutExpired:chrome.kill();chrome.wait(timeout=5)
    fontpath=Path('/usr/share/fonts/opentype/noto/NotoSerifCJK-Regular.ttc')
    assert sha(fontpath)=='a04178ec485dffdff7cc0c0c20e1fce9202d7e2160d805e8e44a4c8841c58481'
    font=ImageFont.truetype(str(fontpath),64,index=3)
    old=json.loads((ROOT/'assets/ui-font/manifest.json').read_text())
    chars=set(chr(g['codepoint']) for entry in old['fonts'] for g in entry['glyphs'])
    chars.update(chr(n) for n in range(32,127));chars.update(json.loads((REVIEW/'font-atlas.json').read_text())['glyphs'])
    chars=sorted(chars|set('　…介面、人物與旅店同一場景，保留隊伍與操作位置'))
    atlas=Image.new('RGBA',(96*64,96*math.ceil(len(chars)/64)))
    draw=ImageDraw.Draw(atlas);glyphs={}
    for i,ch in enumerate(chars):
        x,y=i%64*96,i//64*96
        draw.text((x+10,y+76),ch,font=font,fill='white',anchor='ls')
        glyphs[ch]=dict(x=x,y=y,advance=font.getlength(ch))
    atlas.save(OUT/'font-atlas.png')
    write(OUT/'font-atlas.json',dict(schema='kol-hd-font/1',size=64,cell=96,baseline=76,xpad=10,glyphs=glyphs))
    shutil.copyfile(REVIEW/'font-LICENSE.txt',OUT/'font-LICENSE.txt')
    shutil.copyfile(REVIEW/'innkeeper-painted-v1.png',OUT/'innkeeper-painted-v1.png')
    rgb=lambda name:hashlib.sha256(Image.open(ROOT/('assets/original-derived-png/ega/TWN7PICS/'+name)).convert('RGB').tobytes()).hexdigest()
    files={p.name:dict(sha256=sha(p),width=Image.open(p).width,height=Image.open(p).height) for p in sorted(OUT.glob('*.png'))}
    write(OUT/'manifest.json',dict(schema='kol-hd-interface/1',rights='Project-authored Canvas UI; Noto SIL OFL 1.1; AI-painted NPC, private project review. Not an original soundtrack or original asset license.',files=files,font_json_sha256=sha(OUT/'font-atlas.json'),font_license_sha256=sha(OUT/'font-LICENSE.txt'),npc_rgb_sha256=rgb('twn7pics-012.png'),crest_rgb_sha256=rgb('twn7pics-000.png'),sources={str(p.relative_to(ROOT)):sha(p) for p in [REVIEW/'review.js',REVIEW/'innkeeper-painted-v1.png',ROOT/'assets/ui-font/manifest.json']},tool='tools/build_hd_interface.py',font_source_sha256=sha(fontpath),font_face=3))
    print('BAKED',len(files),'PNG',len(chars),'glyphs')
if __name__=='__main__':main()
