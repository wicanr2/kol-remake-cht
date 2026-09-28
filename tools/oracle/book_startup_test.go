//go:build oracle

package oracle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dosgolem/oracle"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// 等原版片頭推進；以 DOS 主控台 Enter 回應等待，不跳入口或修改記憶體。
// regionOpens 記下每一次開檔的檔名與呼叫當下的 IP。
//
// `region%d.dat` 這個格式字串只在 `KOLBOOK.EXE` 裡（三處：檔案偏移 0x09a02、
// 0x09a48、0x09b30），`KOLTOWN.EXE` 一個都沒有——去開那幾個檔的是**書本**，
// 不是城鎮。這一支要回答的是「`%d` 跑到幾、找不到之後怎麼辦」。
type regionOpens struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	Step uint64 `json:"step"`
}

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
	var opens []regionOpens
	if os.Getenv("KOL_BOOK_REGION_PROBE") == "1" {
		scanned := false
		o.OnFileOpen(func(name string) {
			opens = append(opens, regionOpens{Name: name, IP: o.IP().String(), Step: o.Steps()})
			// 一定要在開檔的當下掃。之後 KOLTOWN.EXE 會取代 KOLBOOK.EXE 佔住同一個段，
			// 兩支又共用同一份 CRT、偏移一樣，事後看記憶體會對到錯的 EXE。
			if scanned || !strings.Contains(strings.ToLower(name), "region") {
				return
			}
			scanned = true
			dumped := false
			// 只比 byte，不要先 ToLower：記憶體裡的非 UTF-8 byte 經 ToLower 會被換成
			// RuneError，長度跟著變，算出來的索引就對不上原始 buffer 了。
			for seg := uint16(0); seg < 0xA000; seg += 0x1000 {
				mem := o.Bytes(oracle.Addr{Seg: seg, Off: 0}, 0x10000-16)
				for at := 0; at+6 <= len(mem); at++ {
					if !bytes.EqualFold(mem[at:at+6], []byte("region")) {
						continue
					}
					end := at + 40
					if end > len(mem) {
						end = len(mem)
					}
					shown := make([]byte, 0, end-at)
					for _, b := range mem[at:end] {
						if b < 32 || b > 126 {
							b = '.'
						}
						shown = append(shown, b)
					}
					t.Logf("模板/檔名 %04x:%04x  %q", seg, at, string(shown))
					if seg == 0x1000 && !dumped {
						dumped = true
						blob := o.Bytes(oracle.Addr{Seg: seg, Off: 0x4400}, 0x0800)
						if out := os.Getenv("KOL_ORACLE_CAPTURE"); out != "" {
							_ = os.WriteFile(filepath.Join(out, "region-installer-text.bin"), blob, 0o644)
						}
						txt := make([]byte, 0, len(blob))
						for _, b := range blob {
							if b < 32 || b > 126 {
								b = '.'
							}
							txt = append(txt, b)
						}
						for i := 0; i < len(txt); i += 72 {
							e := i + 72
							if e > len(txt) {
								e = len(txt)
							}
							t.Logf("安裝器文字 +%04x %s", 0x4400+i, string(txt[i:e]))
						}
					}
				}
			}
		})
		defer func() {
			seen := map[string]bool{}
			for _, r := range opens {
				if !strings.Contains(strings.ToLower(r.Name), "region") {
					continue
				}
				t.Logf("開檔 %-14s IP=%s 第 %d 道指令", r.Name, r.IP, r.Step)
				seen[strings.ToLower(r.Name)] = true
			}
			t.Logf("region 相關開檔共 %d 種：%v；缺檔清單 %v", len(seen), keysOf(seen), o.Missing())
			for _, a := range o.MissingAccesses() {
				if !strings.Contains(strings.ToLower(a.Name), "region") {
					continue
				}
				t.Logf("缺檔 %-14s 呼叫點 %04x:%04x DS=%04x DX=%04x", a.Name, a.CS, a.IP, a.DS, a.DX)
				// BP 鏈在 CRT 的 open() 裡會斷（那層沒建 frame），下面幾筆只作線索不作證據；
				// 模板與檔名的實際位址由上面 OnFileOpen 當下的掃描給出。
				for i, f := range a.Callers {
					if f.CS == 0 && f.IP == 0 {
						break
					}
					t.Logf("    caller[%d] %04x:%04x code=% x args=%v", i, f.CS, f.IP, f.Code, f.Args)
				}
			}
		}()
	}
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
						// 清單要畫完，接下來那個 Enter 才有作用。**不要靠 down 的次數
						// 湊時間**：2026-09-10 量到 `down=2` 時左頁還停在 ORIGIN 封面，
						// Enter 送進去沒有反應，整條路徑因此停在 KOLBOOK
						//（IP 0110:40D7，而畫完的那一條是 0110:40BF）；`down=12`
						// 之所以會過，只是因為十二次各跑 5M 指令、剛好給了六倍的時間。
						// 症狀是「選這個角色進不了遊戲」，看起來完全像角色的問題。
						if err = o.Run(60_000_000); err != nil {
							t.Fatalf("等角色清單畫完：%v", err)
						}
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
					// **不要拿固定指令數當「這一鍵處理完了」**：這一步要跑多久由翻頁動畫
					// 與城鎮載入決定，不是常數。2026-09-11 量到固定 80M 的後果——最後
					// 那個 Space 一直留在待讀（已讀=5／待讀=1），店主招呼沒發生，畫面
					// 停在建築外觀，後面的招募點擊因此落空。改成等鍵真的被讀走。
					if err = o.Run(20_000_000); err != nil {
						t.Fatal(err)
					}
					for n := 0; o.KeysPending() > 0 && n < 20; n++ {
						if err = o.Run(20_000_000); err != nil {
							t.Fatal(err)
						}
					}
					if o.KeysPending() > 0 {
						t.Fatalf("第 %d 步的 %q 送出後 4 億指令仍未被讀走", i, key)
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
	// 觀察模式：拿來看一個沒走通的畫面收不收滑鼠、要按什麼才進得去。
	//
	// 判讀滑鼠輪詢數之前先確認**畫面本身到位了**：輪詢數不增加只說明「遊戲此刻
	// 沒在讀滑鼠」，而遊戲還在處理上一鍵時本來就不讀。先確定前一鍵已經被讀走
	// （`KeysPending()==0`），再用 dosgolem `docs/spec/012` 的三分法談座標。
	if os.Getenv("KOL_BOOK_PROBE_ENTRY") == "1" {
		before := len(o.MousePollSteps())
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

	// 書本那四步跑完，畫面就停在店主招呼、六個圖示（DOOR／MIRROR／LISTEN／TALK／
	// REST／COMPANION）都在，第一下直接點 COMPANION 招募。`town.IconBar` 是
	// X=32、Step=32，第 6 格的中心是 (204, 181)。
	//
	// 這串鍵綁 Halfway House。**換一座城就要重數**——別把某一座城量到的補救步驟
	// 加進這裡，那會讓本來走得通的路徑多按幾下而走偏。
	for i, key := range []string{
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

		// 觀察模式：露營之後如果方向鍵沒反應，用它找「按什麼才回得到可移動狀態」。
		// 每送一個候選鍵就跟一個 `Right` 測方向鍵有沒有生效——**畫面指紋變了才算
		// 走出去**。
		//
		// 判「不能動」之前先排除兩件事：隊伍是不是真的出了城（沒出城時方向鍵一樣
		// 走不動），以及畫面是不是停在等確認的訊息（伏擊訊息就是這種畫面，那時候
		// 方向鍵本來就無效，不是卡住）。
		if os.Getenv("KOL_BOOK_PROBE_CAMP") == "1" {
			probe := []string{"Right", "Space", "Right", "Return", "Right",
				"Escape", "Right", "Down", "Right", "Up", "Right"}
			prev := screenDigest(o)
			polls := len(o.MousePollSteps())
			t.Logf("露營後起點：指紋 %s，累計輪詢 %d", prev, polls)
			// 先什麼都不按，只讓它跑。輪詢每步只 +8（城鎮等待時是 +300,000），
			// 表示遊戲**不在等按鍵**——那形狀和「角色清單還沒畫完」一樣。
			for r := 0; r < 4; r++ {
				if err := o.Run(60_000_000); err != nil {
					t.Fatalf("空跑第 %d 段：%v", r, err)
				}
				now := screenDigest(o)
				np := len(o.MousePollSteps())
				t.Logf("露營後空跑 %d×60M：指紋 %s（%s）輪詢 +%d IP=%s", r+1, now,
					map[bool]string{true: "變了", false: "沒變"}[now != prev], np-polls, o.IP())
				if now != prev {
					captureBook(t, o, fmt.Sprintf("camp-idle-%d.png", r))
				}
				prev, polls = now, np
			}
			for i, key := range probe {
				if err := o.SendKeys(key); err != nil {
					t.Fatalf("送 %q：%v", key, err)
				}
				if err := o.Run(4_000_000); err != nil {
					t.Fatalf("跑 %q：%v", key, err)
				}
				now := screenDigest(o)
				np := len(o.MousePollSteps())
				changed := "－"
				if now != prev {
					changed = "★ 變了"
					captureBook(t, o, fmt.Sprintf("camp-%02d-%s.png", i, key))
				}
				t.Logf("露營探測 step=%2d key=%-7s %s 指紋 %s 輪詢 +%d IP=%s",
					i, key, changed, now, np-polls, o.IP())
				prev, polls = now, np
			}
			// 畫面不動、按鍵無效、輪詢極低、IP 在五個位元組間打轉——那是**死迴圈**
			// 的形狀，不是「在等玩家」。所以問 dosgolem：原版有沒有要到它沒給的東西。
			t.Logf("露營後未實作呼叫：%v", o.Unimplemented())
			t.Logf("露營後缺檔：%v", o.Missing())
			t.Logf("露營後已開檔尾段：%v", tailOf(o.Opened(), 6))
			t.Log("露營探測結束；只是觀察，不宣稱哪一個鍵是正確的解鎖方式")
			return
		}

		// 掃 seed：在同一次執行裡試很多顆，不要一顆跑一次整條路徑。
		//
		// `seed=1` 是**為了避免遭遇**挑的（SPEC-011：野外那兩條收據「選一顆不會在
		// 半路被攔下的種子」），而規劃與結算的對拍要的剛好相反。走到這裡先存一份
		// 狀態，之後每一顆 seed 都從同一個點重來——差別只有那兩個字組。
		if os.Getenv("KOL_BOOK_SEED_SCAN") == "1" {
			const enemyCount = 0x2D9C
			base := o.Save()
			type hit struct {
				Seed int `json:"seed"`
				Step int `json:"step"`
				N    int `json:"enemy_count"`
			}
			var hits []hit
			scanned := 0
			// **一個時段要走六十到七十五步**（`internal/world/clock.go`：一步約 56 個
			// 時間單位、門檻在 3360…4200 之間），而遭遇是每次時段變換才擲一次。
			// 先前掃 40 顆 × 40 步、12 顆 × 25 步都是 0，不是這些 seed 不好，是
			// **每一顆都還沒走到第一次擲點**。所以寧可少掃幾顆、每顆走遠一點。
			for seed := 1; seed <= 3; seed++ {
				o.Restore(base)
				o.WriteU16(o.DS(0x1600), uint16(seed))
				o.WriteU16(o.DS(0x1602), 0)
				scanned++
				w := newWalker(o)
				walked := 0
				for i := 0; i < 90; i++ {
					if w.step(t, o) {
						walked++
					}
					if n := int(o.Word(o.DS(enemyCount))); n > 0 {
						hits = append(hits, hit{seed, i, n})
						t.Logf("seed=%-3d 第 %2d 步遇敵，敵人數 %d（真的走動 %d 步）",
							seed, i, n, walked)
						break
					}
				}
				if len(hits) == 0 || hits[len(hits)-1].Seed != seed {
					t.Logf("seed=%-3d 九十步沒遇敵（真的走動 %d 步）", seed, walked)
				}
			}
			t.Logf("掃了 %d 顆 seed，%d 顆在 90 步內遇敵", scanned, len(hits))
			body, err := json.MarshalIndent(map[string]any{
				"schema": "kol-ambush-seed-scan-v1", "source": "dosgolem-normal-player-path",
				"note":           "在同一份出城後的狀態上逐顆 seed 重來，每顆最多走 40 步",
				"steps_per_seed": 90, "scanned": scanned, "hits": hits,
			}, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if dir := os.Getenv("KOL_ORACLE_CAPTURE"); dir != "" {
				if err = os.WriteFile(filepath.Join(dir, "ambush-seed-scan.json"), append(body, '\n'), 0600); err != nil {
					t.Fatal(err)
				}
			}
			o.Restore(base)
			return
		}

		// 時刻探針（#65）：從出城後的同一份狀態，四個方向各走八步，逐步記下時刻
		// 與 `sub_297F2` 算成本時讀的兩個字組。DS 位移＝IDA 線性 − 0x2DED0：
		// 時刻 `word_2DF40`＝0x70、時段 `word_2DF42`＝0x72、索引 `word_39A54`＝0xBB84、`word_39A4E`＝0xBB7E、
		// 隊伍移動率 `word_2FA2C`＝0x1B5C、地形碼 `word_3499A`＝0x6ACA。
		if os.Getenv("KOL_BOOK_CLOCK_PROBE") == "1" {
			base := o.Save()
			type row struct {
				Key     string `json:"key"`
				Step    int    `json:"step"`
				Before  int    `json:"time_before"`
				After   int    `json:"time_after"`
				Dy      int    `json:"word_39A54"`
				Dx      int    `json:"word_39A4E"`
				Rate    int    `json:"word_2FA2C"`
				Terrain int    `json:"word_3499A"`
				Moved   bool   `json:"screen_changed"`
				X       int    `json:"x"`
				Y       int    `json:"y"`
				SubX    int    `json:"sub_x"`
				SubY    int    `json:"sub_y"`
				Part    int    `json:"word_2DF42"`
			}
			var rows []row
			signed := func(v uint16) int { return int(int16(v)) }
			// 第五組照 `world-road-run1` 走：往北十二步站上道路，再連按九次 Return
			// （白框在 ROAD）。世界座標 `word_2FB50`／`word_2FB56`＝0x1C80／0x1C86，
			// 次格 `word_2DF36`／`word_2DF38`＝0x66／0x68。
			road := append(strings.Split(strings.Repeat("Up,", 12), ",")[:12],
				strings.Split(strings.Repeat("Return,", 9), ",")[:9]...)
			for _, seq := range [][]string{{"Up"}, {"Down"}, {"Left"}, {"Right"}, road} {
				o.Restore(base)
				last := screenDigest(o)
				n := 8
				if seq[0] == "Left" {
					n = 16 // 東西十四步跨過 DAWN 的門檻，多走兩步看時段哪一步換
				}
				if len(seq) > 1 {
					n = len(seq)
				}
				for i := 0; i < n; i++ {
					key := seq[0]
					if len(seq) > 1 {
						key = seq[i]
					}
					before := signed(o.Word(o.DS(0x70)))
					if err := o.SendKeys(key); err != nil {
						t.Fatalf("送 %q：%v", key, err)
					}
					if err := o.Run(2_000_000); err != nil {
						t.Fatalf("走 %q：%v", key, err)
					}
					now := screenDigest(o)
					rows = append(rows, row{Key: key, Step: i, Before: before,
						After: signed(o.Word(o.DS(0x70))), Dy: signed(o.Word(o.DS(0xBB84))),
						Dx: signed(o.Word(o.DS(0xBB7E))), Rate: signed(o.Word(o.DS(0x1B5C))),
						Terrain: signed(o.Word(o.DS(0x6ACA))), Moved: now != last,
						X: signed(o.Word(o.DS(0x1C80))), Y: signed(o.Word(o.DS(0x1C86))),
						SubX: signed(o.Word(o.DS(0x66))), SubY: signed(o.Word(o.DS(0x68))),
						Part: signed(o.Word(o.DS(0x72)))})
					last = now
					r := rows[len(rows)-1]
					t.Logf("時刻探針 %-6s %2d：%d → %d（扣 %d）時段=%d 39A54=%d 移動率=%d 地形=%d 位置 (%d,%d) 次格 (%d,%d)",
						key, i, r.Before, r.After, r.Before-r.After, r.Part, r.Dy, r.Rate, r.Terrain, r.X, r.Y, r.SubX, r.SubY)
				}
			}
			body, err := json.MarshalIndent(map[string]any{
				"schema": "kol-world-clock-probe-v1", "source": "dosgolem-normal-player-path",
				"address_space": "執行期 DGROUP 相對位移；IDA 線性 − 0x2DED0",
				"rows":          rows,
			}, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if dir := os.Getenv("KOL_ORACLE_CAPTURE"); dir != "" {
				if err = os.WriteFile(filepath.Join(dir, "world-clock-probe.json"), append(body, '\n'), 0600); err != nil {
					t.Fatal(err)
				}
			}
			o.Restore(base)
			return
		}

		// 走到遇敵為止，**不要走固定步數**。遭遇是機率性的（SPEC-011），
		// 固定一百步走完之後 `DS:2D9C`（敵人數）可能還是 0，而症狀是
		// 「參戰上限不一致：count=0」——看起來像擷取壞了，其實是根本沒打起來。
		// 每一步問一次那個字組，非零就停，並記下走到第幾步。
		const enemyCount = 0x2D9C
		encounterAt := -1
		w := newWalker(o)
		moved := 0
		for i := 0; i < 100; i++ {
			if w.step(t, o) {
				moved++
			}
			if n := int(o.Word(o.DS(enemyCount))); n > 0 {
				encounterAt = i
				captureBook(t, o, fmt.Sprintf("next-day-%03d.png", i))
				t.Logf("第 %d 步遇敵：敵人數 %d", i, n)
				break
			}
			if i%10 == 9 {
				captureBook(t, o, fmt.Sprintf("next-day-%03d.png", i))
			}
		}
		if encounterAt < 0 {
			t.Logf("次日一百步走完都沒有遇敵（其中真的走動了 %d 步）", moved)
			t.Log("次日一百步走完都沒有遇敵——固定 seed 下這一條路徑不會打起來，" +
				"後面的戰鬥擷取不會有內容")
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
				if os.Getenv("KOL_BOOK_EXECUTE") == "1" {
					executeFirstRound(t, o)
				}
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
	// 決策四欄依 SPEC-010：+0x22 決定完成、+0x23 攻擊選項、+0x24 防禦選項、
	// +0x25 指令碼（圖號）；方向 +0x28 由選目標格自己寫。
	const defenceOff = 0x24
	const facingOff = 0x28
	// 掃**全部**參戰槽，不要拿 DS:0x9E5A 當索引。那個值在規劃期間會跳
	// （2026-09-11 量到 1→0→0→5→5→0→1），語意還沒定案；拿它索引到的可能是
	// 別人的記錄，於是「誰的旗標立起來了」會看成另一個人的。
	const slotCount = 18
	type actorSnap struct {
		Slot    int `json:"slot"`
		Planned int `json:"planned_flag"`
		Option  int `json:"attack_option"`
		Defence int `json:"defence_option"`
		Command int `json:"command_code"`
		Facing  int `json:"facing"`
	}
	type snap struct {
		Step    int         `json:"step"`
		Key     string      `json:"key"`
		Current int         `json:"current_actor_word"`
		Actors  []actorSnap `json:"actors"`
		IP      string      `json:"ip"`
		Frame   string      `json:"frame,omitempty"`
		// Changed 是這一步相對上一步「畫面哪一塊變了」的外接矩形。白框搬一格
		// 只會動兩小塊，用它就量得出圖示列的格距，不必解 PNG。
		Changed string `json:"changed_box,omitempty"`
		// Icons 是圖示列上有內容的格數，Frame 是白框停在第幾格——兩個合起來說明
		// 「現在在哪一層選單、選到哪」。
		Icons     int `json:"icon_slots"`
		FrameSlot int `json:"framed_slot"`
	}
	// iconBar 量圖示列：起點 X=32、格距 32、Y 170..191（2026-09-11 由白框逐格
	// 移動的差異矩形量得）。回傳有內容的格數與白框停在第幾格（-1 表示沒找到）。
	// 有了這兩個數，log 就看得出「現在在哪一層選單、白框在哪」，不必每步開圖。
	const (
		barX0, barStep = 32, 32
		barY0, barY1   = 170, 191
		barSlots       = 8
	)
	iconBar := func() (int, int) {
		w, _, rgb := o.ScreenRGB()
		at := func(x, y int) (uint8, uint8, uint8) {
			i := (y*w + x) * 3
			if i+2 >= len(rgb) {
				return 0, 0, 0
			}
			return rgb[i], rgb[i+1], rgb[i+2]
		}
		filled, frame, bestWhite := 0, -1, 0
		for g := 0; g < barSlots; g++ {
			x0 := barX0 + g*barStep
			colours := map[uint32]bool{}
			white := 0
			for y := barY0; y <= barY1; y++ {
				for x := x0; x < x0+barStep; x++ {
					r, gg, b := at(x, y)
					colours[uint32(r)<<16|uint32(gg)<<8|uint32(b)] = true
					if r > 200 && gg > 200 && b > 200 {
						white++
					}
				}
			}
			// 空格只有圖示列的底色；有圖示的格子顏色種類明顯更多。
			if len(colours) > 3 {
				filled = g + 1
			}
			// 白框是整格外圍一圈亮白。取白像素**最多**的那一格，不要取第一個超過
			// 門檻的——圖案本身也可能有大片白（主指令列第一格就是），那會讓白框
			// 看起來永遠停在第 0 格。
			if white > 60 && white > bestWhite {
				bestWhite, frame = white, g
			}
		}
		return filled, frame
	}

	// screenBox 回傳兩張畫面差異的外接矩形。
	prevRGB := func() (int, int, []uint8) { w, h, rgb := o.ScreenRGB(); return w, h, rgb }
	lastW, lastH, last := prevRGB()
	_ = lastH
	changedBox := func() string {
		w, h, cur := prevRGB()
		defer func() { lastW, lastH, last = w, h, cur }()
		if w != lastW || len(cur) != len(last) {
			return "（畫面尺寸變了）"
		}
		minX, minY, maxX, maxY := w, h, -1, -1
		for i := 0; i+2 < len(cur) && i+2 < len(last); i += 3 {
			if cur[i] == last[i] && cur[i+1] == last[i+1] && cur[i+2] == last[i+2] {
				continue
			}
			px, py := (i/3)%w, (i/3)/w
			if px < minX {
				minX = px
			}
			if px > maxX {
				maxX = px
			}
			if py < minY {
				minY = py
			}
			if py > maxY {
				maxY = py
			}
		}
		if maxX < 0 {
			return "畫面沒變"
		}
		return fmt.Sprintf("X %d..%d Y %d..%d", minX, maxX, minY, maxY)
	}
	readActor := func(slot int) actorSnap {
		rec := uint16(actorBase + actorStride*slot)
		return actorSnap{Slot: slot,
			Planned: int(o.Byte(o.DS(rec + plannedOff))),
			Option:  int(o.Byte(o.DS(rec + optionOff))),
			Defence: int(o.Byte(o.DS(rec + defenceOff))),
			Command: int(o.Byte(o.DS(rec + commandOff))),
			Facing:  int(o.Byte(o.DS(rec + facingOff)))}
	}
	read := func(step int, key string) snap {
		icons, framed := iconBar()
		st := snap{Step: step, Key: key, Icons: icons, FrameSlot: framed,
			Current: int(o.Word(o.DS(currentSlot))), IP: o.IP().String()}
		for slot := 0; slot < slotCount; slot++ {
			st.Actors = append(st.Actors, readActor(slot))
		}
		return st
	}
	// 只把參戰者那幾槽印出來，其餘留在 JSON 裡。
	brief := func(st snap) string {
		var b strings.Builder
		for _, a := range st.Actors {
			if a.Planned == 0 && a.Option == 0 && a.Defence == 0 && a.Command == 0 {
				continue
			}
			fmt.Fprintf(&b, " [槽%d +22=%d +23=%d +24=%d +25=%#02x 面向%d]",
				a.Slot, a.Planned, a.Option, a.Defence, a.Command, a.Facing)
		}
		if b.Len() == 0 {
			return " （沒有任何一槽的決策欄非零）"
		}
		return b.String()
	}

	// 誰寫 DS:0x9E5A（IDA 的 `word_37D2A`）。RESEARCH-LOG 有它的讀取端——訊息的
	// `|G` 與資訊框都讀它、`<6` 走 `0xBB88 + 0x300×i` 取角色記錄——但「誰設它」
	// 一直沒追。靜態 xref 只涵蓋直接參考，間接寫入看不見，所以用寫入監看。
	// 誰寫 DS:0x9E5A（IDA 的 `word_37D2A`）。RESEARCH-LOG 有它的讀取端——訊息的
	// `|G` 與資訊框都讀它、`<6` 走 `0xBB88 + 0x300×i` 取角色記錄——但「誰設它」
	// 一直沒追。靜態 xref 只涵蓋直接參考，間接寫入看不見，所以用寫入監看。
	//
	// **一次只裝得下一個監看**（`machine.WatchWrites` 只有一組 lo／hi／fn，
	// 後裝的會把先裝的換掉）。所以用一個涵蓋兩者的區間，在回呼裡分類：
	// 參戰記錄區當正對照——它在規劃寫回時一定會動，沒有紀錄就是監看沒生效。
	writes := o.WatchWrites(actorBase, 0x9F00)
	defer func() {
		o.StopWatchingWrites()
		var target []oracle.MemWrite
		control := 0
		for _, w := range *writes {
			switch {
			case w.Off == currentSlot || w.Off == currentSlot+1:
				target = append(target, w)
			case w.Off >= actorBase && w.Off < actorBase+actorStride*slotCount:
				control++
			}
		}
		t.Logf("正對照：參戰記錄區被寫 %d 次（監看區間共 %d 次）", control, len(*writes))
		seen := map[string]bool{}
		for _, w := range target {
			t.Logf("  寫 %04X：%d → %d，IP=%s，第 %d 道指令",
				w.Off, w.Old, w.New, w.IP, w.Step)
			seen[w.IP.String()] = true
		}
		t.Logf("DS:%04X 被寫 %d 次，來自 %d 個位置", currentSlot, len(target), len(seen))
	}()

	var out []snap
	first := read(-1, "(進場後，未送鍵)")
	out = append(out, first)
	t.Logf("規劃入口：圖示列%d格 白框在第%d格 DS:9E5A=%d%s",
		first.Icons, first.FrameSlot, first.Current, brief(first))

	// 序列可從 `KOL_BOOK_PLAN_KEYS` 給（逗號分隔），觀察時不必改程式重編。
	// 預設那組刻意保守：先看單獨一個 Return 會不會就決定一個指令，再看方向鍵與
	// `<`／`>` 有沒有搬動選擇。每一步都留畫面。
	seq := []string{"Return", "Right", "Return", ".", "Return", "Return"}
	if raw := os.Getenv("KOL_BOOK_PLAN_KEYS"); raw != "" {
		seq = strings.Split(raw, ",")
	}
	for i, key := range seq {
		// `.`／`,` 這種字元鍵沒有具名掃描碼，要走 TypeKeys；SendKeys 只認具名鍵。
		var err error
		if len([]rune(key)) == 1 {
			err = o.TypeKeys(key)
		} else {
			err = o.SendKeys(key)
		}
		if err != nil {
			t.Fatalf("送 %q：%v", key, err)
		}
		if err := o.Run(20_000_000); err != nil {
			t.Fatalf("跑 %q：%v", key, err)
		}
		// 同書本那一段：等鍵真的被讀走，不要拿指令數當「處理完了」。
		for n := 0; o.KeysPending() > 0 && n < 20; n++ {
			if err := o.Run(20_000_000); err != nil {
				t.Fatalf("跑 %q：%v", key, err)
			}
		}
		if o.KeysPending() > 0 {
			t.Fatalf("規劃第 %d 步的 %q 送出後 4 億指令仍未被讀走", i, key)
		}
		name := fmt.Sprintf("plan-%02d.png", i)
		captureBook(t, o, name)
		st := read(i, key)
		st.Frame = name
		st.Changed = changedBox()
		out = append(out, st)
		t.Logf("規劃 step=%d key=%-8s 圖示列%d格 白框在第%d格 DS:9E5A=%d IP=%s%s",
			i, key, st.Icons, st.FrameSlot, st.Current, st.IP, brief(st))
	}

	data := map[string]any{
		"schema": "kol-battle-planning-probe-v1", "source": "dosgolem-normal-player-path",
		"alignment_seed": 1,
		"evidence":       "SPEC-010 的 +0x22／+0x23／+0x24／+0x25／+0x28；DS 相對位址；逐槽全掃",
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

// screenDigest 是目前畫面的指紋，用來判斷「這一步到底有沒有走出去」。
func screenDigest(o *oracle.Oracle) string {
	sum := sha256.Sum256(o.Tandy16())
	return hex.EncodeToString(sum[:8])
}

// walker 依序試四個方向走。**撞牆的時候畫面不會變**，而一直按同一個鍵不等於
// 一直在走：角色沒動，世界時間就不走，遭遇也不擲，症狀卻長得像「機率不好」。
// 所以走完要看的是**真的走動了幾步**，不是送了幾個鍵。
type walker struct {
	dirs []string
	at   int
	last string
}

func newWalker(o *oracle.Oracle) *walker {
	return &walker{dirs: []string{"Right", "Down", "Left", "Up"}, last: screenDigest(o)}
}

// step 走一步，回傳畫面有沒有變。沒變就換下一個方向，下一次呼叫改走那一邊。
func (w *walker) step(t *testing.T, o *oracle.Oracle) bool {
	t.Helper()
	if err := o.SendKeys(w.dirs[w.at]); err != nil {
		t.Fatalf("送 %q：%v", w.dirs[w.at], err)
	}
	if err := o.Run(2_000_000); err != nil {
		t.Fatalf("走 %q：%v", w.dirs[w.at], err)
	}
	now := screenDigest(o)
	moved := now != w.last
	w.last = now
	if !moved {
		w.at = (w.at + 1) % len(w.dirs)
	}
	return moved
}

// tailOf 回切片的最後 n 個元素，用來看「最近開了哪幾個檔」。
func tailOf(xs []string, n int) []string {
	if len(xs) <= n {
		return xs
	}
	return xs[len(xs)-n:]
}

// keysOf 把集合攤成排序過的清單。
func keysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// executeFirstRound 接在規劃探針之後：其餘隊員選休息、在「Begin next round?」答 YES，
// 讓原版把第一回合的執行段跑完（#9，攻擊結算）。
//
// 記三樣東西，全部是原版自己的值，不從畫面推：
//
//   - 執行段入口（`sub_1E3A8:0x1FD21`）那一刻的參戰者記錄、耐力／疲勞與 32 位元亂數狀態；
//   - 執行段裡每一次 `sub_1B5A8(n)`（rand(n)）的 n、呼叫點與擲之前的狀態，以及每一次
//     `sub_1A8FD(記錄)` 印的 KOLDATA 記錄號；
//   - 回到下一回合（`0x1EF3A`）或戰鬥出口（`0x211AD`）那一刻的同一組狀態。
//
// 等鍵的 `sub_1B100` 在執行段 stub 掉——它只等玩家看完訊息，不擲骰、不改狀態。
func executeFirstRound(t *testing.T, o *oracle.Oracle) {
	t.Helper()
	dir := os.Getenv("KOL_ORACLE_CAPTURE")
	if dir == "" {
		t.Fatal("執行段探測需要 KOL_ORACLE_CAPTURE")
	}
	// IDA 線性位址換成這一次載入的執行期位址：DGROUP 在 IDA 的 0x2DED0。
	base := uint32(o.DGroupSeg())*16 - 0x2DED0
	at := func(ea uint32) oracle.Addr {
		lin := base + ea
		return oracle.Far(uint16(lin>>4), uint16(lin&0xF))
	}
	const (
		randEA    = 0x1B5A8
		printEA   = 0x1A8FD
		waitEA    = 0x1B100
		execEA    = 0x1FD21
		roundEA   = 0x1EF3A
		exitEA    = 0x211AD
		seedLo    = 0x1600
		seedHi    = 0x1602
		actorBase = 0x6712
		stride    = 50
		slots     = 18
	)
	// 位址換算要先驗：rand(n) 開頭是 `push bp; mov bp, sp`。
	if b := []byte{o.Byte(at(randEA)), o.Byte(at(randEA + 1)), o.Byte(at(randEA + 2))}; !bytes.Equal(b, []byte{0x55, 0x8B, 0xEC}) {
		t.Fatalf("sub_1B5A8 換算錯了：%x", b)
	}
	seed := func() uint32 { return uint32(o.Word(o.DS(seedLo))) | uint32(o.Word(o.DS(seedHi)))<<16 }
	snapshot := func() map[string]any {
		var actors []string
		var counters []map[string]int
		for i := 0; i < slots; i++ {
			raw := make([]byte, stride)
			for j := range raw {
				raw[j] = o.Byte(o.DS(uint16(actorBase + i*stride + j)))
			}
			actors = append(actors, hex.EncodeToString(raw))
			counters = append(counters, map[string]int{"slot": i,
				"stamina": int(o.Word(o.DS(uint16(0x1C50 + i*2)))),
				"fatigue": int(int16(o.Word(o.DS(uint16(0x322E + i*2)))))})
		}
		return map[string]any{"seed": seed(), "actors": actors, "counters": counters}
	}
	type roll struct {
		N      int    `json:"n"`
		Caller string `json:"caller_ida"`
		Before uint32 `json:"seed_before"`
		Phase  string `json:"phase"`
		Round  int    `json:"round"`
	}
	type message struct {
		Record int    `json:"record"`
		Caller string `json:"caller_ida"`
		Round  int    `json:"round"`
	}
	type boundary struct {
		Round int            `json:"round"`
		At    string         `json:"at"`
		State map[string]any `json:"state"`
	}
	// 第一次進執行段起開始記，之後的規劃段也記（敵人的決策在那裡擲骰）。
	// 有人受傷（參戰者記錄 +0x2D 或累積傷害 +0x0E 變了）之後再多跑一回合就停，最多八回合。
	const maxRounds = 8
	var (
		boundaries      []boundary
		rolls           []roll
		messages        []message
		recording, done bool
		round           int
		phase           = "plan"
		hurtRound       = -1
		endAt           string
	)
	ida := func(a oracle.Addr) string { return fmt.Sprintf("%#x", a.Linear()-base) }
	hurt := func() bool {
		for i := 0; i < slots; i++ {
			rec := uint16(actorBase + i*stride)
			if o.Byte(o.DS(rec+0x0E)) != 0 || o.Byte(o.DS(rec+0x2D)) != 0 {
				return true
			}
		}
		return false
	}
	o.OnCall(at(execEA), func(p *oracle.Oracle) {
		if done {
			return
		}
		if !recording {
			recording = true
			round = 1
		}
		// 等鍵只在執行段 stub：規劃段靠它等玩家按鍵，stub 掉會空轉。
		p.StubValue(at(waitEA), 0)
		phase = "exec"
		boundaries = append(boundaries, boundary{round, "執行段入口 0x1FD21", snapshot()})
	})
	o.OnCall(at(randEA), func(p *oracle.Oracle) {
		if recording && !done && len(rolls) < 20000 {
			rolls = append(rolls, roll{N: int(p.Arg(0)), Caller: ida(p.Caller()), Before: seed(), Phase: phase, Round: round})
		}
	})
	asked := false // 印過記錄 0x108「Begin next round?」
	o.OnCall(at(printEA), func(p *oracle.Oracle) {
		if recording && !done && len(messages) < 20000 {
			messages = append(messages, message{Record: int(p.Arg(0)), Caller: ida(p.Caller()), Round: round})
		}
		if p.Arg(0) == 0x108 {
			asked = true
		}
	})
	o.OnCall(at(roundEA), func(p *oracle.Oracle) {
		if !recording || done {
			return
		}
		boundaries = append(boundaries, boundary{round, "下一回合規劃 0x1EF3A", snapshot()})
		p.Stub(at(waitEA), nil)
		if hurtRound < 0 && hurt() {
			hurtRound = round
		}
		if round >= maxRounds || (hurtRound >= 0 && round > hurtRound) {
			done, endAt = true, "下一回合規劃 0x1EF3A"
			return
		}
		round++
		phase = "plan"
		asked = false
		p.Stub(at(waitEA), nil)
	})
	o.OnCall(at(exitEA), func(p *oracle.Oracle) {
		if recording && !done {
			boundaries = append(boundaries, boundary{round, "戰鬥出口 0x211AD", snapshot()})
			done, endAt = true, "戰鬥出口 0x211AD"
		}
	})

	// 原版的規劃是「選指令寫回、畫面留在同一名；按 YES（第 0 格，圖號 7）才換下一名」
	// （`0x1F38D`：`sub_228B7` 回 7 才跳 `loc_1FAC4` 設 `+0x22` 換人）。所以這裡每一步都把
	// 白框移回第 0 格再 Return：Theodoric 確認寫回的徒手、其餘隊員不選指令直接確認、
	// 最後在「Begin next round?」按 YES。指令碼 `+0x25` 跨回合保留，之後每回合按 YES 就是
	// 重複上一回合的指令。白框位置由畫面量（同規劃探針的量法）。
	frameSlot := func() int {
		w, _, rgb := o.ScreenRGB()
		frame, best := -1, 0
		for g := 0; g < 8; g++ {
			white := 0
			for y := 170; y <= 191; y++ {
				for x := 32 + g*32; x < 64+g*32; x++ {
					i := (y*w + x) * 3
					if rgb[i] > 200 && rgb[i+1] > 200 && rgb[i+2] > 200 {
						white++
					}
				}
			}
			if white > 60 && white > best {
				best, frame = white, g
			}
		}
		return frame
	}
	press := func(key string) {
		var err error
		if len([]rune(key)) == 1 {
			err = o.TypeKeys(key)
		} else {
			err = o.SendKeys(key)
		}
		if err != nil {
			t.Fatalf("送 %q：%v", key, err)
		}
		for n := 0; n < 20; n++ {
			if err := o.Run(20_000_000); err != nil {
				t.Fatalf("跑 %q：%v", key, err)
			}
			if o.KeysPending() == 0 {
				break
			}
		}
	}
	for n := 0; n < 80 && !done; n++ {
		for k := 0; k < 8 && frameSlot() > 0; k++ {
			press(",")
		}
		if n < 4 {
			captureBook(t, o, fmt.Sprintf("exec-plan-%02d.png", n))
		}
		t.Logf("第 %d 次：回合 %d，白框在第 %d 格，已印 0x108=%v，按 Return", n, round, frameSlot(), asked)
		press("Return")
		for k := 0; k < 50 && phase == "exec" && !done; k++ {
			if err := o.Run(20_000_000); err != nil {
				t.Fatal(err)
			}
		}
	}
	o.Stub(at(waitEA), nil)
	captureBook(t, o, "exec-end.png")
	if !done {
		t.Fatalf("執行段沒有跑完：回合 %d，擲骰 %d 次、訊息 %d 則", round, len(rolls), len(messages))
	}
	data := map[string]any{
		"schema": "kol-battle-rounds-v1", "source": "dosgolem-normal-player-path",
		"input": "KOLTOWN-unpacked.EXE", "input_sha256": "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2",
		"address_space": "記錄為 DGROUP 相對位移；caller_ida 為 IDA linear",
		"evidence":      "RE-132 亂數、RE-080 參戰者記錄、RE-143 執行段入口；#9 攻擊結算比較",
		"alignment_seed": 1, "rounds": round, "hurt_round": hurtRound, "end_at": endAt,
		"boundaries": boundaries, "rolls": rolls, "messages": messages,
		"inputs": "隊員每回合按 YES：Theodoric 第一回合選徒手，之後重複；其餘隊員不選指令",
		"limitations": []string{"等鍵 sub_1B100 在記錄期間 stub 掉", "只到有人受傷後一回合或第八回合"},
	}
	body, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "battle-rounds.json"), append(body, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("跑了 %d 回合（第 %d 回合有人受傷）：擲骰 %d 次、訊息 %d 則，停在 %s", round, hurtRound, len(rolls), len(messages), endAt)
}
