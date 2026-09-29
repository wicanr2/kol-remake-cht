package bitmapedit

import "testing"

func pattern() Bitmap {
	var b Bitmap
	for y := 0; y < Height; y++ {
		for x := 0; x < Width; x++ {
			b.Set(x, y, (x+y*3)%16)
		}
	}
	return b
}

// 進場狀態照 `sub_1293E` 開頭：工具列模式、顏色 3、白框第 0 格、游標 (20, 20)。
func TestNewEditor(t *testing.T) {
	e := New(pattern(), nil)
	if !e.ToolsMode || e.Color != 3 || e.Tool != 0 || e.X != 20 || e.Y != 20 || e.Modified || e.Pen {
		t.Fatalf("進場狀態不對：%+v", e)
	}
}

// 高 nibble 在左。
func TestNibbleOrder(t *testing.T) {
	var b Bitmap
	b.Set(0, 0, 0xA)
	b.Set(1, 0, 0x5)
	if b[0] != 0xA5 || b.At(0, 0) != 0xA || b.At(1, 0) != 0x5 {
		t.Fatalf("第 0 byte 是 %#x", b[0])
	}
}

// 四支變換各做兩次或四十次要回到原圖，而且做一次要真的改變。
func TestTransformsRoundTrip(t *testing.T) {
	orig := pattern()
	for _, c := range []struct {
		name  string
		fn    func(*Bitmap)
		times int
	}{
		{"右移", ShiftRight, Width}, {"上移", ShiftUp, Height},
		{"左右鏡射", MirrorH, 2}, {"上下鏡射", MirrorV, 2},
	} {
		b := orig
		c.fn(&b)
		if b == orig {
			t.Fatalf("%s 做一次沒有改變", c.name)
		}
		for i := 1; i < c.times; i++ {
			c.fn(&b)
		}
		if b != orig {
			t.Fatalf("%s 做 %d 次沒有回到原圖", c.name, c.times)
		}
	}
	b := orig
	ShiftRight(&b)
	if b.At(0, 5) != orig.At(Width-1, 5) || b.At(1, 5) != orig.At(0, 5) {
		t.Fatal("右移：最右一格應繞到最左")
	}
	b = orig
	ShiftUp(&b)
	if b.At(3, Height-1) != orig.At(3, 0) || b.At(3, 0) != orig.At(3, 1) {
		t.Fatal("上移：第 0 列應繞到最底")
	}
	b = orig
	MirrorH(&b)
	if b.At(0, 7) != orig.At(Width-1, 7) || b.At(1, 7) != orig.At(Width-2, 7) {
		t.Fatal("左右鏡射不對")
	}
}

// 工具列：`,`／`.` 移白框並繞回，Enter 執行；方向鍵改目前顏色（上下 ±4），mod 16。
func TestToolsModeKeys(t *testing.T) {
	e := New(pattern(), nil)
	e.Key(',')
	if e.Tool != 7 {
		t.Fatalf("0 往左應繞到 7，得到 %d", e.Tool)
	}
	e.Key('.')
	e.Key(KeyPgDn)
	if e.Tool != 1 {
		t.Fatalf("得到 %d", e.Tool)
	}
	e.Key(KeyUp) // 3 − 4 → 15
	if e.Color != 15 {
		t.Fatalf("顏色 %d", e.Color)
	}
	e.Key(KeyRight) // 16 → 0
	if e.Color != 0 {
		t.Fatalf("顏色 %d", e.Color)
	}
	// 填滿、改過；重載清掉。
	e.Tool = ToolFill
	e.Key(KeyEnter)
	if !e.Modified || e.Pix.At(13, 29) != 0 {
		t.Fatal("填滿沒有生效")
	}
	e.Tool = ToolReload
	e.Key(KeyEnter)
	if e.Modified || e.Pix != e.Orig {
		t.Fatal("重載應還原並清掉已修改")
	}
}

// 沒改過直接離開；改過要問，只認 y／n。
func TestExitAndSavePrompt(t *testing.T) {
	e := New(pattern(), nil)
	e.Key(KeyEnter)
	if !e.Done || e.Save || e.Asking {
		t.Fatal("沒改過按離開應直接結束、不存")
	}
	e = New(pattern(), nil)
	e.Tool = ToolShiftR
	e.Key(KeyEnter)
	e.Tool = ToolExit2
	e.Key(KeyEnter)
	if !e.Asking {
		t.Fatal("改過離開應問存檔")
	}
	e.Key('q')
	if !e.Asking {
		t.Fatal("y／n 以外的鍵應重讀")
	}
	e.Key('Y')
	if !e.Done || !e.Save {
		t.Fatal("答 Y 應存檔離開")
	}
}

