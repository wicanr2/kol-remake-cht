//go:build oracle

package oracle

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dosgolem/oracle"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// 等原版片頭推進；以 DOS 主控台 Enter 回應等待，不跳入口或修改記憶體。
func TestNormalBookEntry(t *testing.T) {
	root := os.Getenv("KOL_ORACLE_ROOT")
	if root == "" {
		t.Skip("需要玩家自備原版 KOL_ORACLE_ROOT")
	}
	down := 0
	if raw := os.Getenv("KOL_BOOK_SELECT_DOWN"); raw != "" {
		var err error
		down, err = strconv.Atoi(raw)
		if err != nil || down < 0 || down > 15 {
			t.Fatal("KOL_BOOK_SELECT_DOWN 必須是 0 至 15")
		}
	}
	o, err := oracle.Load(filepath.Join(root, "KOL.EXE"), root)
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()
	defer captureBook(t, o, "book-final.png")
	for i := 0; i < 20; i++ {
		if i >= 4 {
			o.Type("\r")
		}
		err = o.Run(50_000_000)
		t.Logf("區段=%d 本段送入DOS-Enter=%t 步數=%d IP=%s 模式=%x 主控台=%q 缺檔=%v 未實作=%v 已開=%v 錯誤=%v", i, i >= 4, o.Steps(), o.IP(), o.VideoMode(), o.Console(), o.Missing(), o.Unimplemented(), o.Opened(), err)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range o.Opened() {
			if strings.EqualFold(filepath.Base(name), "KOLBOOK.EXE") {
				if err = o.Run(5_000_000); err != nil {
					t.Fatal(err)
				}
				t.Log("已由正常啟動鏈載入 KOLBOOK.EXE；圖面須另行檢視")
				captureBook(t, o, "book-cover.png")
				for i, key := range []string{" ", "6", "\r", " "} {
					if i == 2 {
						for n := 0; n < down; n++ {
							if err = o.SendKeys("Down"); err != nil {
								t.Fatal(err)
							}
							if err = o.Run(5_000_000); err != nil {
								t.Fatal(err)
							}
						}
						t.Logf("正常角色清單向下移動 %d 次", down)
						captureBook(t, o, "book-selection.png")
					}
					channel := "BIOS"
					if os.Getenv("KOL_BOOK_DOS") == "1" {
						channel = "DOS"
						o.Type(key)
						t.Logf("書本第 %d 步改用 DOS 主控台輸入", i)
					} else if key == "6" && os.Getenv("KOL_BOOK_MOUSE") == "1" {
						channel = "滑鼠"
						err = o.Click(237, 158, oracle.NoCursorWait())
						t.Log("目錄第六列改用滑鼠點擊 (237,158)")
					} else if key == "\r" {
						err = o.SendKeys("Return")
					} else if key == " " {
						err = o.SendKeys("Space")
					} else {
						err = o.TypeKeys(key)
					}
					if err != nil {
						t.Fatal(err)
					}
					// 2,000 萬指令時翻頁尚未完成；先讓既有動畫完整推進再送下一鍵。
					if err = o.Run(80_000_000); err != nil {
						t.Fatal(err)
					}
					t.Logf("書本步驟=%d 介面=%s 按鍵=%q 已讀=%d 待讀=%d IP=%s 主控台=%q 已開=%v 缺檔=%v", i, channel, key, o.KeysConsumed(), o.KeysPending(), o.IP(), o.Console(), o.Opened(), o.Missing())
					captureBook(t, o, fmt.Sprintf("book-action-%d.png", i))
				}
				for _, name := range o.Opened() {
					if strings.EqualFold(filepath.Base(name), "KOLTOWN.EXE") {
						if os.Getenv("KOL_BOOK_GOODBYE_SINGLE") == "1" {
							o.SyncDGroupFromDS()
							if count := o.Word(o.DS(0x5A)); count != 1 {
								t.Fatalf("單人離店前隊員數不是一：%d", count)
							}
							if err = o.SendKeys("Return"); err != nil {
								t.Fatal(err)
							}
							if err = o.Run(20_000_000); err != nil {
								t.Fatal(err)
							}
							captureBook(t, o, "single-goodbye-outside.png")
							return
						}
						if os.Getenv("KOL_BOOK_BATTLE") == "1" {
							probeNormalBattlePath(t, o)
						}
						return
					}
				}
				t.Fatal("已達書本，但正常 SPACE/6/Enter/SPACE 尚未載入 KOLTOWN.EXE")
				return
			}
		}
	}
	t.Fatal("十億指令內未由正常啟動鏈載入 KOLBOOK.EXE")
}

