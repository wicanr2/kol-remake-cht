#!/usr/bin/env python3
"""在既有 Chromium Docker 內，以實際瀏覽器輸入驗證候選及擷取比較圖。"""
import base64,hashlib,json,os,socket,struct,subprocess,time,urllib.request
from pathlib import Path
from urllib.parse import urlparse

assert Path('/.dockerenv').exists()
OUT=Path('/out');CHECKS=[]
PAGE='file:///src/workplace/hd-ui-theme-20261012/review.html'
if not Path('/src/workplace/hd-ui-theme-20261012/review.html').exists():
    PAGE='file:///src/docs/verification/hd-ui-theme-20261012/review.html'

class CDP:
    def __init__(self,url):
        u=urlparse(url);self.sock=socket.create_connection((u.hostname,u.port),timeout=10);self.seq=0
        key=base64.b64encode(os.urandom(16)).decode()
        request='GET '+u.path+' HTTP/1.1\r\nHost: '+u.netloc+'\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: '+key+'\r\nSec-WebSocket-Version: 13\r\n\r\n'
        self.sock.sendall(request.encode());header=b''
        while not header.endswith(b'\r\n\r\n'):header+=self.sock.recv(1)
        assert header.split(b'\r\n',1)[0].split()[1]==b'101',header
        expected=base64.b64encode(hashlib.sha1((key+'258EAFA5-E914-47DA-95CA-C5AB0DC85B11').encode()).digest())
        assert expected in header
    def read(self,n):
        data=b''
        while len(data)<n:
            chunk=self.sock.recv(n-len(data));assert chunk,'websocket closed';data+=chunk
        return data
    def call(self,method,params=None):
        self.seq+=1;payload=json.dumps({'id':self.seq,'method':method,'params':params or {}}).encode();n=len(payload)
        head=b'\x81'+(bytes([0x80|n]) if n<126 else b'\xfe'+struct.pack('!H',n) if n<65536 else b'\xff'+struct.pack('!Q',n))
        mask=os.urandom(4);self.sock.sendall(head+mask+bytes(v^mask[i%4] for i,v in enumerate(payload)))
        while True:
            a,b=self.read(2);n=b&127
            if n==126:n=struct.unpack('!H',self.read(2))[0]
            elif n==127:n=struct.unpack('!Q',self.read(8))[0]
            data=self.read(n)
            assert (a&15)==1,(a,b)
            result=json.loads(data)
            if result.get('id')==self.seq:
                assert 'error' not in result,result
                return result.get('result',{})
    def evaluate(self,code):
        r=self.call('Runtime.evaluate',{'expression':code,'returnByValue':True})
        assert 'exceptionDetails' not in r,r
        return r.get('result',{}).get('value')

def check(name,passed,**details):
    CHECKS.append({'name':name,'passed':bool(passed),**details});assert passed,name

