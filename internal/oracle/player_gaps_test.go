//go:build oracle

package oracle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

// 受控載入測試：只替代 archive 解包，原版 RLE 與轉置均執行，不是正常競技場路徑。
func TestArenaMapOriginal(t *testing.T) {
	path := os.Getenv("KOL_ARENA_PROBE")
	if path == "" {
		t.Skip("需要本機競技場地圖研究輸入")
	}
	var fixture struct {
		Payload int
		Raw     string `json:"raw_hex"`
		SHA     string `json:"raw_sha256"`
	}
	b, err := os.ReadFile(path + "/arena-map.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	raw, err := hex.DecodeString(fixture.Raw)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if fixture.Payload != 0x55 || hex.EncodeToString(sum[:]) != fixture.SHA {
		t.Fatal("地圖來源驗證失敗")
	}
	f, err := os.Open(path + "/arena-map.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	field, err := battle.LoadField(img)
	if err != nil {
		t.Fatal(err)
	}
	o := prepare(t)
	o.WriteU16(o.IDA(0x2FB5C), 0x9000)
	o.WriteU16(o.IDA(0x30C8A), 0x9200)
	calls := 0
	o.Stub(o.IDA(0x1A4FE), func(p *oracle.Oracle) uint32 {
		calls++
		if p.Arg(1) != 0x55 || p.Arg(2) != 0 || p.Arg(3) != 0x9000 {
			t.Fatalf("地圖載入參數不符：%x %x %x", p.Arg(1), p.Arg(2), p.Arg(3))
		}
		p.WriteBytes(oracle.Addr{Seg: p.Arg(3), Off: p.Arg(2)}, raw)
		return 0
	})
	if _, err = o.CallBudget(2_000_000, o.IDA(0x2B3D1), 24); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("載入次數=%d", calls)
	}
	for x := 0; x < 64; x++ {
		for y := 0; y < 64; y++ {
			got := int(o.Byte(oracle.Addr{Seg: 0x9200, Off: uint16(y*64 + x)}))
			if got != field.At(x, y) {
				t.Fatalf("原版轉置格(%d,%d)=%d remake=%d", x, y, got, field.At(x, y))
			}
		}
	}
	t.Log("競技場 payload 85、RLE 與轉置：4096 格與 PNG 載入一致")
}

// 查核既有 PNG 的合成假說；失敗表示 runtime 模型需訂正，不改原版來配合。
func TestWeatherCompositeOriginal(t *testing.T) {
	root := os.Getenv("KOL_WEATHER_ROOT")
	if root == "" {
		t.Skip("需要已驗證天候 PNG")
	}
	lib, err := assets.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	palette := [16]color.RGBA{{0, 0, 0, 255}, {0, 0, 170, 255}, {0, 170, 0, 255}, {0, 170, 170, 255}, {170, 0, 0, 255}, {170, 0, 170, 255}, {170, 85, 0, 255}, {170, 170, 170, 255}, {85, 85, 85, 255}, {85, 85, 255, 255}, {85, 255, 85, 255}, {85, 255, 255, 255}, {255, 85, 85, 255}, {255, 85, 255, 255}, {255, 255, 85, 255}, {255, 255, 255, 255}}
	o := prepare(t)
	o.WriteU16(o.IDA(0x34A1A), 0x9000)
	o.WriteU16(o.IDA(0x2DF1E), 0x0D)
	for _, group := range []int{0, 1, 2, 5} {
		a, _, err := lib.Decode("WEATPICS", group*3)
		if err != nil {
			t.Fatal(err)
		}
		b, _, err := lib.Decode("WEATPICS", group*3+1)
		if err != nil {
			t.Fatal(err)
		}
		c, _, err := lib.Decode("WEATPICS", group*3+2)
		if err != nil {
			t.Fatal(err)
		}
		weather, err := world.NewWeather(a, b, c)
		if err != nil {
			t.Fatal(err)
		}
		raw := make([]byte, 0x2000) // 包含結果尾端 0x1170+0x7C0=0x1930。
		for y := 0; y < 62; y++ {
			for x := 0; x < 64; x++ {
				for layer, img := range []interface{ At(int, int) color.Color }{a, b} {
					r, g, bb, _ := img.At(x, y).RGBA()
					found := -1
					for i, p := range palette {
						if uint8(r>>8) == p.R && uint8(g>>8) == p.G && uint8(bb>>8) == p.B {
							found = i
							break
						}
					}
					if found < 0 {
						t.Fatal("PNG 不在 EGA 色盤")
					}
					for plane := 0; plane < 4; plane++ {
						if found&(1<<plane) != 0 {
							raw[layer*0x7C0+plane*496+y*8+x/8] |= 0x80 >> uint(x%8)
						}
					}
				}
				r, _, _, _ := c.At(x, y).RGBA()
				if r != 0 {
					raw[0xF80+y*8+x/8] |= 0x80 >> uint(x%8)
				}
			}
		}
		// sub_2A0C6:0x2A173…0x2A185 先將完整底圖複製至結果緩衝區。
		copy(raw[0x1170:], raw[:0x7C0])
		o.WriteBytes(oracle.Addr{Seg: 0x9000}, raw)
		if _, err = o.CallBudget(2_000_000, o.IDA(0x2A18E)); err != nil {
			t.Fatal(err)
		}
		got := o.Bytes(oracle.Addr{Seg: 0x9000, Off: 0x1170}, 1984)
		want := weather.Composite(1)
		bad := 0
		for y := 0; y < 62; y++ {
			for x := 0; x < 64; x++ {
				index := 0
				for plane := 0; plane < 4; plane++ {
					if got[plane*496+y*8+x/8]&(0x80>>uint(x%8)) != 0 {
						index |= 1 << plane
					}
				}
				if palette[index] != want[y*64+x] {
					if bad < 3 {
						src := y*128 + (x-1+128)%128
						t.Logf("組%d (%d,%d) 原版%v 重製%v 底圖%v 覆蓋%v 遮罩%v", group, x, y, palette[index], want[y*64+x], a.At(x, y), b.At(src%64, src/64), c.At(src%64, src/64))
					}
					bad++
				}
			}
		}
		if bad != 0 {
			t.Errorf("組%d：既有 64×62 合成模型與原版差 %d 像素", group, bad)
		}
	}
}

