set -eu
export PATH=/usr/local/go/bin:$PATH GOWORK=off GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
export GOMODCACHE=/tmp/mod GOCACHE=/cache GOPATH=/tmp/gopath
cp -a /go/pkg/mod /tmp/mod
python3 - <<'PY'
from pathlib import Path
import json,hashlib,shutil
import re
r,w=Path('/src'),Path('/out'); lab=Path('/tmp/release-hd-scenes'); lab.mkdir()
expected={str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest() for d in ("cmd","internal") for p in sorted((r/d).rglob("*.go"))}
expected.update({n:hashlib.sha256((r/n).read_bytes()).hexdigest() for n in ("go.mod","go.sum")})
for d in ['cmd','internal']:shutil.copytree(r/d,lab/d)
for p in ['go.mod','go.sum']:shutil.copyfile(r/p,lab/p)
excluded=[]
for p in (lab/'cmd/kol').glob('*_test.go'):
    excluded.append(str(p.relative_to(lab)));p.unlink()
for p in ['assets','docs','tools']:(lab/p).symlink_to(r/p,target_is_directory=True)
fixtures={
 'docs/verification/hd-town-crests-runtime-20261015/gallery.go.txt':'cmd/kol/crest_preview_test.go',
 'docs/verification/hd-all-maps-runtime-20261017/maps-gallery.go.txt':'cmd/kol/maps_preview_test.go',
 'docs/verification/hd-scene-events-runtime-20261017/gui-gallery.go.txt':'cmd/kol/events_preview_test.go',
 'docs/verification/hd-full-runtime-20261017/native-monster-gallery.go.txt':'cmd/kol/motion_preview_test.go',
 'docs/verification/hd-scene-events-runtime-20261017/ui-verification.go.txt':'internal/ui/hd_events_verification.go',
 'docs/verification/hd-scene-events-runtime-20261017/battle-verification.go.txt':'internal/battle/hd_events_verification.go',
}
for source,dest in fixtures.items():shutil.copyfile(r/source,lab/dest)
(lab/'internal/ui/hd_motion_verification.go').write_text('package ui\nfunc(b *BattleTiles)PrepareMotionRingVerification(p int){b.ringPhase=p}\n')
p=lab/'cmd/kol/motion_preview_test.go';s=p.read_text()
needle='\treturn m, cases\n}'
assert s.count(needle)==1
selection=['weather-4-01','body-1-1','battle-08-map-16','book-cover','book-open-018','title-02-SET2PICS-05','credit-13','long-inn','monster-native-encounter','event-world-05','event-combat-03','map-world','map-town3']
added='''
	add("map-world", "world-map", true, func(m *game) error {
		m.scene = sceneWorld
		m.worldState = fullWorld(t, m, 23, 29)
		return nil
	})
	add("map-town3", "town-map", true, func(m *game) error {
		m.scene = sceneOutdoor
		m.outdoor = fullMapOutdoor(t, m, 3, 23, 29)
		return nil
	})
	var selected []motionCase
	for _, name := range []string{NAMES} {
		for _, c := range cases { if c.Name == name { selected = append(selected, c) } }
	}
	if len(selected) != 13 { t.Fatalf("missing release HD cases: %d", len(selected)) }
	return m, selected
}'''.replace('NAMES',','.join(json.dumps(x) for x in selection))
p.write_text(s.replace(needle,added))
s=p.read_text().replace("\tvar words []string","\tif m.hd != nil { m.hd.BeginHDFrame() }\n\tvar words []string",1).replace("\tafter, e := json.Marshal(c.State(m))","\tif m.hd != nil { if err := m.hd.EndHDFrame(); err != nil { g.err=err; return } }\n\tafter, e := json.Marshal(c.State(m))",1)
p.write_text(s)

helpers=['crestSetup','fullMapOutdoor','fullWorld','eventSetup','motionSetup']
for p in (lab/'cmd/kol').glob('*_preview_test.go'):
    s=p.read_text()
    for name in helpers:
        s=re.sub(r'(func '+name+r'\(t )\*testing\.T',r'\1releaseCaptureTest',s)
    p.write_text(s)
(lab/'cmd/kol/release_driver_test.go').write_text('''package main
import("encoding/json";"fmt";"os";"testing";"github.com/hajimehoshi/ebiten/v2")
type releaseCaptureTest interface { Helper(); Fatal(...any); Fatalf(string,...any) }
type releaseCaptureValidator struct{}
func(releaseCaptureValidator)Helper(){}
func(releaseCaptureValidator)Fatal(a ...any){panic(fmt.Sprint(a...))}
func(releaseCaptureValidator)Fatalf(f string,a ...any){panic(fmt.Sprintf(f,a...))}
func TestMain(tests *testing.M){
 if os.Getenv("KOL_RELEASE_MAIN")!="1"{os.Exit(tests.Run())}
 buildVersion="v.1.0.4-20261017"
 m,cases:=motionSetup(releaseCaptureValidator{})
 g:=&motionGallery{m:m,cases:cases}
 ebiten.SetWindowSize(960,600)
 if e:=ebiten.RunGame(g);e!=nil{panic(e)}
 if g.err!=nil{panic(g.err)}
 if !g.done||len(g.frames)!=52{panic("incomplete controlled capture")}
 raw,e:=json.MarshalIndent(map[string]any{"version":buildVersion,"scope":"release-main-thread","scenes":len(cases),"frames":g.frames,"controlled":true,"original_oracle":false,"normal_player_path":false,"remake_title_seed":1,"original_seed":nil},"","  ")
 if e!=nil{panic(e)}
 if e=os.WriteFile("/out/motion-captures.json",append(raw,'\\n'),0644);e!=nil{panic(e)}
 fmt.Printf("CONTROLLED CAPTURE PASS: %d scenes, %d frames, main thread\\n",len(cases),len(g.frames))
 os.Exit(0)
}
''')
(w/'source-inputs.json').write_text(json.dumps(expected,indent=2)+'\n')
(w/'fixture-inputs.json').write_text(json.dumps({'sources':fixtures,'excluded_cmd_tests':excluded,'selection':selection,'production_unchanged':True,'controlled':True,'original_oracle':False},indent=2)+'\n')
(w/'motion-runtime').mkdir();(w/'bin').mkdir()
shutil.copy2(lab/'cmd/kol/motion_preview_test.go',w/'motion-gallery.go.txt')
PY
cd /tmp/release-hd-scenes
gofmt -w cmd/kol/*_preview_test.go cmd/kol/release_driver_test.go internal/ui/hd_motion_verification.go
cp cmd/kol/release_driver_test.go /out/release-driver.go.txt
go test -c -o /out/bin/hd-scenes ./cmd/kol >/out/build-linux.log 2>&1
env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o /out/bin/hd-scenes.exe ./cmd/kol >/out/build-windows.log 2>&1
Xvfb :94 -screen 0 2048x1536x24 -nolisten tcp >/out/xvfb-linux.log 2>&1 & xp=$!
trap 'kill "$xp" 2>/dev/null || true; wait "$xp" 2>/dev/null || true; cat /sys/fs/cgroup/memory.peak >/out/memory-peak.txt' EXIT INT TERM
export DISPLAY=:94 LIBGL_ALWAYS_SOFTWARE=1
KOL_RELEASE_MAIN=1 timeout 600 /out/bin/hd-scenes >/out/run-linux.log 2>&1
tail -n 18 /out/run-linux.log
