package battle

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// WildernessScenes 是治療那一條的六段敘述（RE-112／128）。來源是 `TOWN0` 的
// payload 4，原版把整份載進 DGROUP `0x22D0`，敘述從 `+0x80` 起。
type WildernessScenes struct {
	Schema string `json:"schema"`
	// Scenes 是六段，一段四行。偶數段是遠景、奇數段是走近之後看到的。
	Scenes [][]string `json:"scenes"`
}

// SceneCount 是那六段。
const SceneCount = 6

// LoadWildernessScenes 讀那六段敘述。少一段就失敗——空白的一頁比沒有那一頁還糟。
func LoadWildernessScenes(path string) (*WildernessScenes, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var w WildernessScenes
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, fmt.Errorf("%s：%w", path, err)
	}
	if !strings.HasPrefix(w.Schema, "kol-wilderness-scenes-") {
		return nil, fmt.Errorf("%s：schema 是 %q", path, w.Schema)
	}
	if len(w.Scenes) != SceneCount {
		return nil, fmt.Errorf("%s：有 %d 段，應該是 %d", path, len(w.Scenes), SceneCount)
	}
	for i, lines := range w.Scenes {
		if len(lines) == 0 {
			return nil, fmt.Errorf("%s：第 %d 段是空的", path, i)
		}
	}
	return &w, nil
}

// Scene 是第 n 段的四行；超出範圍回 nil。
func (w *WildernessScenes) Scene(n int) []string {
	if w == nil || n < 0 || n >= len(w.Scenes) {
		return nil
	}
	return w.Scenes[n]
}
