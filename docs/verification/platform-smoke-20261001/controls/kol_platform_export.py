from pathlib import Path
import json, shutil, hashlib, os, re
repo=Path('/work'); root=repo/'workplace/platform-smoke-20261001'
out=repo/'docs/verification/platform-smoke-20261001'
def sha(p): return hashlib.sha256(Path(p).read_bytes()).hexdigest()
def load(p): return json.loads(Path(p).read_text())
for p in (repo/'docs/verification',repo/'CONTEXT.md',repo/'WORKLOG.md',repo/'docs/worklist.json',repo/'WORKLIST.md',repo/'docs/verification/manual-fidelity-audit-20260923.md'):
    assert (p.stat().st_uid,p.stat().st_gid)==(os.getuid(),os.getgid()),p
assert not out.exists()
checks=load(root/'checks.json')
assert len(checks['pixel_checks'])==40 and checks['source_count']==167
out.mkdir()
files={}
def take(src,dst):
    target=out/dst;target.parent.mkdir(parents=True,exist_ok=True)
    shutil.copyfile(src,target);files[dst]=sha(target)
take(root/'checks.json','checks.json')
take(root/'inputs.json','inputs.json')
take(root/'purchase-script.txt','purchase-script.txt')
take(root/'reload-script.txt','reload-script.txt')
take(root/'build/build.log','build.log')
for p in (root/'macos').glob('*.txt'): take(p,'macos/'+p.name)
take(root/'macos/build.log','macos/build.log')
controls=['kol_platform_prepare.py','kol_platform_build.sh','kol_platform_macos_build.sh','kol_platform_windows_cover.sh','kol_platform_windows_normal.sh','kol_platform_linux_normal.sh','kol_platform_gui_control.py','kol_platform_gui_engine.sh','kol_platform_gui_reload.sh','kol_platform_verify.py','kol_platform_export.py']
for name in controls:
    take(Path('/control')/name,'controls/'+name)
for p in (root/'controls-attempt1').glob('*'): take(p,'controls-attempt1/'+p.name)
review=[];frames=[]
for platform,gui in [('windows','gui-normal'),('linux','gui-normal2')]:
    take(root/platform/'cover.png',platform+'/cover.png')
    for kind,local in [('normal',gui),('reload','gui-reload')]:
        for p in (root/platform/local).glob('*.png'):
            dst=platform+'/'+kind+'/'+p.name;take(p,dst)
            frames.append({'file':dst,'frame':int(re.search(r'\d+',p.stem)[0]),'platform':platform,'input_method':'X11 XTEST 真實鍵盤；無座標／資料注入','generator':'Go1.26.7／Ebitengine2.9.9，X11 GetImage 深度24擷取','png_sha256':files[dst],'evidence':'已證實此重製版所列流程；未宣稱原版parity','rights':'僅private GitHub PNG特例'})
        for name in ('inputs.json','engine.log'):
            take(root/platform/local/name,platform+'/'+kind+'/'+name)
    take(root/platform/'version.txt',platform+'/version.txt')
    for name in ('cover.log','save-before-reload.sha256'):
        take(root/platform/name,platform+'/'+name)
prior=load(repo/'docs/verification/cjk-ui/longest-equipment-20261001/receipt.json')
source=prior['source_assets']
for p,h in source['png_manifests_sha256'].items(): assert sha(repo/p)==h
manifest=load(repo/'assets/original-derived-png/ega/manifest.json')
source['archives']=[{'name':a['archive'],'zip_entry_root':a['archive_root'],'extractor':manifest['tool'],'manifest':'assets/original-derived-png/ega/manifest.json'} for a in manifest['archives'] if a['archive'] in ('BOOKPICS','BACKPICS','CHARPICS','KOLMPICS','TWN7PICS','TOWN7','PERPICS','HANDPICS','WEAPPICS','SHLDPICS')]
source['generator']='Go1.26.7／Ebitengine2.9.9；Windows Wine9／Linux Xvfb正常互動與最小封面擷取'
for frame in frames:
    frame['source_file']=source['source_file'];frame['source_sha256']=source['source_sha256']
    frame['source_entry_index']='receipt.json/source_assets/archives；多來源合成畫格，逐圖來源見所列固定SHA manifest'
