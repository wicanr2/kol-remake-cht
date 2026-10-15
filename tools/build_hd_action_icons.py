#!/usr/bin/env python3
"""Rebuild SPEC-039 action assets using the existing browser image.

Run only in Docker. Mount this repository at /src read-only and a UID1000 output
directory at /out. Use logh3-hd-preview:20261011-r1, --network none, --memory 2g,
--cpus 2, --pids-limit 128 and a 120-second outer timeout. Command:
python3 /src/tools/build_hd_action_icons.py --output /out/actions-rebuilt
The output must not exist. Source comparisons are private; production uses only
actions.png, manifest.json and font-LICENSE.txt.
"""
from pathlib import Path
from PIL import Image,ImageDraw
from urllib.parse import urlparse
import argparse,ast,base64,hashlib,json,os,socket,struct,subprocess,time,urllib.request

root=Path('/src')
parser=argparse.ArgumentParser(description='Rebuild the approved 68 native action icons inside logh3-hd-preview:20261011-r1. Original PNG comparisons remain private.')
parser.add_argument('--output',type=Path,required=True)
args=parser.parse_args()
work=args.output.parent
dest=args.output
assert Path('/.dockerenv').exists() and work.stat().st_uid==work.stat().st_gid==1000
assert not dest.exists();dest.mkdir()
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
ega=root/'assets/original-derived-png/ega'
m=json.loads((ega/'manifest.json').read_text())
icons=[x for a in m['archives'] if a['archive']=='KOLMPICS' for x in a['images'] if x['unit_index']>=3 and x['unit_index']!=28]
mask=next(x for a in m['archives'] if a['archive']=='KOLMPICS' for x in a['images'] if x['unit_index']==28)
assert (mask['width'],mask['height'])==(56,57)
assert len(icons)==68 and all((x['width'],x['height'])==(24,20) for x in icons)
ids=[x['unit_index'] for x in icons];assert ids==sorted(set(ids))
for row in icons:assert sha(ega/row['file'])==row['png_sha256']
oldpath=root/'docs/verification/hd-ui-theme-20261012/review.js'
old=oldpath.read_text();a=old.index('function path(');b=old.index('function panel(',a)
c=old.index('function icon(');d=old.index('function background(',c)
canonical=old[a:b]+old[c:d]
html='''<!doctype html><meta charset="utf-8"><style>body{background:#21151a;color:#eee0ca;font:20px "Noto Serif CJK TC"}canvas{max-width:100%}</style><script>'''+canonical+'</script><script src="file:///src/tools/hd_action_icons.js"></script><script>const ids='+json.dumps(ids)+''';document.fonts.ready.then(()=>{window.iconCanvas=renderActionAtlas(ids);document.body.append(iconCanvas);document.body.dataset.ready='true'});</script>'''
(dest/'review.html').write_text(html)
helper=root/'docs/verification/hd-ui-theme-20261012/verify_browser.py'
tree=ast.parse(helper.read_text());cls=next(x for x in tree.body if isinstance(x,ast.ClassDef) and x.name=='CDP')
ns=dict(globals());exec(compile(ast.Module(body=[cls],type_ignores=[]),str(helper),'exec'),ns)
profile=Path('/tmp/kol-action-icons-v10-browser');profile.mkdir()
env=dict(os.environ)
with (dest/'chromium.log').open('w') as log:
    browser=subprocess.Popen(['chromium','--headless','--no-sandbox','--disable-dev-shm-usage','--disable-gpu','--allow-file-access-from-files','--remote-debugging-port=0','--user-data-dir='+str(profile),'about:blank'],env=env,stdout=log,stderr=log)
    try:
        for _ in range(150):
            if (profile/'DevToolsActivePort').exists():break
            assert browser.poll() is None;time.sleep(.05)
        port=int((profile/'DevToolsActivePort').read_text().splitlines()[0])
        tabs=json.load(urllib.request.urlopen('http://127.0.0.1:'+str(port)+'/json'))
        client=ns['CDP'](next(x['webSocketDebuggerUrl'] for x in tabs if x['type']=='page'))
        browser_version=client.call('Browser.getVersion')
        client.call('Page.navigate',{'url':(dest/'review.html').as_uri()})
        for _ in range(200):
            if client.evaluate('document.body && document.body.dataset.ready')=='true':break
            time.sleep(.05)
        assert client.evaluate('document.body.dataset.ready')=='true'
        encoded=client.evaluate('iconCanvas.toDataURL("image/png").split(",")[1]')
        (dest/'actions.png').write_bytes(base64.b64decode(encoded))
    finally:
        browser.terminate()
        try:browser.wait(timeout=5)
        except subprocess.TimeoutExpired:browser.kill();browser.wait(timeout=5)