profile=Path('/tmp/kol-hd-ui-review-chromium');profile.mkdir()
log=(OUT/'chromium.log').open('w')
chrome=subprocess.Popen(['chromium','--headless','--no-sandbox','--disable-dev-shm-usage','--disable-gpu','--hide-scrollbars','--allow-file-access-from-files','--remote-debugging-port=0','--user-data-dir='+str(profile),'about:blank'],stdout=log,stderr=log)
try:
    active=profile/'DevToolsActivePort'
    for _ in range(100):
        if active.exists():break
        assert chrome.poll() is None;time.sleep(.05)
    port=int(active.read_text().splitlines()[0])
    pages=json.load(urllib.request.urlopen('http://127.0.0.1:'+str(port)+'/json',timeout=5))
    c=CDP(next(p['webSocketDebuggerUrl'] for p in pages if p['type']=='page'))
    c.call('Page.enable');c.call('Runtime.enable')
    def size(w,h):c.call('Emulation.setDeviceMetricsOverride',{'width':w,'height':h,'deviceScaleFactor':1,'mobile':False})
    def navigate(query,w=1920,h=1200):
        size(w,h);c.call('Page.navigate',{'url':PAGE+'?'+query})
        for _ in range(100):
            ready=c.evaluate('document.body && document.body.dataset.ready')
            if ready=='true':break
            failure=c.evaluate('document.body && document.body.dataset.tests')
            assert failure!='FAIL',c.evaluate('document.querySelector("output").textContent')
            time.sleep(.05)
        assert ready=='true','page did not become ready'
        check('browser selftest '+query,c.evaluate('document.body.dataset.tests')=='PASS',checks=c.evaluate('window.hdReview.checks.length'))
    def key(name,code,text=None):
        p={'type':'keyDown','key':name,'windowsVirtualKeyCode':code,'nativeVirtualKeyCode':code}
        if text:p['text']=text;p['unmodifiedText']=text
        c.call('Input.dispatchKeyEvent',p);c.call('Input.dispatchKeyEvent',{'type':'keyUp','key':name,'windowsVirtualKeyCode':code,'nativeVirtualKeyCode':code})
    def click(x,y):
        for kind in ('mousePressed','mouseReleased'):c.call('Input.dispatchMouseEvent',{'type':kind,'x':x,'y':y,'button':'left','clickCount':1})
    def shot(name):
        c.call('Runtime.evaluate',{'expression':'window.hdReview.render()'});data=c.call('Page.captureScreenshot',{'format':'png','captureBeyondViewport':False,'fromSurface':True})
        path=OUT/name;path.write_bytes(base64.b64decode(data['data']));return hashlib.sha256(path.read_bytes()).hexdigest()
    navigate('capture=single&scene=inn')
    key('6',54,'6');check('real key selects sixth party member',c.evaluate('window.hdReview.state.member')==5)
    identity=c.evaluate('window.hdReview.gameIdentity()');key('F10',121);check('real F10 opens settings',c.evaluate('window.hdReview.state.menu') is True)
    key('ArrowUp',38);key('Enter',13);check('real keyboard selects EGA',c.evaluate('window.hdReview.state.theme')=='ega')
    key('ArrowDown',40);key('Enter',13);key('Escape',27)
    check('real keyboard returns HD and closes settings',c.evaluate('window.hdReview.state.theme==="hd" && !window.hdReview.state.menu'))
    check('real theme roundtrip keeps scene party command',c.evaluate('window.hdReview.gameIdentity()')==identity)
    for w,h in [(1920,1200),(1280,800),(960,600),(640,400)]:
        size(w,h)
        click(1800*w/1920,1176*h/1200);check('mouse opens settings '+str(w),c.evaluate('window.hdReview.state.menu') is True)
        click(1500*w/1920,570*h/1200);check('mouse selects EGA '+str(w),c.evaluate('window.hdReview.state.theme')=='ega')
        click(1500*w/1920,738*h/1200);check('mouse selects HD '+str(w),c.evaluate('window.hdReview.state.theme')=='hd')
        click(1600*w/1920,988*h/1200);check('mouse closes settings '+str(w),c.evaluate('window.hdReview.state.menu') is False)
        click(780*w/1920,150*h/1200);check('mouse selects third member '+str(w),c.evaluate('window.hdReview.state.member')==2)
    navigate('capture=single&scene=battle')
    for i in range(7):
        click(264+i*192,1088);check('real command click '+str(i),c.evaluate('window.hdReview.state.command')==i)
    screenshots={}
    for scene in ('inn','door','battle'):
        navigate('capture=single&scene='+scene);screenshots[scene+'-hd.png']=shot(scene+'-hd.png')
        navigate('capture=pair&scene='+scene,3840,1264);screenshots[scene+'-compare.png']=shot(scene+'-compare.png')
    navigate('capture=single&scene=inn&menu=1');screenshots['inn-hd-menu.png']=shot('inn-hd-menu.png')
    navigate('capture=pair&scene=inn&baseline=ega',3840,1264);screenshots['inn-ega-compare.png']=shot('inn-ega-compare.png')
    report={'schema':'kol-hd-ui-browser-review/1','status':'PASS','production_changed':False,'tool':c.call('Browser.getVersion')['product'],'input':'Chrome DevTools Protocol Input.dispatchKeyEvent / Input.dispatchMouseEvent','checks':CHECKS,'screenshots':screenshots,'limits':['Disposable browser prototype only; not gameplay or production HUD acceptance.','Healthy body silhouette has no validated injury mask.','No save file or RNG is created or changed.']}
    (OUT/'browser-check.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print('BROWSER PASS',len(CHECKS),'checks,',len(screenshots),'screenshots',flush=True)
finally:
    if chrome.poll() is None:
        chrome.terminate()
        try:chrome.wait(timeout=5)
        except subprocess.TimeoutExpired:chrome.kill();chrome.wait(timeout=5)
    log.close()
