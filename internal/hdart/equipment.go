// Package hdart 定義HD呈現資料，不包含GPU、檔案載入或遊戲規則。
package hdart

type ImageInfo struct {
	SHA256 string `json:"sha256"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
type Source struct {
	Archive     string `json:"archive"`
	Index       int    `json:"index"`
	Size        [2]int `json:"source_size"`
	PNG         string `json:"source_png_sha256"`
	RGB         string `json:"source_rgb_sha256"`
	RenderedRGB string `json:"rendered_rgb_sha256"`
}
type Part struct {
	File  string     `json:"file"`
	Rect  [4]int     `json:"rect"`
	World [4]float64 `json:"world"`
	Kind  string     `json:"kind"`
	RGBA  string     `json:"native_rgba_sha256"`
}
type Prop struct {
	File       string `json:"file"`
	Rect       [4]int `json:"rect"`
	Paint      [4]int `json:"paint"`
	SourceSize [2]int `json:"source_size"`
	RGBA       string `json:"native_rgba_sha256"`
}
type Grip struct {
	Rect      [4]int `json:"rect"`
	HandIndex int    `json:"hand_index"`
}
type Mesh struct {
	NativeWaist int     `json:"native_waist"`
	WaistWorld  float64 `json:"waist_world"`
}
type Equipment struct {
	Schema        string               `json:"schema"`
	Rights        string               `json:"rights"`
	SourceZIP     string               `json:"source_zip_sha256"`
	LayoutSHA256  string               `json:"source_layout_sha256"`
	Files         map[string]ImageInfo `json:"files"`
	Parts         map[string]Part      `json:"parts"`
	Props         map[string]Prop      `json:"props"`
	Grips         map[string][]Grip    `json:"grips"`
	Meshes        map[string]Mesh      `json:"meshes"`
	BaseWaists    map[string]float64   `json:"base_waists"`
	HeadBands     map[string][3]int    `json:"head_bands"`
	HandPaint     map[string][4]int    `json:"hand_paint"`
	Sources       []Source             `json:"sources"`
	ArchiveCounts map[string]int       `json:"archive_counts"`
}

// Binding 只在記憶體存在，由合法匯入的typed版面推導，沒有JSON欄位。
type Binding struct {
	BaseOrigins      [6]int     `json:"-"`
	SourceHands      [12][4]int `json:"-"`
	DefaultShieldRGB string     `json:"-"`
}
