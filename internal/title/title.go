// Package title 依 RE-047 重建的腳本播放原版標題序列。
//
// 序列本身出自 `TITLE.EXE` 的 `_main`，是一段線性程式碼而不是資料表；這裡讀的
// `docs/re/title-sequence.json` 是那段程式碼的逐呼叫轉寫，每一步都帶原始呼叫位址。
// 等待單位是 BIOS tick，與原版相同；兩處亂數（SET 動畫的閃光張數、BACKPICS[11] 與
// [13] 的選擇）由可注入的亂數來源決定，因此可以重播。
package title

import (
	"encoding/json"
	"fmt"
	"os"
)

// TickHz 是 BIOS 計時器的頻率；原版所有等待都以它為單位。
const TickHz = 18.2065

// Rect 是一個 archive 的疊圖矩形，座標與尺寸都出自原版程式碼裡的常數。
type Rect struct {
	Archive string `json:"archive"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}

// Op 是序列裡的一步。欄位隨 op 而異，沒有用到的留空。
type Op struct {
	EA         string `json:"ea"`
	Op         string `json:"op"`
	Archive    string `json:"archive"`
	Image      *int   `json:"image"`
	ImageRef   string `json:"image_ref"`
	Frame      *int   `json:"frame"`
	Background *int   `json:"background"`
	Ticks      int    `json:"ticks"`
	Set        *int   `json:"set"`
	Name       string `json:"name"`
	Options    []int  `json:"options"`
}

type setOption struct {
	Frame      int `json:"frame"`
	Background int `json:"background"`
}

type setOp struct {
	EA      string      `json:"ea"`
	Op      string      `json:"op"`
	Frame   *int        `json:"frame"`
	Backgnd *int        `json:"background"`
	Ticks   int         `json:"ticks"`
	Name    string      `json:"name"`
	Options []setOption `json:"options"`
}

type setAnimation struct {
	EA         string  `json:"ea"`
	Iterations int     `json:"iterations"`
	Steps      []setOp `json:"steps"`
	After      string  `json:"after"`
}

type Sequence struct {
	Schema       string       `json:"schema"`
	OverlayRects []Rect   `json:"overlay_rects"`
	SetAnimation setAnimation `json:"set_animation"`
	Steps        []Op         `json:"steps"`
}

// Load 讀取序列並做基本一致性檢查。
func Load(path string) (*Sequence, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取標題序列失敗：%w", err)
	}
	var s Sequence
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("解析標題序列失敗：%w", err)
	}
	if s.Schema != "kol-title-sequence-v1" {
		return nil, fmt.Errorf("非預期的標題序列 schema：%s", s.Schema)
	}
	if len(s.Steps) == 0 || len(s.OverlayRects) != 4 || s.SetAnimation.Iterations != 4 {
		return nil, fmt.Errorf("標題序列內容不完整")
	}
	return &s, nil
}

// Rect 回傳指定 archive 的疊圖矩形。
func (s *Sequence) Rect(archive string) (Rect, bool) {
	for _, r := range s.OverlayRects {
		if r.Archive == archive {
			return r, true
		}
	}
	return Rect{}, false
}

// Rand 是可注入的亂數來源；n 為選項數，回傳 0..n-1。
type Rand interface {
	Choose(n int) int
}

// Frame 是序列送出的一個畫面。
type Frame struct {
	Index   int    // 在序列中的順序
	EA      string // 原版送出這個畫面的呼叫位址
	Kind    string // present_back / present_composed / draw_overlay
	Back    int    // 背景 BACKPICS 圖號
	Archive string // 疊圖 archive；沒有疊圖時為空
	Overlay int    // 疊圖張號
	Ticks   int    // 這個畫面停留的 tick 數；雷聲等非 tick 等待記為 0
	Thunder bool   // 送出這個畫面之後是否放雷聲
}

// Expand 把序列展開成一輪播放的畫面清單。
//
// 這裡只做畫面與等待，喇叭雷聲另外標記：`sub_10F75` 的長度是以開機期校正值算出的
// 忙等，不是 tick 數，靜態讀不出來，不能假裝知道。
func (s *Sequence) Expand(rng Rand) ([]Frame, error) {
	var out []Frame
	back, archive, overlay := -1, "", -1
	storm := 0
	emit := func(ea, kind string) {
		out = append(out, Frame{Index: len(out), EA: ea, Kind: kind, Back: back,
			Archive: archive, Overlay: overlay})
	}
	wait := func(t int) {
		if len(out) > 0 {
			out[len(out)-1].Ticks += t
		}
	}
	thunder := func() {
		if len(out) > 0 {
			out[len(out)-1].Thunder = true
		}
	}
	lightning := func(ea string) {
		back, archive, overlay = 6, "", -1
		emit(ea, "present_back")
		thunder()
	}
	runSet := func(set int) error {
		name := fmt.Sprintf("SET%dPICS", set)
		for i := 0; i < s.SetAnimation.Iterations; i++ {
			for _, st := range s.SetAnimation.Steps {
				switch st.Op {
				case "compose_overlay":
					back, archive, overlay = *st.Backgnd, name, *st.Frame
				case "present_composed":
					emit(st.EA, "present_composed")
				case "draw_overlay":
					overlay = *st.Frame
					emit(st.EA, "draw_overlay")
				case "wait_ticks":
					wait(st.Ticks)
				case "random_choice":
					opt := st.Options[rng.Choose(len(st.Options))]
					back, archive, overlay = opt.Background, name, opt.Frame
					emit(st.EA, "present_composed")
				case "thunder":
					thunder()
				default:
					return fmt.Errorf("未知的 SET 動畫 op %q", st.Op)
				}
			}
		}
		lightning(s.SetAnimation.EA)
		return nil
	}

	for _, op := range s.Steps {
		switch op.Op {
		case "load_archive":
			archive = op.Archive
		case "lightning":
			lightning(op.EA)
		case "set_animation":
			if err := runSet(*op.Set); err != nil {
				return nil, err
			}
		case "present_back":
			if op.ImageRef == "storm_screen" {
				back = storm
			} else {
				back = *op.Image
			}
			archive, overlay = "", -1
			emit(op.EA, "present_back")
		case "wait_ticks":
			wait(op.Ticks)
		case "random_choice":
			storm = op.Options[rng.Choose(len(op.Options))]
		case "compose_overlay":
			back, archive, overlay = *op.Background, op.Archive, *op.Frame
		case "present_composed":
			emit(op.EA, "present_composed")
		case "draw_overlay":
			archive, overlay = op.Archive, *op.Frame
			emit(op.EA, "draw_overlay")
		case "thunder":
			thunder()
		case "loop":
			return out, nil
		default:
			return nil, fmt.Errorf("未知的序列 op %q", op.Op)
		}
	}
	return out, nil
}