// 從已到達的旅店繼續正常操作；只固定測試亂數，不注入隊員、座標或遭遇。
func probeNormalBattlePath(t *testing.T, o *oracle.Oracle) {
	t.Helper()
	// RE-132：KOLTOWN DGROUP 的 +0x1600／+0x1602 是 32-bit seed。
	// 原始 IDA 線性位址 0x2F4D0／0x2F4D2，DGROUP 基底 0x2DED0。
	o.SyncDGroupFromDS()
	o.WriteU16(o.DS(0x1600), 1)
	o.WriteU16(o.DS(0x1602), 0)
	t.Logf("原版測試 seed=1，設定點=旅店招呼後，DGROUP=%04X；尚未宣稱與 remake 戰鬥同狀態", o.DGroupSeg())
	// 送鍵之前先留一張。點擊落空時「座標不對」與「那個位置本來就沒東西」在
	// 畫面上一模一樣（`oracle/input.go` 的 Button 註解），有這張才分得開。
	captureBook(t, o, "battle-path-before.png")
	if steps := o.MousePollSteps(); len(steps) > 0 {
		t.Logf("送鍵前遊戲已輪詢滑鼠 %d 次，最後一次在第 %d 道指令", len(steps), steps[len(steps)-1])
	} else {
		t.Log("送鍵前遊戲一次都沒讀過滑鼠——這個畫面可能只收鍵盤")
	}
	// 觀察模式：這個畫面收不收滑鼠、要按什麼才進得了旅店，先看清楚再送固定序列。
	// 2026-09-10 量到的事實：進場後滑鼠輪詢數停在 426,415，點擊期間**一次都沒有
	// 增加**——依 dosgolem `docs/spec/012` 的三分法，那是「按住的期間遊戲沒讀
	// 滑鼠」，不是座標不對。畫面停在 Halfway House 的建築外觀，還沒進門。
	if os.Getenv("KOL_BOOK_PROBE_ENTRY") == "1" {
		before := len(o.MousePollSteps())
		// 已知的兩步：`Down` 叫出圖示列（輪詢從 +43 跳到 +305,850），`Return` 進門
		// （Trollsbane Inn 的敘述頁）。後面繼續觀察怎麼走到招募。
		for i, key := range []string{"Down", "Return", "Space", "Right", "Return", "Space", "Return", "Down"} {
			if err := o.SendKeys(key); err != nil {
				t.Fatalf("送 %q：%v", key, err)
			}
			if err := o.Run(20_000_000); err != nil {
				t.Fatalf("跑 %q：%v", key, err)
			}
			captureBook(t, o, fmt.Sprintf("entry-%02d-%s.png", i, key))
			now := len(o.MousePollSteps())
			t.Logf("進門探測 step=%d key=%-7s IP=%s 滑鼠輪詢 +%d（累計 %d）",
				i, key, o.IP(), now-before, now)
			before = now
		}
		t.Log("進門探測結束；只是觀察，不宣稱任何一步是正確的進門方式")
		return
	}

	// 進旅店的三步。少了它們，後面那一下點擊會落在**建築外觀**上——圖示列還沒
	// 出現，遊戲在那個畫面根本不讀滑鼠（2026-09-10 量到：點擊前後滑鼠輪詢數都是
	// 426,415，一次都沒增加）。座標本身沒有問題：`town.IconBar` 是 X=32、Step=32，
	// 第 6 格 COMPANION 的中心正好是 (204, 181)，也就是下面點的那個位置。
	//
	//	Down   叫出圖示列（輪詢從 +43 跳到 +305,850）
	//	Return 進門，出現旅店敘述頁
	//	Space  店主招呼，六個圖示（DOOR／MIRROR／LISTEN／TALK／REST／COMPANION）出現
	for i, key := range []string{"Down", "Return", "Space",
		"recruit-click", "Return", "Return", "Return", "Return", "Space"} {
		var err error
		if key == "recruit-click" {
			err = o.Click(202, 182, oracle.NoCursorWait(), oracle.Frame(func(o *oracle.Oracle) []uint8 {
				_, _, rgb := o.ScreenRGB()
				return rgb
			}))
		} else if key == "Return" || key == "Space" {
			err = o.SendKeys(key)
		} else {
			err = o.TypeKeys(key)
		}
		if err != nil {
			t.Fatal(err)
		}
		if err = o.Run(20_000_000); err != nil {
			t.Fatal(err)
		}
		captureBook(t, o, fmt.Sprintf("battle-path-%02d.png", i))
		t.Logf("正常旅店延伸 step=%d key=%q IP=%s 已讀=%d 待讀=%d", i, key, o.IP(), o.KeysConsumed(), o.KeysPending())
	}
	if os.Getenv("KOL_BOOK_WALK") != "1" {
		return
	}
	// ⚠ **下面這串方向鍵是綁城鎮的**：它是照 Halfway House 的地圖數出來的。
	// 換一個起始角色就換一座城，同一串鍵走不出去——2026-09-10 用
	// `KOL_BOOK_SELECT_DOWN=12`（Mula，在 Trollsbane）跑，一百步之後
	// `fight-entry.png` 還是城鎮外觀，`exportNormalBattleState` 因此以
	// 「參戰上限不一致：count=0 limit=1」失敗。那不是遭遇沒觸發，是**根本沒出城**。
	keys := []string{"Right", "Right"}
	for n := 0; n < 9; n++ {
		keys = append(keys, "Up")
	}
	keys = append(keys, "Return")
	for n := 0; n < 100; n++ {
		keys = append(keys, "Right")
	}
	for i, key := range keys {
		if err := o.SendKeys(key); err != nil {
			t.Fatal(err)
		}
		if err := o.Run(2_000_000); err != nil {
			t.Fatal(err)
		}
		if i < 12 || (i-11)%10 == 0 {
			captureBook(t, o, fmt.Sprintf("walk-%03d.png", i))
		}
	}
	captureBook(t, o, "walk-final.png")
	t.Log("正常方向鍵步行探測完成；是否出城／遇敵須檢視畫格，不代表戰鬥同狀態通過")
	if os.Getenv("KOL_BOOK_NEXT_DAY") == "1" {
		if err := o.SendKeys("Space"); err != nil {
			t.Fatal(err)
		}
		if err := o.Run(20_000_000); err != nil {
			t.Fatal(err)
		}
		captureBook(t, o, "next-day-start.png")
		for i := 0; i < 100; i++ {
			if err := o.SendKeys("Right"); err != nil {
				t.Fatal(err)
			}
			if err := o.Run(2_000_000); err != nil {
				t.Fatal(err)
			}
			if i%10 == 9 {
				captureBook(t, o, fmt.Sprintf("next-day-%03d.png", i))
			}
		}
		if os.Getenv("KOL_BOOK_ENTER_FIGHT") == "1" {
			if err := o.SendKeys("Space"); err != nil {
				t.Fatal(err)
			}
			if err := o.Run(20_000_000); err != nil {
				t.Fatal(err)
			}
			captureBook(t, o, "fight-entry.png")
			exportNormalBattleState(t, o)
			t.Log("已確認伏擊訊息；只擷取戰鬥入口，不送戰鬥結算命令")
			if os.Getenv("KOL_BOOK_PLAN") == "1" {
				probeNormalPlanning(t, o)
			}
		}
	}
}