atlas=Image.open(dest/'actions.png').convert('RGBA');assert atlas.size==(1296,960)
rows=[];review=Image.new('RGB',(9*144,8*184),(28,23,25));draw=ImageDraw.Draw(review)
for i,src in enumerate(icons):
    x,y=i%9*144,i//9*120;r=[x,y,x+144,y+120];pic=atlas.crop(r)
    alpha=pic.getchannel('A');bb=alpha.point(lambda v:255 if v>=8 else 0).getbbox();assert bb is not None
    assert bb[0]>0 and bb[1]>0 and bb[2]<144 and bb[3]<120,(src['unit_index'],bb)
    original=Image.open(ega/src['file']).convert('RGBA')
    rows.append(dict(index=src['unit_index'],source_rgb_sha256=hashlib.sha256(original.convert('RGB').tobytes()).hexdigest(),
                     source_png=src['file'],source_png_sha256=src['png_sha256'],source_size=[24,20],rect=r,
                     native_rgba_sha256=hashlib.sha256(pic.tobytes()).hexdigest(),alpha_bbox=list(bb)))
    cy=i//9*184;review.paste(original.resize((48,40),Image.Resampling.NEAREST),(x+8,cy+126),original.resize((48,40),Image.Resampling.NEAREST))
    review.paste(pic,(x,cy),pic);draw.text((x+63,cy+134),f'PNG {src["unit_index"]}',fill=(238,220,186))
review.save(dest/'source-native-review.png')
font=Path('/usr/share/fonts/opentype/noto/NotoSerifCJK-Regular.ttc')
assert sha(font)=='a04178ec485dffdff7cc0c0c20e1fce9202d7e2160d805e8e44a4c8841c58481'
license=root/'assets/hd/first-slice/interface/font-LICENSE.txt'
import shutil
shutil.copyfile(license,dest/'font-LICENSE.txt')
sources=[dict(archive='KOLMPICS',index=x['index'],source_rgb_sha256=x['source_rgb_sha256'],source_png_sha256=x['source_png_sha256'],rect=x['rect'],native_rgba_sha256=x['native_rgba_sha256'],alpha_bbox=x['alpha_bbox']) for x in rows]
data=dict(schema='kol-hd-actions/1',rights='Project-authored code-native icon redraw; source identities and original comparisons private GitHub only; Noto SIL OFL1.1',source_zip_sha256=m['source_zip']['sha256'],image='actions.png',files={'actions.png':dict(sha256=sha(dest/'actions.png'),width=1296,height=960)},sources=sources,font_license_sha256=sha(license),font_source_sha256=sha(font),source_manifest_sha256=sha(ega/'manifest.json'),generator=dict(code='tools/hd_action_icons.js',code_sha256=sha(root/'tools/hd_action_icons.js'),tool='tools/build_hd_action_icons.py',tool_sha256=sha(Path(__file__)),canonical_style='docs/verification/hd-ui-theme-20261012/review.js',canonical_style_sha256=sha(oldpath),browser=browser_version),excluded_data_mask=dict(index=28,size=[56,57],source_png_sha256=mask['png_sha256'],evidence='RE-012; internal/town/outdoor.go Crest'))
data['mother_entry']=dict(entry='Knights-of-legend/KOLMPICS.0',sha256='b1768c864bab065c2c59cc9d4d22ba33de1de7abe9d9d9f132472b497919756d')
(dest/'manifest.json').write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
print('Rebuilt 68 native action icons, preserved crest AND-mask28, no clipped cells.')
