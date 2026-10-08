package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/wicanr2/kol-remake-cht/internal/title"
)

const titleSampleRate = 44100

type titleCueCatalog struct {
	Schema string `json:"schema"`
	Cues   []struct {
		ID             string  `json:"id"`
		File           string  `json:"file"`
		SHA256         string  `json:"sha256"`
		SampleRate     int     `json:"sample_rate_hz"`
		Channels       int     `json:"channels"`
		Duration       float64 `json:"duration_seconds"`
		Loop           bool    `json:"loop"`
		HumanListening string  `json:"human_listening"`
	} `json:"cues"`
}

// loadTitleCue 只讀已驗收OGG；PCM只存在記憶體，runtime不讀研究WAV或原版資料。
func loadTitleCue(manifest string) ([]byte, error) {
	raw, err := os.ReadFile(manifest)
	if err != nil {
		return nil, fmt.Errorf("讀取音訊目錄：%w", err)
	}
	var cat titleCueCatalog
	if err = json.Unmarshal(raw, &cat); err != nil {
		return nil, fmt.Errorf("解析音訊目錄：%w", err)
	}
	if cat.Schema != "kol-audio-catalog-v1" || len(cat.Cues) != 1 {
		return nil, fmt.Errorf("音訊目錄格式不符")
	}
	c := cat.Cues[0]
	if c.ID != "title-thunder" || c.File != "title-thunder.ogg" || c.SampleRate != titleSampleRate || c.Channels != 1 || c.Loop || !(c.Duration > 0 && c.Duration < 5) || !strings.HasPrefix(c.HumanListening, "已驗收：") {
		return nil, fmt.Errorf("標題雷聲契約或聆聽驗收不符")
	}
	want, err := hex.DecodeString(c.SHA256)
	if err != nil || len(want) != sha256.Size {
		return nil, fmt.Errorf("雷聲SHA-256格式不符")
	}
	f, err := os.Open(filepath.Join(filepath.Dir(manifest), c.File))
	if err != nil {
		return nil, fmt.Errorf("讀取雷聲：%w", err)
	}
	defer f.Close()
	encoded, err := io.ReadAll(io.LimitReader(f, 1<<20+1))
	if err != nil || len(encoded) > 1<<20 {
		return nil, fmt.Errorf("雷聲檔案讀取失敗或過大")
	}
	got := sha256.Sum256(encoded)
	if !bytes.Equal(got[:], want) {
		return nil, fmt.Errorf("雷聲SHA-256不符")
	}
	stream, err := vorbis.DecodeF32(bytes.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("雷聲OGG解碼：%w", err)
	}
	if stream.SampleRate() != titleSampleRate {
		return nil, fmt.Errorf("雷聲取樣率不符")
	}
	pcm, err := io.ReadAll(io.LimitReader(stream, 5*titleSampleRate*8+1))
	if err != nil || len(pcm) == 0 || len(pcm)%8 != 0 || len(pcm) > 5*titleSampleRate*8 {
		return nil, fmt.Errorf("雷聲解碼長度不符")
	}
	seconds := float64(len(pcm)) / float64(titleSampleRate*8)
	if seconds < c.Duration-1.0/titleSampleRate || seconds > c.Duration+1.0/titleSampleRate {
		return nil, fmt.Errorf("雷聲實際時長不符")
	}
	return pcm, nil
}

type titleCuePlayer interface {
	Play()
	Pause()
	IsPlaying() bool
	Rewind() error
}

// titleSound 依實際播放完成閘門順序送cue，不用畫面幀數推定聲音已播放。
type titleSound struct {
	player    titleCuePlayer
	frame     int
	remaining int
	started   int
}

func newTitleSound(manifest string) (*titleSound, error) {
	pcm, err := loadTitleCue(manifest)
	if err != nil {
		return nil, err
	}
	ctx := audio.CurrentContext()
	if ctx == nil {
		ctx = audio.NewContext(titleSampleRate)
	}
	if ctx.SampleRate() != titleSampleRate {
		return nil, fmt.Errorf("音訊context取樣率不符")
	}
	return &titleSound{player: ctx.NewPlayerF32FromBytes(pcm), frame: -1}, nil
}

func (a *titleSound) step(f title.Frame) (bool, error) {
	if a.frame != f.Index {
		a.frame = f.Index
		a.remaining = f.CueCount()
	}
	if a.player.IsPlaying() {
		return true, nil
	}
	if a.remaining == 0 {
		return false, nil
	}
	if err := a.player.Rewind(); err != nil {
		return false, fmt.Errorf("雷聲重播：%w", err)
	}
	a.player.Play()
	a.remaining--
	a.started++
	return true, nil
}

func (a *titleSound) stop() {
	a.player.Pause()
	a.frame = -1
	a.remaining = 0
}

func (g *game) advanceTitleAudio() error {
	if !g.drawn {
		return nil
	}
	if g.titleSound == nil {
		a, err := newTitleSound(g.titleAudioPath)
		if err != nil {
			return err
		}
		g.titleSound = a
	}
	before := g.titleSound.started
	wait, err := g.titleSound.step(g.frame)
	if err != nil {
		return err
	}
	if g.logInput && g.titleSound.started != before {
		fmt.Fprintf(os.Stderr, "audio cue=title-thunder frame=%d started=%d remaining=%d\n", g.frame.Index, g.titleSound.started, g.titleSound.remaining)
	}
	if !wait {
		beforeFrame := g.frame.Index
		g.player.Advance(title.TickHz / 60)
		g.frame = g.player.Current()
		if g.frame.Index != beforeFrame {
			g.drawn = false
		}
	}
	return nil
}
