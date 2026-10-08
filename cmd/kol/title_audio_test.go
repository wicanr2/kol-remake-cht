package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/title"
)

func TestTitleCueCatalog(t *testing.T) {
	manifest := "../../assets/audio/manifest.json"
	pcm, err := loadTitleCue(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(pcm) != 42400*8 {
		t.Fatalf("解碼樣本數 %d，預期42400", len(pcm)/8)
	}
	raw, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := os.ReadFile("../../assets/audio/title-thunder.ogg")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		change  func(*titleCueCatalog)
		data    []byte
		missing bool
	}{
		{name: "未聆聽", change: func(c *titleCueCatalog) { c.Cues[0].HumanListening = "待聆聽" }},
		{name: "路徑逃逸", change: func(c *titleCueCatalog) { c.Cues[0].File = "../title-thunder.ogg" }},
		{name: "未知格式", change: func(c *titleCueCatalog) { c.Schema = "v2" }},
		{name: "循環", change: func(c *titleCueCatalog) { c.Cues[0].Loop = true }},
		{name: "取樣率", change: func(c *titleCueCatalog) { c.Cues[0].SampleRate = 22050 }},
		{name: "聲道", change: func(c *titleCueCatalog) { c.Cues[0].Channels = 2 }},
		{name: "時長", change: func(c *titleCueCatalog) { c.Cues[0].Duration = 1.5 }},
		{name: "雜湊不符", data: []byte("damaged ogg")},
		{name: "解碼錯誤", data: []byte("damaged ogg"), change: func(c *titleCueCatalog) {
			s := sha256.Sum256([]byte("damaged ogg"))
			c.Cues[0].SHA256 = hex.EncodeToString(s[:])
		}},
		{name: "缺檔", missing: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			var cat titleCueCatalog
			if err := json.Unmarshal(raw, &cat); err != nil {
				t.Fatal(err)
			}
			if tc.change != nil {
				tc.change(&cat)
			}
			b, err := json.Marshal(cat)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "manifest.json")
			if err := os.WriteFile(path, b, 0600); err != nil {
				t.Fatal(err)
			}
			if !tc.missing {
				data := encoded
				if tc.data != nil {
					data = tc.data
				}
				if err := os.WriteFile(filepath.Join(dir, "title-thunder.ogg"), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := loadTitleCue(path); err == nil {
				t.Fatal("應拒絕不符契約的音訊")
			}
		})
	}
}

type fakeTitleCue struct {
	playing       bool
	plays, pauses int
	rewindErr     error
}

func (p *fakeTitleCue) Play()           { p.plays++; p.playing = true }
func (p *fakeTitleCue) Pause()          { p.pauses++; p.playing = false }
func (p *fakeTitleCue) IsPlaying() bool { return p.playing }
func (p *fakeTitleCue) Rewind() error   { return p.rewindErr }

func TestTitleCueWaitAndStop(t *testing.T) {
	p := &fakeTitleCue{}
	a := &titleSound{player: p, frame: -1}
	f := title.Frame{Index: 0, Thunder: true, ThunderCount: 2}
	for i := 0; i < 5; i++ {
		wait, err := a.step(f)
		if err != nil || !wait || p.plays != 1 {
			t.Fatalf("第一聲播放時不得換格或重播：%v %v %d", wait, err, p.plays)
		}
	}
	p.playing = false
	if wait, err := a.step(f); err != nil || !wait || p.plays != 2 {
		t.Fatal("第一聲完成後應播第二聲")
	}
	p.playing = false
	for i := 0; i < 5; i++ {
		wait, err := a.step(f)
		if err != nil || wait || p.plays != 2 {
			t.Fatal("第二聲完成後應開放tick等待，不能重播")
		}
	}
	f.Index = 1
	f.ThunderCount = 1
	if wait, err := a.step(f); err != nil || !wait || p.plays != 3 {
		t.Fatal("下一個cue未播放")
	}
	a.stop()
	if p.playing || a.remaining != 0 || a.frame != -1 || p.pauses != 1 {
		t.Fatal("離開標題未停止")
	}
	p.rewindErr = errors.New("seek failed")
	if _, err := a.step(f); err == nil {
		t.Fatal("重播錯誤應傳回")
	}
}

func TestTitleAudioWaitsForDrawAndCompletion(t *testing.T) {
	f := []title.Frame{{Index: 0, Thunder: true, ThunderCount: 2}, {Index: 1}}
	p := &fakeTitleCue{}
	g := &game{player: title.NewPlayer(f, 0), frame: f[0], titleSound: &titleSound{player: p, frame: -1}}
	if err := g.advanceTitleAudio(); err != nil {
		t.Fatal(err)
	}
	if p.plays != 0 {
		t.Fatal("尚未呈現畫格就播音")
	}
	g.drawn = true
	for i := 0; i < 120; i++ {
		if err := g.advanceTitleAudio(); err != nil {
			t.Fatal(err)
		}
	}
	if g.frame.Index != 0 || p.plays != 1 {
		t.Fatal("音訊未完成就換格")
	}
	p.playing = false
	if err := g.advanceTitleAudio(); err != nil {
		t.Fatal(err)
	}
	if p.plays != 2 || g.frame.Index != 0 {
		t.Fatal("第二聲未依序播放")
	}
	p.playing = false
	for i := 0; i < 120 && g.frame.Index == 0; i++ {
		if err := g.advanceTitleAudio(); err != nil {
			t.Fatal(err)
		}
	}
	if g.frame.Index != 1 || g.drawn {
		t.Fatal("完成後應換格並等待下一次呈現")
	}
}