visual=['windows/cover.png','windows/normal/step-002.png','windows/normal/step-116.png','windows/normal/step-137.png','windows/normal/step-198.png','windows/reload/step-015.png']
receipt={'schema':'kol-platform-smoke-receipt-v1','date':'2026-10-01','status':checks['status'],'scope':'目前未提交工作樹的中間建置；不建立正式版號／dist-all封包',
 'checks':'checks.json','inputs':'inputs.json','files':files,'png_frames':frames,'source_assets':source,
 'cover_frames':[{'file':p+'/cover.png','frame':'cover','png_sha256':files[p+'/cover.png'],'source_file':source['source_file'],'source_sha256':source['source_sha256'],'source_entry_root':'Knights-of-legend/BOOKPICS','generator':'Go1.26.7／Ebitengine2.9.9 -scene cover -lang zh-Hant','evidence':'已證實重製封面；未宣稱原版parity','rights':'僅private GitHub PNG特例'} for p in ('linux','windows')],
 'visual_review':{'method':'主代理實際檢視列示Windows完整畫格，中文物品與住宿文字可讀；Linux所列靜止畫格逐像素相同','reviewed':visual,'linux_same_static_pixels':True},
 'toolchain':{'go_ebiten':'Go1.26.7／Ebitengine2.9.9','go_image':'sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa','package_image':'sha256:3d32441befe02810c3220a2e9f2db5b84ee65c703e5b65ee563712a3519ad193','wine_image':'sha256:4e41c4dca4d6a4450945fe13d0ebc198d07d36090fac87929d7f3696f79315ce','macos_image':'sha256:1d7655e745573fb79c1adf0a9e2b304b500dae53b539def23b556ee4733d1268'},
 'execution_controls':{'windows_normal':'controls-attempt1/kol_platform_gui_control.py；controls-attempt1/kol_platform_gui_engine.sh','linux_normal_and_reload':'controls/kol_platform_gui_control.py；controls/kol_platform_gui_engine.sh／kol_platform_gui_reload.sh'},
 'local_evidence_sha256':checks['local_evidence_sha256'],
 'failures_preserved':{'batch_no_default_save':'兩側normal/trace.json共199筆；批次旗標排除互動預設保存，產品未失敗','linux_first_quit_race':'linux/gui-normal及config保存首輪；Ctrl+Q後X11關閉使捕捉器BrokenPipe，修控制後config2／gui-normal2以相同198鍵乾淨重跑成功','diagnostic_timeout':'一次診斷迴圈反覆解碼同圖而逾時，已停止精確容器；以單次讀取核對同SHA，未改產品'},
 'rights':'PNG僅沿用private GitHub特例；原版ZIP／typed資料／存檔／binary／Wineprefix／SDK不複製到正式文件；不新增公開散布授權',
 'limitations':checks['limitations'],'worklist':'仍23未完成／83完成；#37／#42／#7未結案'}