// 僅匯出已證實的欄位；未解欄位留在原始記錄，不猜測語意。
func exportNormalBattleState(t *testing.T, o *oracle.Oracle) {
	t.Helper()
	dir := os.Getenv("KOL_ORACLE_CAPTURE")
	if dir == "" {
		t.Fatal("戰鬥狀態匯出需要 KOL_ORACLE_CAPTURE")
	}
	type actor struct {
		Slot   int    `json:"slot"`
		Name   string `json:"name"`
		Record string `json:"record_hex"`
		Side   int    `json:"side"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
	}
	var actors []actor
	for i := 0; i < 18; i++ {
		raw := make([]byte, 50)
		for j := range raw {
			raw[j] = o.Byte(o.DS(uint16(0x6712 + i*50 + j)))
		}
		if raw[1] > 2 {
			t.Fatalf("參戰者 %d 的陣營非法：%d", i, raw[1])
		}
		name := ""
		if i < 6 && raw[1] == 0 {
			for j := 0; j < 16; j++ {
				b := o.Byte(o.DS(uint16(0xBB88 + i*0x300 + j)))
				if b == 0 {
					break
				}
				name += string([]byte{b})
			}
		}
		actors = append(actors, actor{i, name, hex.EncodeToString(raw), int(raw[1]), int(raw[0x1F]), int(raw[0x20])})
	}
	current := int(o.Word(o.DS(0x9E5A)))
	// SPEC-020 原版上界：DS:2D9C 是敵人數；DS:1CF0 是最後有效槽位。
	count := int(o.Word(o.DS(0x2D9C)))
	limit := int(o.Word(o.DS(0x1CF0))) + 1
	if count < 1 || count > 12 || limit != count+6 {
		t.Fatalf("參戰上限不一致：count=%d limit=%d", count, limit)
	}
	if current < 0 || current >= 6 || actors[current].Side != 0 {
		t.Fatalf("目前規劃角色不是隊員：%d", current)
	}
	seedBefore := uint32(o.Word(o.DS(0x1600))) | uint32(o.Word(o.DS(0x1602)))<<16
	// 在共同規劃比較點重新固定兩方測試 seed；不改動原版角色或站位。
	o.WriteU16(o.DS(0x1600), 1)
	o.WriteU16(o.DS(0x1602), 0)
	data := map[string]any{"schema": "kol-battle-alignment-v1", "source": "dosgolem-normal-player-path",
		"enemy_count": count, "active_slot_limit": limit,
		"dgroup_segment": o.DGroupSeg(), "actor_ds_base": 0x6712, "actor_stride": 50,
		"evidence":      "RE-080、RE-132；DS 相對位址，原始記錄與已證實欄位並列",
		"current_actor": current, "seed_before_alignment": seedBefore, "alignment_seed": 1,
		"seed_point": "正常戰鬥第一名隊員規劃入口，畫格擷取後", "actors": actors,
		"limitations": []string{"只對齊角色、站位及後續測試 seed；地形、完整戰鬥狀態與結算尚未對拍"}}
	body, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "battle-alignment.json"), append(body, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	// 只擴充唯讀研究收據，不建立新的正式存檔格式。
	// RE-080／129：角色 DS:BB88、步幅 0x300；怪物 DS:CD92、步幅 0x80；
	// 耐力 DS:1C50、疲勞 DS:322E，後兩者各為每槽一個 word。
	readHex := func(base, size int) string {
		raw := make([]byte, size)
		for i := range raw {
			raw[i] = o.Byte(o.DS(uint16(base + i)))
		}
		return hex.EncodeToString(raw)
	}
	var characters, monsters []map[string]any
	for i := 0; i < 6; i++ {
		if actors[i].Side == 0 {
			characters = append(characters, map[string]any{"slot": i,
				"ds_offset": 0xBB88 + i*0x300, "record_hex": readHex(0xBB88+i*0x300, 0x300)})
		}
	}
	for i := 0; i < count; i++ {
		monsters = append(monsters, map[string]any{"slot": i + 6,
			"ds_offset": 0xCD92 + i*0x80, "record_hex": readHex(0xCD92+i*0x80, 0x80)})
	}
	var counters []map[string]any
	for i := 0; i < limit; i++ {
		if actors[i].Side != 2 {
			counters = append(counters, map[string]any{"slot": i,
				"stamina_ds_offset": 0x1C50 + i*2, "stamina": o.Word(o.DS(uint16(0x1C50 + i*2))),
				"fatigue_ds_offset": 0x322E + i*2, "fatigue": o.Word(o.DS(uint16(0x322E + i*2)))})
		}
	}
	records := map[string]any{"schema": "kol-battle-state-records-v1",
		"source": "dosgolem-normal-player-path", "dgroup_segment": o.DGroupSeg(),
		"input": "KOLTOWN-unpacked.EXE", "input_sha256": "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2",
		"address_space": "執行期 DGROUP 相對位移；不是 IDA 線性位址",
		"evidence":      "RE-080、RE-129；記錄基址及耐力／疲勞欄位已證實；原始記錄未解 byte 不附推測語意",
		"characters":    characters, "monsters": monsters, "counters": counters,
		"alignment_seed": 1, "monster_level_ds_offset": 0xD39A,
		"monster_level": o.Word(o.DS(0xD39A)),
		"limitations":   []string{"原版正常入口唯讀記錄，尚未與 remake 完整狀態比較", "不包含地形或完整規劃狀態"}}
	body, err = json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "battle-state-records.json"), append(body, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	// sub_223C0:2254E 直接以 word_30C8A 為 ES；2B9A/2B9C 是可重用暫存指標。
	fieldOff, fieldSeg := uint16(0), o.Word(o.DS(0x2DBA))
	if fieldSeg == 0 || uint32(fieldOff)+4096 > 65536 {
		t.Fatal("戰場遠指標無效或超過段界")
	}
	rawField := make([]byte, 4096)
	for i := range rawField {
		rawField[i] = o.Byte(oracle.Addr{Seg: fieldSeg, Off: fieldOff + uint16(i)})
		if rawField[i] != 9 && rawField[i] != 10 && rawField[i] != 13 {
			t.Fatalf("正常野外戰場格 %d 出現非預期地形 %d", i, rawField[i])
		}
	}
	terrain := map[string]any{
		"schema": "kol-battle-terrain-receipt-v1", "source": "dosgolem-normal-player-path",
		"input": "KOLTOWN-unpacked.EXE", "input_sha256": records["input_sha256"],
		"address_space":  "執行期段位址；指標欄位為 DGROUP 相對位移",
		"dgroup_segment": o.DGroupSeg(), "pointer_segment_ds": 0x2DBA, "offset_contract": "繪製直接使用 ES:SI，SI 為 Y*64+X，無額外基底位移",
		"buffer_segment": fieldSeg, "buffer_offset": fieldOff,
		"width": 64, "height": 64, "index": "Y*64+X", "record_hex": hex.EncodeToString(rawField),
		"confidence": "已證實", "evidence": "RE-080；docs/re/20260908-battle-field-index.md",
		"alignment_seed": 1, "current_actor": current,
		"limitations": []string{"正常原版地形快照；不是完整戰鬥結算驗收"},
	}
	body, err = json.MarshalIndent(terrain, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "battle-terrain.json"), append(body, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
}

func captureBook(t *testing.T, o *oracle.Oracle, name string) {
	t.Helper()
	dir := os.Getenv("KOL_ORACLE_CAPTURE")
	if dir == "" {
		return
	}
	w, h, rgb := o.ScreenRGB()
	im := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			n := (y*w + x) * 3
			im.SetRGBA(x, y, color.RGBA{rgb[n], rgb[n+1], rgb[n+2], 255})
		}
	}
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Error(err)
		return
	}
	if err = png.Encode(f, im); err != nil {
		t.Error(err)
	}
	if err = f.Close(); err != nil {
		t.Error(err)
	}
}

// probeNormalPlanning 在規劃入口逐鍵觀察原版把什麼寫回參戰記錄。
//
// 規劃段每一名挑一個指令，寫回的是三格（SPEC-010）：`+0x22` 決定完成、
// `+0x23` 攻擊選項、`+0x25` 指令碼。整輪挑完停在記錄 0x108「Begin next round?」。
//
// 這一支**不預設按鍵序列會走到哪**——原版的圖示列停在哪一格、哪些鍵有作用都還沒
// 逐項確認過，所以先逐鍵記錄「按下去之後那三格變成什麼」，把畫面也存下來。
// 有了這份觀察才談得上比對；先寫死一組鍵再宣稱對拍，等於拿猜的輸入去比。
func probeNormalPlanning(t *testing.T, o *oracle.Oracle) {
	t.Helper()
	dir := os.Getenv("KOL_ORACLE_CAPTURE")
	if dir == "" {
		t.Fatal("規劃探測需要 KOL_ORACLE_CAPTURE")
	}
	const (
		actorBase   = 0x6712
		actorStride = 50
		plannedOff  = 0x22
		optionOff   = 0x23
		commandOff  = 0x25
		currentSlot = 0x9E5A
	)
	type snap struct {
		Step    int    `json:"step"`
		Key     string `json:"key"`
		Current int    `json:"current_actor"`
		Planned int    `json:"planned_flag"`
		Option  int    `json:"attack_option"`
		Command int    `json:"command_code"`
		IP      string `json:"ip"`
		Frame   string `json:"frame,omitempty"`
	}
	read := func(step int, key string) snap {
		cur := int(o.Word(o.DS(currentSlot)))
		rec := uint16(actorBase + actorStride*cur)
		return snap{Step: step, Key: key, Current: cur,
			Planned: int(o.Byte(o.DS(rec + plannedOff))),
			Option:  int(o.Byte(o.DS(rec + optionOff))),
			Command: int(o.Byte(o.DS(rec + commandOff))),
			IP:      o.IP().String()}
	}

	var out []snap
	first := read(-1, "(進場後，未送鍵)")
	out = append(out, first)
	t.Logf("規劃入口：目前第 %d 名，+0x22=%d +0x23=%d +0x25=%#02x",
		first.Current, first.Planned, first.Option, first.Command)

	// 送鍵的順序刻意保守：先看單獨一個 Return 會不會就決定一個指令，
	// 再看方向鍵與 `<`／`>` 有沒有搬動選擇。每一步都留畫面。
	for i, key := range []string{"Return", "Right", "Return", "period", "Return", "Return"} {
		if err := o.SendKeys(key); err != nil {
			t.Fatalf("送 %q：%v", key, err)
		}
		if err := o.Run(20_000_000); err != nil {
			t.Fatalf("跑 %q：%v", key, err)
		}
		name := fmt.Sprintf("plan-%02d.png", i)
		captureBook(t, o, name)
		st := read(i, key)
		st.Frame = name
		out = append(out, st)
		t.Logf("規劃 step=%d key=%-8s 目前第 %d 名 +0x22=%d +0x23=%d +0x25=%#02x IP=%s",
			i, key, st.Current, st.Planned, st.Option, st.Command, st.IP)
	}

	data := map[string]any{
		"schema": "kol-battle-planning-probe-v1", "source": "dosgolem-normal-player-path",
		"alignment_seed": 1,
		"evidence":       "SPEC-010 的 +0x22／+0x23／+0x25；DS 相對位址",
		"actor_ds_base":  actorBase, "actor_stride": actorStride,
		"dgroup_segment": o.DGroupSeg(),
		"steps":          out,
		"limitations": []string{
			"這是逐鍵觀察，不是對拍：按鍵序列本身還沒與 remake 對齊，" +
				"也還沒宣稱任何一步的結果與 remake 相同",
		},
	}
	body, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "battle-planning.json"), append(body, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	t.Log("規劃逐鍵觀察已匯出；尚未宣稱與 remake 規劃同狀態")
}
