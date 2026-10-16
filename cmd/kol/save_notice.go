package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
	"image/color"
)

var saveNotices = map[string]struct{ zh, en []string }{
	"import": {[]string{"讀取舊版存檔完成，原始檔案保留。", "舊格式未保存的能力、裝備變更", "與世界時間無法還原。", "缺項使用原始角色資料與起始時間。", "按任意鍵繼續。"}, []string{"V1 SAVE IMPORTED. ORIGINAL KEPT.", "MISSING CHANGES CANNOT BE RECOVERED.", "MISSING DATA USES ORIGINAL CHARACTERS", "AND THE STARTING WORLD DATE.", "PRESS ANY KEY TO CONTINUE."}},
	"create": {[]string{"角色保存失敗，資料尚未改變。", "請確認存檔路徑與空間。", "按任意鍵繼續，再按 Enter 重試。"}, []string{"CHARACTER SAVE FAILED. DRAFT KEPT.", "CHECK SAVE PATH AND FREE SPACE.", "PRESS ANY KEY, THEN ENTER TO RETRY."}},
	"delete": {[]string{"刪除保存失敗，角色已保留。", "角色在隊伍中時不能刪除。", "請確認存檔路徑與空間。", "按任意鍵繼續。"}, []string{"DELETE SAVE FAILED. CHARACTER KEPT.", "PARTY MEMBERS CANNOT BE DELETED.", "CHECK SAVE PATH AND FREE SPACE.", "PRESS ANY KEY TO CONTINUE."}},
	"full":   {[]string{"角色已滿，最多十六人。", "按任意鍵繼續。"}, []string{"ROSTER FULL: MAXIMUM 16 CHARACTERS.", "PRESS ANY KEY TO CONTINUE."}},
}

func (g *game) saveNotice(kind string) []string {
	if g.cjk != nil {
		return saveNotices[kind].zh
	}
	return saveNotices[kind].en
}

func (g *game) drawNotice(dst *ebiten.Image) {
	dst.Fill(color.RGBA{A: 255})
	f := g.cjk
	if f == nil {
		f = g.large
	}
	if f == nil {
		return
	}
	// 留出四邊安全範圍；長錯誤訊息逐列換行，不讓檔案路徑蓋過提示。
	cell, lineHeight, gap, left, top := 16, 22, 8, 24, 32
	if g.cjk == nil {
		cell, lineHeight, gap, left, top = 8, 12, 4, 16, 16
	}
	width, height := dst.Bounds().Dx(), dst.Bounds().Dy()
	if ui.NativeRendering(dst) {
		width /= 2
		height /= 2
	}
	columns := (width - 2*left) / cell
	if columns < 1 {
		return
	}
	y := top
	for _, line := range g.notice {
		runes := []rune(line)
		for len(runes) > 0 && y+cell < height-left {
			n := len(runes)
			if n > columns {
				n = columns
			}
			f.Draw(dst, left, y, string(runes[:n]), 255, 255, 255)
			runes = runes[n:]
			y += lineHeight
		}
		y += gap
	}
}
