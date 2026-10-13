package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/jfreymuth/oggvorbis"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/flow"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
)

type musicCatalog struct {
	Schema, Rights string
	Tracks         []musicTrack
}
type musicTrack struct {
	ID, File, SHA256, Source string
	SampleRate               int `json:"sample_rate"`
	Channels                 int
	Duration                 float64
	Loop                     bool
	HumanListening           string `json:"human_listening"`
}

// 完整驗四首已接受曲目，才回傳PCM。原雷聲catalog與cue保持獨立。
func loadMusicCatalog(path string) (map[string][]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cat musicCatalog
	if err = json.Unmarshal(raw, &cat); err != nil {
		return nil, err
	}
	if cat.Schema != "kol-hd-music/1" || cat.Rights == "" || len(cat.Tracks) != 4 {
		return nil, fmt.Errorf("配樂目錄版本或曲數不符")
	}
	expected := map[string]float64{"title": 72, "inn": 80, "exploration": 90, "combat": 60}
	result := map[string][]byte{}
	for _, t := range cat.Tracks {
		seconds, ok := expected[t.ID]
		if !ok || result[t.ID] != nil || t.File != t.ID+"-chamber.ogg" || t.Duration != seconds || t.SampleRate != 44100 || t.Channels != 2 || !t.Loop || !strings.HasPrefix(t.HumanListening, "已驗收：") || t.Source == "" {
			return nil, fmt.Errorf("配樂 %s 契約或聆聽紀錄不符", t.ID)
		}
		want, e := hex.DecodeString(t.SHA256)
		if e != nil || len(want) != sha256.Size {
			return nil, fmt.Errorf("配樂雜湊格式不符")
		}
		f, e := os.Open(filepath.Join(filepath.Dir(path), t.File))
		if e != nil {
			return nil, e
		}
		encoded, e := io.ReadAll(io.LimitReader(f, (8<<20)+1))
		f.Close()
		if e != nil || len(encoded) > 8<<20 {
			return nil, fmt.Errorf("配樂讀取失敗或過大")
		}
		got := sha256.Sum256(encoded)
		if !bytes.Equal(want, got[:]) {
			return nil, fmt.Errorf("配樂 %s 雜湊不符", t.ID)
		}
		header, e := oggvorbis.NewReader(bytes.NewReader(encoded))
		if e != nil {
			return nil, e
		}
		if header.Channels() != 2 || header.SampleRate() != 44100 {
			return nil, fmt.Errorf("配樂實際聲道或取樣率不符")
		}
		stream, e := vorbis.DecodeF32(bytes.NewReader(encoded))
		if e != nil {
			return nil, e
		}
		pcm, e := io.ReadAll(io.LimitReader(stream, 90*44100*8+1))
		if e != nil || len(pcm) != int(seconds)*44100*8 {
			return nil, fmt.Errorf("配樂 %s 實際時長不符", t.ID)
		}
		result[t.ID] = pcm
	}
	return result, nil
}

type musicPlayer interface {
	Play()
	Pause()
	IsPlaying() bool
	Rewind() error
	SetVolume(float64)
	Close() error
	Position() time.Duration
}
type musicSound struct {
	players map[string]musicPlayer
	gain    map[string]float64
	cue     string
}

func newMusicSound(path string) (*musicSound, error) {
	pcm, err := loadMusicCatalog(path)
	if err != nil {
		return nil, err
	}
	ctx := audio.CurrentContext()
	if ctx == nil {
		ctx = audio.NewContext(44100)
	}
	if ctx.SampleRate() != 44100 {
		return nil, fmt.Errorf("配樂音訊context取樣率不符")
	}
	s := &musicSound{players: map[string]musicPlayer{}, gain: map[string]float64{}}
	for _, id := range []string{"title", "inn", "exploration", "combat"} {
		loop := audio.NewInfiniteLoopF32(bytes.NewReader(pcm[id]), int64(len(pcm[id])))
		p, e := ctx.NewPlayerF32(loop)
		if e != nil {
			s.close()
			return nil, e
		}
		p.SetVolume(0)
		s.players[id] = p
	}
	return s, nil
}

