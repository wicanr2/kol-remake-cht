package ui

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
)

// 固定正常原版收據的格、背景、人物與最後繪製相位，驗真正共用繪製出口。
// 這是同狀態呈現抽樣；正常玩家路徑另驗，不把本測試稱為整場戰鬥對拍。
func TestBattleMarkerOriginalPixels(t *testing.T) {
	if os.Getenv("KOL_MARKER_ORACLE") != "1" {
		t.Skip("需要 Docker/Xvfb 與 KOL_MARKER_ORACLE=1")
	}
	// Ebitengine 結束 RunGame 後禁止再建圖片；隔離子行程，避免影響其他 UI 測試。
	if os.Getenv("KOL_MARKER_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestBattleMarkerOriginalPixels$", "-test.v")
		cmd.Env = append(os.Environ(), "KOL_MARKER_CHILD=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("圓環原版像素子行程：%v\n%s", err, out)
		}
		t.Log(string(out))
		return
	}
	root := "../../docs/verification/battle-target-cursor-20261001"
	body, err := os.ReadFile(filepath.Join(root, "lifecycle-receipt.json"))
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		Checks []markerCase `json:"checks"`
	}
	if err := json.Unmarshal(body, &receipt); err != nil {
		t.Fatal(err)
	}
	if len(receipt.Checks) != 52 {
		t.Fatalf("生命週期收據框數=%d，應為52", len(receipt.Checks))
	}
	g := &markerOracleGame{t: t, root: root, cases: receipt.Checks}
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
}

type markerCase struct {
	Run, Frame     string
	Row, Col, Tile int
	Sprite         int `json:"sprite_code"`
	Shown          bool
	Unit           *int
}

type markerOracleGame struct {
	t     *testing.T
	root  string
	cases []markerCase
	done  bool
}

func (g *markerOracleGame) Update() error {
	if g.done {
		return ebiten.Termination
	}
	return nil
}
func (g *markerOracleGame) Layout(int, int) (int, int) { return 320, 200 }

type ringSprites struct{ img *ebiten.Image }

func (s ringSprites) Get(archive string, index int) (*ebiten.Image, error) {
	if archive != "KOLPICS-MARKERS" || index != 0 {
		return nil, fmt.Errorf("非預期素材 %s[%d]", archive, index)
	}
	return s.img, nil
}