// ESC 翻模式。畫布：D 落筆、之後每一鍵先畫目前那一格；X 畫一點並抬筆；兩軸繞回。
func TestCanvasPenAndWrap(t *testing.T) {
	var blank Bitmap
	e := New(blank, nil)
	e.Key(KeyEsc)
	if e.ToolsMode {
		t.Fatal("ESC 應切到畫布模式")
	}
	e.Key('D')
	if e.Pix.At(20, 20) != 3 || !e.Pen || !e.Modified {
		t.Fatal("D 應落筆並畫目前那一格")
	}
	e.Key(KeyRight) // 先畫 (20,20)，再移到 21
	e.Key(KeyRight) // 畫 (21,20)，移到 22
	if e.Pix.At(21, 20) != 3 || e.Pix.At(22, 20) != 0 || e.X != 22 {
		t.Fatalf("落筆移動：(21,20)=%d (22,20)=%d x=%d", e.Pix.At(21, 20), e.Pix.At(22, 20), e.X)
	}
	e.Key('u')
	e.Key(KeyRight) // 抬筆那一輪還是先畫 (22,20)（讀鍵前筆還落著）
	if e.Pix.At(22, 20) != 3 {
		t.Fatal("抬筆那一鍵之前的一格應已畫上")
	}
	e.Key(KeyRight)
	if e.Pix.At(23, 20) != 0 {
		t.Fatal("抬筆之後不該再畫")
	}
	e.X, e.Y = 0, 0
	e.Key(KeyHome)
	if e.X != Width-1 || e.Y != Height-1 {
		t.Fatalf("左上角往左上應繞到右下，得到 (%d,%d)", e.X, e.Y)
	}
	e.Pix.Set(e.X, e.Y, 9)
	e.Key('c')
	if e.Color != 9 {
		t.Fatalf("C 應吸色，得到 %d", e.Color)
	}
	e.Key(KeyEsc)
	if !e.ToolsMode {
		t.Fatal("再按 ESC 應回到工具列模式")
	}
}

// 滑鼠：工具列點未選中的格只移白框，點已選中的等於 Enter；在畫布模式點工具列會先翻模式。
func TestMouse(t *testing.T) {
	var blank Bitmap
	s := &Session{}
	e := New(blank, s)
	e.Click(32+3*32+5, 180) // 第 3 格（填滿）
	if e.Tool != 3 || e.Modified {
		t.Fatal("點未選中的工具格只移白框")
	}
	e.Click(200+24*2+1, 61+22+1) // 色盤第 1 列第 2 格 → 6
	if e.Color != 6 || !e.ToolsMode {
		t.Fatalf("色盤應換色，得到 %d", e.Color)
	}
	e.Click(16+4*5+1, 4+4*7+1) // 畫布 (5,7)：工具列模式先翻到畫布，再畫一點
	if e.ToolsMode || e.X != 5 || e.Y != 7 || e.Pix.At(5, 7) != 6 {
		t.Fatalf("畫布點擊：模式=%v 游標=(%d,%d) 顏色=%d", e.ToolsMode, e.X, e.Y, e.Pix.At(5, 7))
	}
	e.Pix.Set(5, 7, 0)
	e.Click(16+4*5+2, 4+4*7+2) // 同一格不再畫
	if e.Pix.At(5, 7) != 0 {
		t.Fatal("點到與上一次同一格不應再畫")
	}
	e.Click(32+3*32+5, 180) // 畫布模式點已選中的填滿：先翻回工具列，再當 Enter
	if !e.ToolsMode || e.Pix.At(30, 30) != 6 {
		t.Fatal("點已選中的工具格應執行它")
	}
}

// 影像與十六進位字串都能往返；調色盤以外的顏色要失敗。
func TestImageAndHexRoundTrip(t *testing.T) {
	b := pattern()
	back, err := FromImage(b.Image())
	if err != nil || back != b {
		t.Fatalf("影像往返失敗：%v", err)
	}
	h := b.Hex()
	if len(h) != 2*Bytes {
		t.Fatalf("十六進位長度 %d", len(h))
	}
	back, err = ParseHex(h)
	if err != nil || back != b {
		t.Fatalf("十六進位往返失敗：%v", err)
	}
	if _, err = ParseHex("00"); err == nil {
		t.Fatal("長度不對應該失敗")
	}
	img := b.Image()
	img.Pix[0] = 0x12
	if _, err = FromImage(img); err == nil {
		t.Fatal("非 EGA 色應該失敗")
	}
}