// 只觀測繪製／等待呼叫參數；不把它當作像素或硬體 wall-clock 對拍。
func TestWeatherDispatchOriginal(t *testing.T) {
	o := prepare(t)
	for _, ea := range []uint32{0x2DF60, 0x2DF2C, 0x2DF48, 0x2DF28} {
		o.WriteU16(o.IDA(ea), 0)
	}
	o.WriteU16(o.IDA(0x2DF44), 1)
	o.WriteU16(o.IDA(0x34A1A), 0x9000)
	var draws []uint16
	cloud, waits := 0, 0
	o.Stub(o.IDA(0x19E9D), func(p *oracle.Oracle) uint32 { draws = append(draws, p.Arg(0)); return 0 })
	o.Stub(o.IDA(0x2A18E), func(*oracle.Oracle) uint32 { cloud++; return 0 })
	o.Stub(o.IDA(0x1A8DF), func(p *oracle.Oracle) uint32 {
		if p.Arg(0) != 1 {
			t.Fatal("等待不是一個 BIOS tick")
		}
		waits++
		return 0
	})
	for group := 0; group < 6; group++ {
		o.WriteU16(o.IDA(0x2DF5A), uint16(group))
		o.WriteU16(o.IDA(0x2DF5C), 0)
		o.WriteU16(o.IDA(0x2DF28), 0)
		draws = nil
		cloud = 0
		waits = 0
		for step := 0; step < 12; step++ {
			if _, err := o.CallBudget(100_000, o.IDA(0x11390)); err != nil {
				t.Fatal(err)
			}
		}
		if len(draws) != 4 || waits != 8 {
			t.Fatalf("組%d 繪製%d 等待%d", group, len(draws), waits)
		}
		for i, seg := range draws {
			want := uint16(0x9117)
			if group == 3 || group == 4 {
				want = uint16(0x9000 + 0x7C*(i%3))
			}
			if seg != want {
				t.Fatalf("組%d 畫格%d 段=%x 預期%x", group, i, seg, want)
			}
		}
		wantCloud := 4
		if group == 3 || group == 4 {
			wantCloud = 0
		}
		if cloud != wantCloud {
			t.Fatalf("組%d 雲層次數%d", group, cloud)
		}
	}
	t.Log("六組天候：雨雪三畫格、其他捲雲，每次更新間兩個 BIOS tick")
}