func (g *markerOracleGame) Draw(*ebiten.Image) {
	if g.done {
		return
	}
	g.done = true
	load := func(path string) image.Image {
		f, err := os.Open(path)
		if err != nil {
			g.t.Fatal(err)
		}
		defer f.Close()
		img, err := png.Decode(f)
		if err != nil {
			g.t.Fatal(err)
		}
		return img
	}
	atlas := load("../../assets/original-derived-png/ega/BATTLE/battle-tiles-08.png")
	markers := load("../../assets/original-derived-png/ega/KOLPICS-MARKERS/kolpics-markers.png")
	tiles := NewBattleTiles(atlas)
	sprites := ringSprites{ebiten.NewImageFromImage(markers)}
	var results []map[string]any
	for _, c := range g.cases {
		s := &battle.State{}
		s.Plan.Prompt = battle.PromptCommand
		s.Plan.Summary, s.Plan.SummaryShown = battle.Cell{Row: c.Row, Col: c.Col}, c.Shown
		s.Plan.Cursor = battle.TargetCursor{Row: c.Row, Col: c.Col}
		// 用已確認的圖示列／輸入路由設定層，不從重製版結果反推預期。
		if c.Shown {
			switch {
			case c.Run == "gui7" && c.Frame == "gui7-plan-07.png":
				s.Plan.Prompt = battle.PromptMoveForm
			case c.Run == "gui7" && c.Frame == "gui7-plan-31.png":
				s.Plan.Prompt = battle.PromptUnarmedForm
			case c.Run == "gui7" && c.Frame == "gui7-plan-34.png":
				s.Plan.Prompt = battle.PromptGuardForm
			case c.Run == "gui6", c.Run == "gui9":
				s.Plan.Prompt = battle.PromptTarget
			case c.Run == "gui10" && c.Frame != "gui10-plan-25.png" && c.Frame != "gui10-plan-26.png":
				s.Plan.Prompt = battle.PromptTarget
			}
			tiles.ringPhase = *c.Unit - 5
		}
		base := tiles.Tile(c.Tile)
		if c.Sprite != 0 {
			base = tiles.WithSprite(c.Tile, c.Sprite-1)
		}
		original := load(filepath.Join(g.root, c.Frame))
		for _, scale := range []int{1, 3} {
			dst := ebiten.NewImage(320*scale, 200*scale)
			x, y := (152+c.Col*16)*scale, (8+c.Row*16)*scale
			drawScaled(dst, base, x, y, scale)
			if err := drawBattleTargetMarker(dst, sprites, tiles, s, scale); err != nil {
				g.t.Fatal(err)
			}
			crop := dst.SubImage(image.Rect(x, y, x+16*scale, y+16*scale)).(*ebiten.Image)
			actual := make([]byte, 16*scale*16*scale*4)
			crop.ReadPixels(actual)
			diff := 0
			for yy := 0; yy < 16*scale; yy++ {
				for xx := 0; xx < 16*scale; xx++ {
					r, gr, b, a := original.At(152+c.Col*16+xx/scale, 8+c.Row*16+yy/scale).RGBA()
					i := (yy*16*scale + xx) * 4
					if actual[i] != byte(r>>8) || actual[i+1] != byte(gr>>8) || actual[i+2] != byte(b>>8) || actual[i+3] != byte(a>>8) {
						diff++
					}
				}
			}
			if diff != 0 {
				g.t.Errorf("%s 倍率%d差%d像素", c.Frame, scale, diff)
			}
			results = append(results, map[string]any{"frame": c.Frame, "scale": scale, "different_pixels": diff, "prompt": s.Plan.Prompt})
			if out := os.Getenv("KOL_MARKER_OUTPUT"); out != "" {
				im := image.NewRGBA(image.Rect(0, 0, 16*scale, 16*scale))
				copy(im.Pix, actual)
				f, err := os.Create(filepath.Join(out, fmt.Sprintf("remake-%s-x%d.png", c.Frame, scale)))
				if err != nil {
					g.t.Fatal(err)
				}
				err = png.Encode(f, im)
				closeErr := f.Close()
				if err != nil {
					g.t.Fatal(err)
				}
				if closeErr != nil {
					g.t.Fatal(closeErr)
				}
			}
			dst.Deallocate()
		}
	}
	if _, err := tiles.TargetRing(image.NewRGBA(image.Rect(0, 0, 16, 16))); err == nil {
		g.t.Error("錯誤圖集尺寸未被拒絕")
	}
	if out := os.Getenv("KOL_MARKER_OUTPUT"); out != "" {
		body, err := json.MarshalIndent(map[string]any{"schema": "kol-remake-target-pixels-v1", "checks": results, "full_battle_parity": false}, "", "  ")
		if err != nil {
			g.t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(out, "remake-pixel-checks.json"), append(body, '\n'), 0600); err != nil {
			g.t.Fatal(err)
		}
	}
}

func TestBattleMarkerPhaseWaitsAndReverses(t *testing.T) {
	b := &BattleTiles{}
	for _, want := range []int{1, 2, 3, 4, 3, 2, 1, 0, 1} {
		old := b.ringPhase
		b.AdvanceTargetMarker(2)
		if b.ringPhase != old {
			t.Fatal("三tick以前不應換相位")
		}
		b.AdvanceTargetMarker(1)
		if b.ringPhase != want {
			t.Fatalf("相位=%d，應為%d", b.ringPhase, want)
		}
	}
}
