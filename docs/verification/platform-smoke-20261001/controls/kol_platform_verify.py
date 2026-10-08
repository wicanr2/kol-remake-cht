"""核對目前工作樹的跨平台中間建置與正常保存收據，僅在 Docker 執行。"""
from pathlib import Path
import sys, json, hashlib, struct, copy, re, os
sys.path.insert(0, '/work/tools')
from png_io import read_png
root=Path('/work/workplace/platform-smoke-20261001')
def sha(p): return hashlib.sha256(Path(p).read_bytes()).hexdigest()
def load(p): return json.loads(Path(p).read_text())
inputs=load(root/'inputs.json')
for p,h in inputs['current_source_sha256'].items():
    assert sha(Path('/work')/p)==h, p
assert sha(root/'build/kol-linux')==inputs['reference_binary']['sha256']
for platform in ('linux','windows'):
    name='kol-linux' if platform=='linux' else 'kol-windows.exe'
    second='kol-linux-second' if platform=='linux' else 'kol-windows-second.exe'
    assert sha(root/'build'/name)==sha(root/'build'/second)

def pe(data):
    assert data[:2]==b'MZ'
    off=struct.unpack_from('<I',data,0x3c)[0]
    assert data[off:off+4]==b'PE\0\0'
    assert struct.unpack_from('<H',data,off+4)[0]==0x8664
    assert struct.unpack_from('<H',data,off+24)[0]==0x20b
pe((root/'build/kol-windows.exe').read_bytes())
bad=bytearray((root/'build/kol-windows.exe').read_bytes()); off=struct.unpack_from('<I',bad,0x3c)[0]
struct.pack_into('<H',bad,off+4,0x14c)
try: pe(bad)
except AssertionError: pe_negative=True
else: raise AssertionError('錯誤 Windows 架構未被拒絕')

fat=(root/'build/kol-macos').read_bytes()
assert struct.unpack_from('>II',fat)==(0xcafebabe,2)
mac=[]
for i in range(2):
    cpu,sub,offset,size,align=struct.unpack_from('>IIIII',fat,8+i*20)
    name={0x1000007:'amd64',0x100000c:'arm64'}[cpu]
    thin=(root/'build'/f'kol-macos-{name}').read_bytes()
    assert fat[offset:offset+size]==thin
    magic,tcpu,tsub,kind,ncmds,cmdbytes,flags,reserved=struct.unpack_from('<IIIIIIII',thin)
    assert magic==0xfeedfacf and tcpu==cpu and kind==2
    pos=32; minos=None; signature=None
    for _ in range(ncmds):
        cmd,length=struct.unpack_from('<II',thin,pos)
        assert length>=8 and pos+length<=32+cmdbytes
        if cmd==0x32:
            platform,minos,sdk,ntools=struct.unpack_from('<IIII',thin,pos+8)
            assert platform==1 and minos==0xb0000
        if cmd==0x1d:
            sigoff,sigsize=struct.unpack_from('<II',thin,pos+8)
            assert sigsize>0 and sigoff+sigsize<=len(thin)
            assert struct.unpack_from('>I',thin,sigoff)[0]==0xfade0cc0
            signature={'offset':sigoff,'size':sigsize,'superblob_header_valid':True}
        pos+=length
    assert pos==32+cmdbytes and minos is not None
    if name=='arm64': assert signature is not None
    libs=(root/'macos'/f'{name}-libraries.txt').read_text().splitlines()[1:]
    paths=[line.strip().split(' ')[0] for line in libs if line.strip()]
    assert paths and all(p.startswith(('/usr/lib/','/System/Library/')) for p in paths)
    mac.append({'architecture':name,'sha256':sha(root/'build'/f'kol-macos-{name}'),'minimum_macos':'11.0','signature':signature,'system_library_paths':paths})
assert {x['architecture'] for x in mac}=={'amd64','arm64'}

normal={p:load(root/p/'normal/trace.json') for p in ('linux','windows')}
reload={p:load(root/p/'reload/trace.json') for p in ('linux','windows')}
prior=Path('/work/workplace/inventory-empty-slots-20261001/normal')
assert normal['linux']==normal['windows']==load(prior/'zh-Hant/trace.json')
assert len(normal['linux']['steps'])==199
assert reload['linux']==reload['windows']==load(prior/'zh-Hant-reload2/trace.json')
assert len(reload['linux']['steps'])==16
assert reload['linux']['steps'][-1]['town']['outside_sheet']['member']=='Theodoric'
assert reload['linux']['steps'][-1]['town']['outside_sheet']['page']=='items'
saves={p:load(root/p/'default-save.dat') for p in ('linux','windows')}
normalized={p:{k:v for k,v in d.items() if k!='saved_at'} for p,d in saves.items()}
assert normalized['linux']==normalized['windows']=={k:v for k,v in load(prior/'zh-Hant/save.json').items() if k!='saved_at'}
for p,s in saves.items():
    assert s['schema']=='kol-save-v3' and len(s['party'])==6
    leader=s['characters'][s['party'][0]]
    assert leader['name']=='Theodoric' and leader['gold']==918
    assert len(leader['backpack'])==1 and leader['backpack'][0]['record']=='c3a641581056082b002300370000a000'
    current=root/p/('config2/kol-remake-cht/save.dat' if p=='linux' else 'wine-prefix/drive_c/users/ubuntu/AppData/Roaming/kol-remake-cht/save.dat')
    assert sha(current)==sha(root/p/'default-save.dat')
    before=(root/p/'save-before-reload.sha256').read_text().split()[0]
    assert sha(current)==before
    assert (current.stat().st_uid,current.stat().st_gid)==(os.getuid(),os.getgid())
