from pathlib import Path
import json,hashlib
r=Path('/src');w=Path('/out');old=r/'docs/verification/hd-equipment-runtime-20261015'
text=(old/'save.py').read_text().replace("out=w/'equipment-formal-save'","out=w/'monster-formal-save'").replace("'formal-actions'","'baseline'").replace("'equipment-formal'","'formal-monsters'")
(w/'save-monsters.py').write_text(text)
text=(old/'normal.py').read_text().replace("'formal-actions'","'baseline'").replace("'equipment-formal'","'formal-monsters'").replace("'player-pages-v2'","'monster-normal'").replace("'page-prototype-results.json'","'monster-normal-results.json'")
(w/'normal-monsters.py').write_text(text)
prefix=(old/'gui.py').read_text().split('xlog = ')[0]
prefix=prefix.replace("DEST = OUT / os.environ.get('KOL_GUI_OUTPUT', 'gui-prototype')","DEST = OUT / ('gui-' + os.environ['KOL_MONSTER_GUI_CASE'])")
footer=(r/'docs/verification/hd-npc-mounts-runtime-20261015/prepare-regression.py').read_text()
a=footer.index("footer='''")+len("footer='''");b=footer.index("'''\n(w/'gui-npc.py')",a)
footer=footer[a:b].replace("\\\\n","\\n").replace('formal-npc-gui','formal-monster-gui-v3').replace('TestHDNPCGUI','TestHDMonsterGUI').replace('KOL_NPC_GUI_CASE','KOL_MONSTER_GUI_CASE').replace('full_Inn_unchanged','serializable_battle_state_unchanged').replace('complete Inn preserved','serializable battle state preserved')
(w/'gui-monsters.py').write_text(prefix+footer)
# Same typed controlled fixtures as the gallery; only two representative GUI cases.
gallery=(w/'monster-gallery_test.go.txt').read_text()
a=gallery.index(' root:=');b=gallery.index('\n ebiten.SetWindowSize',a)
setup=gallery[a:b]
setup=setup.replace('dir:="/out/first-four-gallery-v2";if e:=os.Mkdir(dir,0755);e!=nil{t.Fatal(e)}','dir:=""')
setup=setup.replace('useMonsterPrototype=i==1','')
setup=setup.replace('play:&town.Session{}','play:&town.Session{Outdoor:&town.Outdoor{}}')
header='''package main
import("encoding/json";"fmt";"image/png";"os";"testing";"github.com/hajimehoshi/ebiten/v2";"github.com/wicanr2/kol-remake-cht/internal/assets";"github.com/wicanr2/kol-remake-cht/internal/battle";"github.com/wicanr2/kol-remake-cht/internal/font";"github.com/wicanr2/kol-remake-cht/internal/flow";"github.com/wicanr2/kol-remake-cht/internal/i18n";"github.com/wicanr2/kol-remake-cht/internal/screen";"github.com/wicanr2/kol-remake-cht/internal/town";"github.com/wicanr2/kol-remake-cht/internal/ui")
type monsterGUIObserver struct{g *game}
func(n *monsterGUIObserver)Update()error{return n.g.Update()}
func(n *monsterGUIObserver)Layout(w,h int)(int,int){return n.g.Layout(w,h)}
func(n *monsterGUIObserver)Draw(dst *ebiten.Image){n.g.Draw(dst);raw,_:=json.Marshal(map[string]any{"battle":n.g.quest.state,"hd":n.g.hd!=nil,"settings":n.g.displaySettings});fmt.Printf("equipment-frame %s\\n",raw)}
func TestHDMonsterGUI(t *testing.T){name:=os.Getenv("KOL_MONSTER_GUI_CASE");if name==""{t.Skip("explicit prepared GUI")}
'''
# Reuse gallery struct/case declarations through its test-only file, not production.
setup=setup.replace('index<4','index<32')
ending='''
g:=gallery.models[1];var chosen *monsterGalleryCase
for i:=range gallery.cases {c:=&gallery.cases[i];if name=="info"&&c.Index==10&&c.Kind=="info"||name=="encounter"&&c.Index==31&&c.Count==12{chosen=c;break}}
if chosen==nil{t.Fatal("unknown GUI case")};g.quest.state=chosen.State;g.quest.lines=chosen.Lines
g.scene=sceneFlow;g.flow=&flow.Flow{Scene:flow.SceneTown,Town:g.play};g.interactive=true;g.logInput=true
ebiten.SetWindowSize(1920,1200);ebiten.SetWindowTitle("怪物HD遊戲內切換驗收");ebiten.SetTPS(60)
must(ebiten.RunGame(&monsterGUIObserver{g}));must(g.drawErr)
}
'''
(w/'monster-gui_test.go.txt').write_text(header+setup+ending)
(w/'regression-source-inputs.json').write_text(json.dumps({str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest()for p in [old/'normal.py',old/'save.py',old/'gui.py',r/'docs/verification/hd-npc-mounts-runtime-20261015/prepare-regression.py']},indent=2)+'\n')
print('Prepared normal, save/reload and two actual F10 GUI fixtures')
