//go:build oracle

// 本檔只供隔離驗證，不參與正式遊戲。
package oracle

import (
	"crypto/sha256"
	"fmt"
	"github.com/wicanr2/dosgolem/oracle"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// 僅使用原版入口與 BIOS 鍵盤，不改記憶體、不替換執行檔。
// 這是有界啟動探針；執行無錯不代表已到達書本或可玩。
func TestNormalStartupUpdatedDosgolem(t *testing.T) {
	root := os.Getenv("KOL_ORACLE_ROOT")
	if root == "" {
		t.Skip("需要 KOL_ORACLE_ROOT 指向玩家自備原版目錄")
	}
	o, err := oracle.Load(filepath.Join(root, "KOL.EXE"), root)
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()
	visibleFrames := 0
	for i, key := range []string{"", "Return", "Return", "6", "Return", "Return"} {
		if key == "6" {
			err = o.TypeKeys(key)
		} else if key != "" {
			err = o.SendKeys(key)
		}
		if err != nil {
			t.Fatal(err)
		}
		err = o.Run(20_000_000)
		w, h, px := o.Screen()
		litIndices := 0
		for _, p := range px {
			if p != 0 {
				litIndices++
			}
		}
		paletteColors := 0
		for _, c := range o.Palette() {
			if c != [3]uint8{} {
				paletteColors++
			}
		}
		_, _, frameRGB := o.ScreenRGB()
		for _, c := range frameRGB {
			if c != 0 {
				visibleFrames++
				break
			}
		}
		t.Logf("畫面診斷：非零色號像素=%d 非黑DAC色數=%d", litIndices, paletteColors)
		if dir := os.Getenv("KOL_ORACLE_CAPTURE"); dir != "" {
			rw, rh, rgb := o.ScreenRGB()
			if rw > 0 && rh > 0 && len(rgb) == rw*rh*3 {
				im := image.NewRGBA(image.Rect(0, 0, rw, rh))
				for y := 0; y < rh; y++ {
					for x := 0; x < rw; x++ {
						n := (y*rw + x) * 3
						im.SetRGBA(x, y, color.RGBA{rgb[n], rgb[n+1], rgb[n+2], 255})
					}
				}
				f, e := os.Create(filepath.Join(dir, fmt.Sprintf("normal-%02d.png", i)))
				if e != nil {
					t.Fatal(e)
				}
				e = png.Encode(f, im)
				closeErr := f.Close()
				if e != nil {
					t.Fatal(e)
				}
				if closeErr != nil {
					t.Fatal(closeErr)
				}
			}
		}
		t.Logf("步驟=%d 按鍵=%q 步數=%d IP=%s 模式=%x 尺寸=%dx%d 畫面SHA256=%x 已消耗鍵=%d 待讀鍵=%d 主控台=%q 已開檔=%v 缺檔=%v 未實作=%v 錯誤=%v", i, key, o.Steps(), o.IP(), o.VideoMode(), w, h, sha256.Sum256(px), o.KeysConsumed(), o.KeysPending(), o.Console(), o.Opened(), o.Missing(), o.Unimplemented(), err)
		if err != nil {
			t.Fatal(err)
		}
	}
	if visibleFrames == 0 {
		t.Fatal("原版畫面驗收未通過：所有 RGB 擷取均全黑；不可把可執行宣稱為正常顯示")
	}
}
