package font

import "testing"

func TestAlternativesPreserveMetricsASCIIAndFallback(t *testing.T) {
	f, err := Load("../../assets/ui-font", "zh-hant-16")
	if err != nil {
		t.Fatal(err)
	}
	normal, err := Load("../../cmd/kol/roundedfont", "rounded-16")
	if err != nil {
		t.Fatal(err)
	}
	large, err := Load("../../cmd/kol/roundedfont", "rounded-32")
	if err != nil {
		t.Fatal(err)
	}
	text := "生命值 金幣 9000，攻擊上段。"
	w, h := f.Measure(text)
	original, _ := f.glyphForDraw('金')
	ascii, _ := f.glyphForDraw('9')
	fallback, _ := f.glyphForDraw('\u575f') // 粉圓2.1缺此字；原字表保留。
	if err := f.SetAlternatives(normal, large); err != nil {
		t.Fatal(err)
	}
	if actualW, actualH := f.Measure(text); actualW != w || actualH != h {
		t.Fatal("字距或行高改變")
	}
	if alt, ok := f.glyphForDraw('金'); !ok || alt == original {
		t.Fatal("中文字未替換")
	}
	if same, _ := f.glyphForDraw('9'); same != ascii {
		t.Fatal("英數字被替換")
	}
	if same, _ := f.glyphForDraw('\u575f'); same != fallback {
		t.Fatal("粉圓缺字未沿原字模")
	}
	bad := *large
	bad.mapped = make(map[rune]glyphEntry)
	for ch, glyph := range large.mapped {
		bad.mapped[ch] = glyph
	}
	glyph := bad.mapped['金']
	glyph.Advance++
	bad.mapped['金'] = glyph
	if err := f.SetAlternatives(normal, &bad); err == nil {
		t.Fatal("錯字距未拒絕")
	}
	if f.alternativeLarge != large {
		t.Fatal("失敗後破壞目前字型")
	}
	if err := f.SetAlternatives(nil, nil); err != nil {
		t.Fatal(err)
	}
	if restored, _ := f.glyphForDraw('金'); restored != original {
		t.Fatal("未恢復原字模")
	}
}