// 60TPS下約0.3秒淡入淡出。只改音量，場景相同與圖像切換都不重播。
func (s *musicSound) step(cue string, enabled bool) error {
	if cue != "" && s.players[cue] == nil {
		return fmt.Errorf("未知配樂：%s", cue)
	}
	if cue != s.cue {
		if p := s.players[cue]; p != nil && !p.IsPlaying() {
			if err := p.Rewind(); err != nil {
				return err
			}
		}
		s.cue = cue
	}
	for id, p := range s.players {
		want := enabled && cue == id
		if want {
			s.gain[id] = math.Min(1, s.gain[id]+1.0/18)
			p.SetVolume(s.gain[id])
			if !p.IsPlaying() {
				p.Play()
			}
		} else {
			s.gain[id] = math.Max(0, s.gain[id]-1.0/18)
			p.SetVolume(s.gain[id])
			if s.gain[id] == 0 && p.IsPlaying() {
				p.Pause()
			}
		}
	}
	return nil
}
func (s *musicSound) close() {
	for _, p := range s.players {
		p.Pause()
		p.Close()
	}
}

func isMusicInn(in *town.Inn) bool {
	// 原始旅店樣本與正常EnterBuilding都保留這組素材定位，Apply會改寫Title。
	return in != nil && in.Portrait.Archive == "TWN7PICS" && in.Portrait.Index == 12
}
func (g *game) musicCue() string {
	if g.endingReached() {
		return ""
	}
	if g.scene == sceneTitle {
		return "title"
	}
	if g.scene == sceneInn || g.scene == sceneBuilding {
		if isMusicInn(g.inn) {
			return "inn"
		}
		return ""
	}
	s := g.play
	if g.scene == sceneFlow && g.flow != nil {
		switch g.flow.Scene {
		case flow.SceneTitle:
			return "title"
		case flow.SceneInn:
			if isMusicInn(g.flow.Inn) {
				return "inn"
			}
			return ""
		case flow.SceneTown:
			if g.leftTown() {
				s = g.play
			} else {
				s = g.flow.Town
			}
		default:
			return ""
		}
	} else if g.scene != scenePlay {
		return ""
	}
	if g.quest != nil && g.quest.state != nil {
		switch g.quest.state.Stage {
		case battle.StageEncounter, battle.StageField, battle.StageFlee:
			return "combat"
		}
		return ""
	}
	if s != nil && s.LeftTown {
		return "exploration"
	}
	if s != nil && !s.LeftTown && s.Mode == town.ModeInside && isMusicInn(s.Inside) {
		return "inn"
	}
	return ""
}
func (g *game) setMusic(enabled bool) error {
	if enabled && g.music == nil && g.interactive {
		s, err := newMusicSound(g.musicPath)
		if err != nil {
			return fmt.Errorf("配樂無法使用：%w", err)
		}
		g.music = s
	}
	g.musicEnabled = enabled
	return nil
}
func (g *game) advanceMusic() error {
	if g.music == nil {
		return nil
	}
	cue := g.musicCue()
	previous := g.music.cue
	if err := g.music.step(cue, g.musicEnabled); err != nil {
		return err
	}
	if g.logInput && previous != cue {
		fmt.Fprintf(os.Stderr, "music cue=%s enabled=%t\n", cue, g.musicEnabled)
	}
	return nil
}

func (g *game) musicUnavailableNotice() {
	g.notice = []string{"配樂無法使用，已保留目前畫面。"}
	if g.strings == nil {
		g.notice = []string{"Background music is unavailable."}
	}
}
func (g *game) musicDiagnostics() {
	if g.logInput && g.music != nil {
		fmt.Fprintf(os.Stderr, "music-position cue=%s enabled=%t title_ms=%d inn_ms=%d\n", g.music.cue, g.musicEnabled, g.music.players["title"].Position().Milliseconds(), g.music.players["inn"].Position().Milliseconds())
	}
}
func (g *game) updateMusicControls() bool {
	x, y := g.cursorPosition()
	clicked := g.strings != nil && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && image.Pt(x, y).In(ui.MusicControl)
	if !inpututil.IsKeyJustPressed(ebiten.KeyF9) && !clicked {
		return false
	}
	if err := g.setMusic(!g.musicEnabled); err != nil {
		fmt.Fprintln(os.Stderr, err)
		g.musicUnavailableNotice()
	}
	g.displayEvent(fmt.Sprintf("music-%t", g.musicEnabled))
	return true
}
