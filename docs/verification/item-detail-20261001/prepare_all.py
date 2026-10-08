"""在容器暫存副本附加正常物品頁輸入與唯讀呼叫觀察。"""
from pathlib import Path
import hashlib,json,os,shutil
out=Path('/out')/os.environ['KOL_ITEM_ATTEMPT']
assert out.parent.stat().st_uid==os.getuid()
assert not out.exists()
out.mkdir()
shutil.copytree('/src/workplace/oracle/game',out/'game')
p=Path('/tmp/kol/tools/oracle/book_startup_test.go')
s=p.read_text()
needle='\t\t\t\t\t\tif os.Getenv("KOL_BOOK_GOODBYE_SINGLE") == "1" {'
assert s.count(needle)==1
s=s.replace(needle,'\t\t\t\t\t\tif os.Getenv("KOL_BOOK_ITEM_PAGE") == "1" { probeNormalItemPage(t,o); return }\n'+needle)
s+=r'''

// 只由正常啟動後的旅店畫面送鍵；只讀呼叫端及像素。
func probeNormalItemPage(t *testing.T,o *oracle.Oracle) {
 o.SyncDGroupFromDS()
 o.WriteU16(o.DS(0x1600),1); o.WriteU16(o.DS(0x1602),0)
 if err:=o.MoveMouse(638,199); err!=nil {t.Fatal(err)}
 if err:=o.Run(20_000_000); err!=nil {t.Fatal(err)}
 base:=uint32(o.DGroupSeg())*16-0x2DED0
 var calls []map[string]any
 for _,ea:=range []uint32{0x16986,0x153FF,0x18857,0x1BA9D,0x1103F} {
  address:=ea; lin:=base+ea
  o.OnCall(oracle.Far(uint16(lin>>4),uint16(lin&15)),func(p *oracle.Oracle){
   if len(calls)>1000 {t.Fatal("物品頁觀察超出界限")}
   var args []uint16;for i:=0;i<6;i++ {args=append(args,p.Arg(i))}
   caller:=p.Caller()
   calls=append(calls,map[string]any{"ida_ea":fmt.Sprintf("0x%X",address),"args":args,
    "caller_ida":fmt.Sprintf("0x%X",uint32(caller.Seg)*16+uint32(caller.Off)-base),"step":p.Steps()})
  })
 }
 keys:=strings.Split(os.Getenv("KOL_ITEM_KEYS"),",")
 captureBook(t,o,"item-entry.png")
 for i,key:=range keys {
  var err error;if len([]rune(key))==1 {err=o.TypeKeys(key)} else {err=o.SendKeys(key)}
  if err!=nil {t.Fatal(err)}
  if err=o.Run(20_000_000);err!=nil {t.Fatal(err)}
  for n:=0;o.KeysPending()>0&&n<20;n++ {if err=o.Run(20_000_000);err!=nil {t.Fatal(err)}}
  if o.KeysPending()>0 {t.Fatalf("第%d鍵%s未被讀走",i,key)}
  captureBook(t,o,fmt.Sprintf("item-%02d.png",i))
  t.Logf("唯讀定位 DS:BB86(word_39A56)=%d 圖示數DS:6C=%d 圖示選取DS:6A=%d 圖形模式DS:4E=%d",o.Word(o.DS(0xBB86)),o.Word(o.DS(0x6C)),o.Word(o.DS(0x6A)),o.Word(o.DS(0x4E)))
  t.Logf("物品頁 step=%d key=%q IP=%s 已讀=%d",i,key,o.IP(),o.KeysConsumed())
 }
 body,err:=json.MarshalIndent(map[string]any{"schema":"kol-item-page-original-v1","seed":1,
  "normal_inputs":true,"state_injection":false,"direct_entry":false,"mouse_api":"MoveMouse(638,199)",
  "address_space":"IDA linear = runtime linear - (DGROUP segment*16 - 0x2DED0)","keys":keys,"calls":calls},"","  ")
 if err!=nil {t.Fatal(err)}
 if err=os.WriteFile(filepath.Join(os.Getenv("KOL_ORACLE_CAPTURE"),"item-calls-private.json"),append(body,'\n'),0600);err!=nil {t.Fatal(err)}
}
'''
p.write_text(s)
shutil.copyfile(p,out/'instrumented-book-startup-private.go')
shutil.copyfile('/scratch/original_all.sh',out/'original.sh')
shutil.copyfile('/scratch/prepare_all.py',out/'prepare.py')
inputs=[]
for root in ['/src/internal','/src/tools/oracle','/src/workplace/oracle/game','/tmp/dosgolem/internal','/tmp/dosgolem/oracle']:
 for f in sorted(Path(root).rglob('*')):
  if f.is_file():inputs.append(dict(path=str(f),sha256=hashlib.sha256(f.read_bytes()).hexdigest()))
for name in ['/src/go.mod','/src/go.sum','/tmp/dosgolem/go.mod','/scratch/original_all.sh','/scratch/prepare_all.py']:
 f=Path(name);inputs.append(dict(path=name,sha256=hashlib.sha256(f.read_bytes()).hexdigest()))
(out/'inputs.json').write_text(json.dumps(dict(schema='kol-item-page-inputs-v1',sources=inputs,
 keys=os.environ['KOL_ITEM_KEYS'],seed=1,revision='968ca774705853276ae55c8d1163c369481ac71f',
 normal_startup=True,state_injection=False,direct_entry=False),ensure_ascii=False,indent=2)+'\n')
