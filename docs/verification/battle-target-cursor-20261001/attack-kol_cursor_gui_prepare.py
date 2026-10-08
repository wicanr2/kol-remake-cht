from pathlib import Path
import json,hashlib,shutil,os
out=Path('/out')/os.environ.get('KOL_CURSOR_ATTEMPT','gui1')
assert not out.exists()
assert out.parent.stat().st_uid==os.getuid()
out.mkdir()
shutil.copytree('/src/workplace/oracle/game',out/'game')
p=Path('/tmp/kol/tools/oracle/book_startup_test.go');s=p.read_text()
marker='func probeNormalPlanning(t *testing.T, o *oracle.Oracle) {\n'
assert s.count(marker)==1
addition=r'''

 if err := o.MoveMouse(620,4); err != nil { t.Fatal(err) }
 if err := o.Run(20_000_000); err != nil { t.Fatal(err) }
 mouseX, mouseY := o.Mouse()
 t.Logf("圖示列觀察正常滑鼠移動：x=%d y=%d", mouseX, mouseY)
 // 只讀取呼叫參數與螢幕；不替換原版常式、不寫記憶體。
 base := uint32(o.DGroupSeg())*16 - 0x2DED0
 at := func(ea uint32) oracle.Addr { lin := base + ea; return oracle.Far(uint16(lin>>4), uint16(lin&15)) }
 var calls []map[string]any
 for _, ea := range []uint32{0x1C683, 0x1C601, 0x1C40D, 0x1CB28, 0x1D21D, 0x1C365, 0x1A8FD, 0x1EF3A} {
   address := ea
   o.OnCall(at(address), func(p *oracle.Oracle) {
     if len(calls)>=5000 { t.Fatal("呼叫紀錄超出界限") }
     var args []uint16
     for i:=0; i<6; i++ { args=append(args,p.Arg(i)) }
     caller := p.Caller()
     calls=append(calls,map[string]any{"ida_ea":fmt.Sprintf("0x%X",address), "caller_runtime":caller.String(),
       "caller_ida":fmt.Sprintf("0x%X",uint32(caller.Seg)*16+uint32(caller.Off)-base),"args":args,
       "step":p.Steps(),"cursor_row":p.Word(p.DS(0x68)),"cursor_col":p.Word(p.DS(0x66)),
       "cursor_mode":p.Word(p.DS(0x78)),"cursor_enabled":p.Word(p.DS(0x9E)),
       "cursor_x":p.Word(p.DS(0xBB80)),"cursor_y":p.Word(p.DS(0xCD8C)),
       "animation_phase":p.Word(p.DS(0x54)),"animation_direction":p.Word(p.DS(0x56)),"animation_countdown":p.Word(p.DS(0x58))})
   })
 }
 defer func() {
   payload:=map[string]any{"schema":"kol-target-cursor-calls-v1","tool":"dosgolem","revision":"968ca774705853276ae55c8d1163c369481ac71f",
     "address_space":"IDA linear = runtime linear - (DGROUP segment*16 - 0x2DED0)", "seed_before_travel":1,"writes_or_stubs":false,"calls":calls}
   body,err:=json.MarshalIndent(payload,"","  ");if err!=nil {t.Fatal(err)}
   if err=os.WriteFile(filepath.Join(os.Getenv("KOL_ORACLE_CAPTURE"),"cursor-calls.json"),append(body,'\n'),0600);err!=nil {t.Fatal(err)}
 }()
 captureBook(t,o,"plan-entry.png")
'''
s=s.replace(marker,marker+addition)
# 名冊翻頁先完成繪製再送方向鍵；gui8 的兩鍵曾在繪製途中被讀走而未換人。
selection='\t\t\t\t\t\tfor n := 0; n < down; n++ {'
assert s.count(selection)==1
s=s.replace(selection,'\t\t\t\t\t\tif err = o.Run(60_000_000); err != nil { t.Fatal(err) }\n'+selection)
capture='func captureBook(t *testing.T, o *oracle.Oracle, name string) {\n'
meta=r'''
 if strings.HasPrefix(name,"plan-") {
   var actors []string
   for slot:=0;slot<18;slot++ {actors=append(actors,hex.EncodeToString(o.Bytes(o.DS(uint16(0x6712+slot*50)),50)))}
   payload:=map[string]any{"schema":"kol-target-frame-state-v1","frame":name,"step":o.Steps(),
     "dgroup_segment":o.DGroupSeg(),"cursor_row":o.Word(o.DS(0x68)),"cursor_col":o.Word(o.DS(0x66)),
     "cursor_mode":o.Word(o.DS(0x78)),"animation_disabled":o.Word(o.DS(0x90)),"acting_actor":o.Word(o.DS(0xB48)),"last_command":o.Word(o.DS(0x3252)),"phase_next":o.Word(o.DS(0x54)),"direction_next":o.Word(o.DS(0x56)),"icon_slots":o.Word(o.DS(0x6C)),"summary_flag":o.Word(o.DS(0xB4A)),"icon_selected":o.Word(o.DS(0x6A)),
     "actor_records":actors,"visible_tiles":hex.EncodeToString(o.Bytes(o.DS(0x2A48),81)),
     "visible_sprites":hex.EncodeToString(o.Bytes(o.DS(0x6AE4),81)),"view_actor_word":o.Word(o.DS(0x9E5A))}
   b,err:=json.MarshalIndent(payload,"","  ");if err!=nil {t.Fatal(err)}
   if err=os.WriteFile(filepath.Join(os.Getenv("KOL_ORACLE_CAPTURE"),name+".state.json"),append(b,'\n'),0600);err!=nil {t.Fatal(err)}
 }
'''
assert s.count(capture)==1
s=s.replace(capture,capture+meta);p.write_text(s)
shutil.copyfile(p,out/'instrumented-book-startup-private.go')
records=[]
for root in ['/src/internal','/src/tools/oracle','/src/workplace/oracle/game','/tmp/dosgolem/internal','/tmp/dosgolem/oracle']:
 for f in sorted(Path(root).rglob('*')):
  if f.is_file():records.append({'path':str(f),'sha256':hashlib.sha256(f.read_bytes()).hexdigest()})
for name in ['/src/go.mod','/src/go.sum','/tmp/dosgolem/go.mod','/scratch/attack-kol_cursor_gui.sh','/scratch/attack-kol_cursor_gui_prepare.py','/scratch/attack-kol_cursor_gui_finish.py']:
 f=Path(name);records.append({'path':name,'sha256':hashlib.sha256(f.read_bytes()).hexdigest()})
(out/'inputs.json').write_text(json.dumps({'schema':'kol-target-cursor-gui-inputs-v1','sources':records,'seed':1,'planning_keys':os.environ.get('KOL_CURSOR_KEYS'),'initial_roster_down':int(os.environ.get('KOL_BOOK_SELECT_DOWN','0')),
 'mouse_event':{'api':'dosgolem.Oracle.MoveMouse','x':620,'y':4,'buttons':0},'normal_inputs':True,'direct_entry':False,'state_injection':False,'rng_control':'出城前固定 DS:1600=1、DS:1602=0；不重擲、不掃種子',
 'image':'kol-package:20260907','dosgolem_revision':'968ca774705853276ae55c8d1163c369481ac71f'},ensure_ascii=False,indent=2)+'\n')