(out/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
summary='2026-10-01 現行工作樹三平台中間建置續核：Linux／Windows各編譯兩次雜湊一致，Linux仍f969a63f；Windows304890ea在Wine9實際啟動，預設繁中與APPDATA保存、正常標題／六人招募／購買／住宿198個互動按鍵及重載15鍵通過。兩平台199／16筆批次狀態與既有收據完全相同，40組整圖比較0差；保存除saved_at外一致、重載未改檔。macOS universal兩架構、最低11.0、arm64簽章區塊與系統相依核實通過，但尚未實際啟動；Wine不是Windows真機。收據docs/verification/platform-smoke-20261001/receipt.json。只驗當前source資源入口，未建正式封包或版號，本項不結案。'
work=repo/'docs/worklist.json'; raw=work.read_text(); data=json.loads(raw)
for item in data['items']:
    if item.get('issue') in (7,37,42):
        old=item['body'];new=old+' '+summary
        assert raw.count(json.dumps(old,ensure_ascii=False))==1
        raw=raw.replace(json.dumps(old,ensure_ascii=False),json.dumps(new,ensure_ascii=False),1)
        item['body']=new
assert json.loads(raw)==data
work.write_text(raw)
context=repo/'CONTEXT.md'; text=context.read_text(); anchor='\n\n2026-10-01 正常裝備欄續核'
assert anchor in text
paragraph='\n\n2026-10-01 跨平台啟動續核（最新中間建置）：Linux／Windows各兩次編譯SHA一致，\nLinux仍f969a63f，Windows304890ea在Wine9實際啟動。兩側預設繁中、正常六人招募／\n購買／住宿198個互動按鍵、預設設定目錄保存與重載15鍵通過；199／16筆批次状态\n與既有收據相同，40組整圖0差，保存除時間外相同且重載未改寫。\nmacOS universal兩架構、最低11.0、arm64簽章區塊及系統函式庫相依通過，但未執行；\nWine不等於Windows真機。見[正式收據](docs/verification/platform-smoke-20261001/receipt.json)。\n只驗當前source資源工作目錄，未建dist-all或正式版號；#37／#42／#7仍開放，\n本地23未完成／83完成。未改產品、提交／推送、遠端Issue或發布。\n'
context.write_text(text.replace(anchor,paragraph+anchor,1))
audit=repo/'docs/verification/manual-fidelity-audit-20260923.md'
audit.write_text(audit.read_text().rstrip()+'\n\n## 2026-10-01：現行三平台建置與 Windows Wine 正常保存\n\n'+summary.replace('收據docs/verification/platform-smoke-20261001/receipt.json。','見[跨平台收據](platform-smoke-20261001/receipt.json)。')+'\n')
log=repo/'WORKLOG.md'
log.write_text(log.read_text().rstrip()+'''\n\n## 2026-10-01：現行三平台中間建置、Wine 正常互動與預設保存\n\n- 路由命中跨平台試玩、dist-all輸出位置及文件職責；沿用既有Go1.26.7／Ebitengine2.9.9、Wine9、osxcross與Python工具image，未新建image。原始輸入與專案資源唯讀，task輸出與Wineprefix以1000:1000寫入；全部Docker工作有界、--rm、--network none。首個image inspect模板因Config缺Entrypoint鍵失敗，改用index後確認原image；沒有換image或搬入主機runtime。\n- Linux／Windows各編譯兩次逐byte一致；LinuxSHA仍f969a63f，Windows304890ea。macOS universalSHA2b6370c9，兩個slice與thin檔逐byte一致、最低11.0、arm64簽章superblob及所有系統相依通過；只做靜態核實，未宣稱macOS啟動或雙次重建一致。\n- Wine64 prefix初始化成功；wineboot的32位元／OLE警告保留，不誤判64位元產品故障。批次flow首輪兩側199筆正常購買軌跡相同但不寫預設保存，控制腳本test失敗；核實batchFlags後分類入口問題，保留原結果，不改正式程式。\n- 用既有x11_key_inject.py在另一Docker容器連共用唯讀socket，XTEST逐鍵送出並核對log-input收到同token；X11 GetImage深度24擷取。直接互動啟動不指定scene／lang／save，兩側198鍵從標題、六人招募、酒館購買帶走、檢視第一背包格到正常住宿。Windows寫APPDATA、Linux写XDG_CONFIG_HOME的save.dat，六人、gold918與食品record均符合；重啟-load本次自然保存，15互動鍵與16筆批次重載一致、原檔SHA不變。\n- Linux首輪Ctrl+Q快速退出，擁有容器已关X11而捕捉器keyup遇BrokenPipe；198鍵與存檔完整保留，修控制器先記成功、最後keyup允許X11关闭，產品退出碼仍由wait嚴格核實。新config2／gui-normal2用相同198鍵乾淨重跑，退出0；首輪控制快照另存。一次PNG診斷迴圈重複解碼而逾時，停止精確容器，再以單次讀取同SHA核對，屬診斷腳本問題。\n- 167筆sourceSHA執行前後一致；兩側199／16筆批次狀態與既有初始物品收據相同，40組完整靜止畫格0像素差，保存只差saved_at。錯誤PE架構、金冠差異與單像素差異反例拒絕／不相同。主代理檢视Windows封面、目錄、商品、物品、住宿與重載明細，中文可讀，Linux對應靜止畫格一致。\n- 正式收據及36張PNG、控制來源、編譯／静態日誌放docs/verification/platform-smoke-20261001/，同次接入CONTEXT、稽核與#7／#37／#42。原版ZIP、typed payload、binary、SDK、Wineprefix及自然保存留workplace/platform-smoke-20261001；PNG只沿用private GitHub特例。工作樹继承異動不動，不提交／推送、不寫遠端Issue、不建正式版號／dist-all、不發布。三平台整項仍開放；macOS實際啟動、Windows／macOS真機、正式封包、FUSE、音訊及推廣分別待驗。\n'''.replace('写','寫').replace('关','關').replace('闭','閉').replace('视','視').replace('态','態').replace('继','繼'))
print('正式附件',len(files),'PNG',len(frames)+2,'已匯出；三條待辦與索引同步，本輪未結案整項',flush=True)
