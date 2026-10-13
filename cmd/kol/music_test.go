package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/flow"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

func TestAcceptedMusicCatalogAndLoop(t *testing.T) {
	path := "../../assets/audio/hd/manifest.json"
	pcm, err := loadMusicCatalog(path)
	if err != nil {
		t.Fatal(err)
	}
	for id, seconds := range map[string]int{"title": 72, "inn": 80, "exploration": 90, "combat": 60} {
		if len(pcm[id]) != seconds*44100*8 {
			t.Fatal("PCM實際長度錯誤")
		}
		// 從真實解碼PCM抵達循環接點，跨兩輪讀取，驗reader沒有截短、補零或換相位。
		p := pcm[id]
		r := audio.NewInfiniteLoopF32(bytes.NewReader(p), int64(len(p)))
		buffer := make([]byte, 8192)
		total := 0
		for total < len(p)*2+8192 {
			n, e := r.Read(buffer)
			if e != nil {
				t.Fatal(e)
			}
			if n == 0 {
				t.Fatal("循環停止")
			}
			for i, b := range buffer[:n] {
				if b != p[(total+i)%len(p)] {
					t.Fatal("循環PCM不連續")
				}
			}
			total += n
		}
	}
}

func TestMusicCatalogRejectsPartialOrUnapproved(t *testing.T) {
	root := "../../assets/audio/hd"
	raw, e := os.ReadFile(filepath.Join(root, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	for _, bad := range []string{"unheard", "escape", "duplicate", "corrupt", "mono", "wrong-duration", "missing-second"} {
		t.Run(bad, func(t *testing.T) {
			dir := t.TempDir()
			var cat musicCatalog
			if e := json.Unmarshal(raw, &cat); e != nil {
				t.Fatal(e)
			}
			for _, track := range cat.Tracks {
				data, e := os.ReadFile(filepath.Join(root, track.File))
				if e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(filepath.Join(dir, track.File), data, 0600); e != nil {
					t.Fatal(e)
				}
			}
			switch bad {
			case "unheard":
				cat.Tracks[0].HumanListening = "待聆聽"
			case "escape":
				cat.Tracks[0].File = "../title-chamber.ogg"
			case "duplicate":
				cat.Tracks[1] = cat.Tracks[0]
			case "corrupt":
				if e = os.WriteFile(filepath.Join(dir, cat.Tracks[0].File), []byte("bad"), 0600); e != nil {
					t.Fatal(e)
				}
			case "missing-second":
				if e = os.Remove(filepath.Join(dir, cat.Tracks[1].File)); e != nil {
					t.Fatal(e)
				}
			default:
				name := "../../assets/audio/title-thunder.ogg"
				if bad == "wrong-duration" {
					name = filepath.Join(root, "inn-chamber.ogg")
				}
				data, e := os.ReadFile(name)
				if e != nil {
					t.Fatal(e)
				}
				sum := sha256.Sum256(data)
				cat.Tracks[0].SHA256 = hex.EncodeToString(sum[:])
				if e = os.WriteFile(filepath.Join(dir, cat.Tracks[0].File), data, 0600); e != nil {
					t.Fatal(e)
				}
			}
			data, e := json.Marshal(cat)
			if e != nil {
				t.Fatal(e)
			}
			p := filepath.Join(dir, "manifest.json")
			if e = os.WriteFile(p, data, 0600); e != nil {
				t.Fatal(e)
			}
			if got, e := loadMusicCatalog(p); e == nil || got != nil {
				t.Fatal("拒絕整包，不可回傳部分PCM")
			}
		})
	}
}

type fakeMusicPlayer struct {
	plays, rewinds, pauses, closes int
	playing                        bool
	volume                         float64
	position                       time.Duration
}

func (p *fakeMusicPlayer) Play()                   { p.plays++; p.playing = true }
func (p *fakeMusicPlayer) Pause()                  { p.pauses++; p.playing = false }
func (p *fakeMusicPlayer) IsPlaying() bool         { return p.playing }
func (p *fakeMusicPlayer) Rewind() error           { p.rewinds++; p.position = 0; return nil }
func (p *fakeMusicPlayer) SetVolume(v float64)     { p.volume = v }
func (p *fakeMusicPlayer) Close() error            { p.closes++; return nil }
func (p *fakeMusicPlayer) Position() time.Duration { return p.position }
func TestMusicSceneFadeAndIndependentMute(t *testing.T) {
	a, b := &fakeMusicPlayer{}, &fakeMusicPlayer{}
	s := &musicSound{players: map[string]musicPlayer{"title": a, "inn": b}, gain: map[string]float64{}}
	for i := 0; i < 30; i++ {
		if e := s.step("title", true); e != nil {
			t.Fatal(e)
		}
	}
	a.position = 10 * time.Second
	if a.plays != 1 || a.rewinds != 1 || a.volume != 1 || b.playing {
		t.Fatal("相同場景不可每幀重播")
	}
	for i := 0; i < 30; i++ {
		s.step("title", false)
	}
	if a.playing || a.volume != 0 {
		t.Fatal("F9關閉未淡出")
	}
	s.step("title", true)
	if a.position != 10*time.Second || a.rewinds != 1 {
		t.Fatal("F9重新開啟應從原位置續播")
	}
	for i := 0; i < 30; i++ {
		s.step("inn", true)
	}
	if a.playing || !b.playing || b.volume != 1 || b.rewinds != 1 {
		t.Fatal("場景轉換未切換曲目")
	}
	for i := 0; i < 30; i++ {
		s.step("", true)
	}
	if a.playing || b.playing {
		t.Fatal("無配樂場景未停止")
	}
	s.close()
	if a.closes != 1 || b.closes != 1 {
		t.Fatal("離開未釋放播放器")
	}
}
func TestMusicFollowsActualPlayerScene(t *testing.T) {
	inn := &town.Inn{Title: []string{"半途之家"}}
	inn.Portrait.Archive, inn.Portrait.Index = "TWN7PICS", 12
	active := &town.Session{Mode: town.ModeInside, Inside: inn}
	g := game{scene: sceneFlow, play: &town.Session{}, flow: &flow.Flow{Scene: flow.SceneTown, Town: active}}
	if got := g.musicCue(); got != "inn" {
		t.Fatal("應使用目前flow.Town場景", got)
	}
	g.flow.Scene = flow.SceneTitle
	if g.musicCue() != "title" {
		t.Fatal("標題曲未識別")
	}
	g.flow.Scene = flow.SceneContents
	if g.musicCue() != "" {
		t.Fatal("書本不借用旅店曲")
	}
	g.scene = scenePlay
	g.play = active
	g.play.LeftTown = true
	if g.musicCue() != "exploration" {
		t.Fatal("離城探索曲未識別")
	}
	g.quest = &questBattle{state: &battle.State{}}
	for _, stage := range []battle.Stage{battle.StageEncounter, battle.StageField, battle.StageFlee} {
		g.quest.state.Stage = stage
		if g.musicCue() != "combat" {
			t.Fatal("戰鬥曲未識別", stage)
		}
	}
	for _, stage := range []battle.Stage{battle.StageIntro, battle.StageSpoils, battle.StageDone} {
		g.quest.state.Stage = stage
		if g.musicCue() != "" {
			t.Fatal("任務說明與戰後不可播放戰鬥曲", stage)
		}
	}
	g.quest = nil
	g.endingArmed = true
	if g.musicCue() != "" {
		t.Fatal("結局不借用旅店曲")
	}
	g.interactive = false
	g.musicPath = t.TempDir()
	if e := g.setMusic(true); e != nil || g.music != nil {
		t.Fatal("批次模式不得載音訊")
	}
}
