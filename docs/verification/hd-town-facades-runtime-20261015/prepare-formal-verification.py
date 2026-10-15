from pathlib import Path
import json
import hashlib

r = Path('/src')
w = Path('/out')
text = (w / 'gallery-full-group-v2.go.txt').read_text()
anchor = 'if g.j==3{restore:=m.hd.SwapFacadePrototype(c.Paint,c.RGB);defer restore()}'
assert text.count(anchor) == 1
text = text.replace(anchor, '')
anchor = 'if g.j==3&&c.Key=="twn5pics-019"{if g.err=m.hd.DrawFacadeSignPrototype(m.outputCanvas().Image());g.err!=nil{return}};'
assert text.count(anchor) == 1
text = text.replace(anchor, '')
anchor = 'path:=row.File;if !filepath.IsAbs(path){path="/out/"+path};file,e:=os.Open(path);must(e);paint,e:=png.Decode(file);must(e);must(file.Close())'
assert text.count(anchor) == 1
text = text.replace(anchor, 'var paint image.Image')
text = text.replace('full-group-v2', 'formal-gallery')
(w / 'formal-gallery.go.txt').write_text(text)

# Reuse the same real typed door preparation in the controlled GUI harness.
start = text.index('root:="/src";')
end = text.index(' ebiten.SetWindowSize', start)
setup = text[start:end]
setup = setup.replace('must(os.Mkdir("/out/formal-gallery-gallery",0755))', '')
header = '''package main
import("encoding/json";"fmt";"image";"os";"testing";"github.com/hajimehoshi/ebiten/v2";"github.com/wicanr2/kol-remake-cht/internal/assets";"github.com/wicanr2/kol-remake-cht/internal/font";"github.com/wicanr2/kol-remake-cht/internal/i18n";"github.com/wicanr2/kol-remake-cht/internal/screen";"github.com/wicanr2/kol-remake-cht/internal/town";"github.com/wicanr2/kol-remake-cht/internal/flow")
type facadeGUIObserver struct{g *game}
func(n *facadeGUIObserver)Update()error{return n.g.Update()}
func(n *facadeGUIObserver)Layout(w,h int)(int,int){return n.g.Layout(w,h)}
func(n *facadeGUIObserver)Draw(dst *ebiten.Image){n.g.Draw(dst);raw,_:=json.Marshal(map[string]any{"Outdoor":n.g.play.Outdoor,"hd":n.g.hd!=nil,"settings":n.g.displaySettings});fmt.Printf("equipment-frame %s\\n",raw)}
func TestHDFacadeGUI(t *testing.T){name:=os.Getenv("KOL_FACADE_GUI_CASE");if name==""{t.Skip("explicit prepared GUI")};must:=func(e error){if e!=nil{t.Fatal(e)}};
'''
footer = '''
var chosen *facadeCase;for i:=range g.cases{if g.cases[i].Key==name{chosen=&g.cases[i];break}};if chosen==nil{t.Fatal("unknown facade GUI case")}
m.play=&town.Session{Outdoor:chosen.State,Party:chosen.State.Party};m.flow=&flow.Flow{Scene:flow.SceneTown,Town:m.play};m.scene=sceneFlow;m.interactive=true;m.logInput=true
// Controlled static marker phase only. Production animation and rules are unchanged.
m.walkTick=-1073741824
ebiten.SetWindowSize(1920,1200);ebiten.SetWindowTitle("七城門面HD遊戲內切換驗收");ebiten.SetTPS(60)
must(ebiten.RunGame(&facadeGUIObserver{m}));must(m.drawErr)
}
'''
(w / 'formal-gui.go.txt').write_text(header + setup + footer)
old = r / 'docs/verification/hd-equipment-runtime-20261015'
normal = (old / 'normal.py').read_text().replace("'formal-actions'", "'baseline'").replace("'equipment-formal'", "'kol-formal-town-facades'")
(w / 'normal-facades.py').write_text(normal)
save = (old / 'save.py').read_text().replace("out=w/'equipment-formal-save'", "out=w/'facade-formal-save'").replace("'formal-actions'", "'baseline'").replace("'equipment-formal'", "'kol-formal-town-facades'")
(w / 'save-facades.py').write_text(save)
prefix = (old / 'gui.py').read_text().split('xlog = ')[0]
prefix = prefix.replace("DEST = OUT / os.environ.get('KOL_GUI_OUTPUT', 'gui-prototype')", "DEST = OUT / ('gui-' + os.environ['KOL_FACADE_GUI_CASE'])")
source = (r / 'docs/verification/hd-npc-mounts-runtime-20261015/prepare-regression.py').read_text()
a = source.index("footer='''") + len("footer='''")
b = source.index("'''\n(w/'gui-npc.py')", a)
footer = source[a:b].replace('\\\\n', '\\n').replace('formal-npc-gui', 'formal-facade-gui').replace('TestHDNPCGUI', 'TestHDFacadeGUI').replace('KOL_NPC_GUI_CASE', 'KOL_FACADE_GUI_CASE').replace('full_Inn_unchanged', 'serializable_Outdoor_unchanged').replace('complete Inn preserved', 'serializable Outdoor preserved')
(w / 'gui-facades.py').write_text(prefix + footer)
(w / 'regression-tool-sources.json').write_text(json.dumps({str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest() for p in [old/'normal.py',old/'save.py',old/'gui.py',r/'docs/verification/hd-npc-mounts-runtime-20261015/prepare-regression.py']},indent=2)+'\n')
print('Prepared formal gallery and real F10 GUI, plus existing normal/v3 regression tools', flush=True)