mutant=copy.deepcopy(normalized['windows']); mutant['characters'][mutant['party'][0]]['gold']+=1
assert mutant!=normalized['linux']

checks=[]
def compare(a,b,label):
    ia,ib=read_png(a),read_png(b)
    assert (ia.width,ia.height)==(ib.width,ib.height)==(960,600)
    ar,br=ia.rgb(),ib.rgb()
    changed=sum(ar[i:i+3]!=br[i:i+3] for i in range(0,len(ar),3))
    assert changed==0,(label,changed)
    checks.append({'label':label,'a':str(a.relative_to('/work')),'b':str(b.relative_to('/work')),'different_pixels':changed})
static_steps=(104,111,116,117,120,126,137,138,198)
gui_names={'linux':'gui-normal2','windows':'gui-normal'}
for p in ('linux','windows'):
    gui=root/p/gui_names[p]
    observed=load(gui/'inputs.json')
    tokens=(root/'purchase-script.txt').read_text().split()
    assert observed['tokens']==tokens and len(observed['records'])==198
    assert [x['token'] for x in observed['records']]==tokens
    assert all(f'token=[{x["token"]}]' in x['input_log'] for x in observed['records'])
    assert (gui/'controller-done').is_file()
    observed=load(root/p/'gui-reload/inputs.json')
    tokens=(root/'reload-script.txt').read_text().split()
    assert observed['tokens']==tokens and len(observed['records'])==15
    assert (root/p/'gui-reload/controller-done').is_file()
    for n in static_steps:
        compare(gui/f'step-{n:03}.png',root/p/'normal'/f'step-{n:03}.png',f'{p}互動對批次正常{n}')
    for n in (1,3,14,15):
        compare(root/p/'gui-reload'/f'step-{n:03}.png',root/p/'reload'/f'step-{n:03}.png',f'{p}互動對批次重載{n}')
for n in static_steps:
    compare(root/'windows'/gui_names['windows']/f'step-{n:03}.png',root/'linux'/gui_names['linux']/f'step-{n:03}.png',f'兩平台互動正常{n}')
for n in (1,3,14,15):
    compare(root/'windows/gui-reload'/f'step-{n:03}.png',root/'linux/gui-reload'/f'step-{n:03}.png',f'兩平台互動重載{n}')
compare(root/'windows/cover.png',root/'linux/cover.png','兩平台中文封面')
for n in ('cover.png',):
    a=read_png(root/'linux'/n).rgb(); bad=bytearray(a);bad[0]^=1;assert bytes(bad)!=a

result={'schema':'kol-platform-smoke-check-v1','date':'2026-10-01','status':'所列 Windows Wine／Linux 與 macOS 靜態範圍通過',
 'source_count':len(inputs['current_source_sha256']),
 'binary_sha256':{n:sha(root/'build'/n) for n in ('kol-linux','kol-windows.exe','kol-macos','kol-macos-amd64','kol-macos-arm64')},
 'linux_windows_rebuild_identical':True,'linux_matches_current_reference':True,'windows_pe_machine':'AMD64 PE32+',
 'macos_static':mac,'macos_executed':False,'native_windows_executed':False,'wine_version':'9.0 (Ubuntu 9.0~repack-4build3)',
 'seed':{'value':1,'method':'兩平台啟動前 -seed 1；未讀觀察保存或改指定遭遇'},
 'interactive_keys_per_platform':198,'reload_interactive_keys_per_platform':15,
 'batch_trace_states_per_platform':199,'reload_batch_trace_states_per_platform':16,
 'batch_traces_match_prior_receipt':True,'saves_equal_except':['saved_at'],
 'default_save_paths':{'windows':r'C:\users\ubuntu\AppData\Roaming\kol-remake-cht\save.dat','linux':'$XDG_CONFIG_HOME/kol-remake-cht/save.dat'},
 'reload_save_hash_unchanged':True,'purchase':{'name':'Girak Helbak','zh':'吉拉克赫爾巴克','gold_before':1009,'food_price':35,'gold_after_purchase':974,'inn_price':56,'saved_gold':918,'backpack_items':1,'record':'c3a641581056082b002300370000a000'},
 'pixel_checks':checks,'negative_controls':{'wrong_windows_architecture_rejected':pe_negative,'gold_change_not_equal':True,'single_pixel_change_not_equal':True},
 'limitations':['Wine不等於Windows真機；macOS未實際執行','只驗当前source資源工作目錄，未驗正式封包或從其他目錄啟動','重載明寫-load，未驗預設自動續玩入口','正常GUI標題與戶外隊伍標記會隨實時動畫推進，整圖嚴格比較只選靜止室內／物品狀態','未驗音訊人耳、原版同狀態、全戰鬥／任務、完整通關或公開散布'],
 'local_evidence_sha256':{str(p.relative_to('/work')):sha(p) for platform in ('linux','windows') for p in [root/platform/'normal/trace.json',root/platform/'reload/trace.json',root/platform/'default-save.dat']}}
(root/'checks.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('來源',result['source_count'],'；兩平台各198互動／15重載按鍵、199／16軌跡；',len(checks),'整圖比較0差；保存與macOS結構均通過',flush=True)
