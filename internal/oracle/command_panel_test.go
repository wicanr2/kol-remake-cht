//go:build oracle

package oracle

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
)

// 實際執行目標常式與智力檢定，僅替換繪圖及輸入出口。
func TestBattleCommandVisibilityOriginal(t *testing.T) {
	for _, target := range []int{0, 6, 17} {
		for _, threshold := range []uint16{39, 40, 41} {
			o := prepare(t)
			for _, ea := range []uint32{0x1449C, 0x11703, 0x11722, 0x1C683, 0x1C601, 0x1B90B, 0x223C0, 0x1BA9D, 0x22EFC} {
				o.StubValue(o.IDA(ea), 0)
			}
			o.WriteU16(o.IDA(0x37D2A), 0)
			o.WriteU16(o.IDA(0x2EA18), 0)
			o.WriteU16(o.IDA(0x2DF3A), 0)
			o.WriteU8(o.DS(0x6713), 0)
			o.WriteU8(o.DS(0x673A), 0)
			o.WriteU8(o.DS(0x6712+0xC), 60)
			o.WriteU8(o.DS(0x6712+2), 9)
			if target >= 6 {
				o.WriteU8(o.DS(uint16(0x6712+50*target+2)), 2)
			}
			o.WriteU16(o.IDA(0x3B264), threshold)
			for i := 0; i < 81; i++ {
				code := byte(1)
				if target >= 6 {
					code = 7
				}
				o.WriteU8(o.DS(uint16(0x6AE4+i)), code)
				o.WriteU8(o.DS(uint16(0x1C90+i)), byte(target))
			}
			draws, visible := 0, false
			o.Stub(o.IDA(0x2324F), func(p *oracle.Oracle) uint32 { draws++; visible = p.Arg(0) != 0; return 0 })
			o.StubValue(o.IDA(0x1B100), 0xD)
			o.Stub(o.IDA(0x1B5A8), func(p *oracle.Oracle) uint32 { t.Fatal("檢視不應另擲骰"); return 0 })
			if _, err := o.Call(o.IDA(0x24131), 1); err != nil {
				t.Fatal(err)
			}
			s := &battle.State{Stage: battle.StageField, Party: []battle.Member{{}}, Actors: make([]battle.Actor, 18), Round: battle.Round{Order: []int{0}, InspectionRoll: int(threshold)}, Plan: battle.Planning{BodyActor: target + 1, Command: 7}}
			s.Actors[0] = battle.Actor{Side: battle.SideParty, Member: 0, Intellect: 60, Level: 9}
			if target >= 6 {
				s.Actors[target] = battle.Actor{Side: battle.SideMonster, Level: 2, Command: 7}
			}
			want := target < 6 || threshold < 40
			if draws != 1 || visible != want || (len(s.CommandPanel()) > 0) != visible {
				t.Fatalf("target=%d threshold=%d 原版=%v remake=%v 預期=%v", target, threshold, visible, s.CommandPanel(), want)
			}
		}
	}
}

// 有界取消返回切片：只在主規劃重畫入口設返回出口，原始跳轉保持不變。
func TestBattlePlanningCancelOriginal(t *testing.T) {
	o := prepare(t)
	o.WriteU8(o.IDA(0x1F16A), 0xCB)
	o.WriteU16(o.IDA(0x3B264), 73)
	o.WriteU16(o.IDA(0x2F972), 2)
	o.Stub(o.IDA(0x1B5A8), func(p *oracle.Oracle) uint32 { t.Fatal("取消不得重擲門檻"); return 0 })
	entry := o.IDA(0x1E3A8)
	entry.Off += 0x1FAC1 - 0x1E3A8 // 保留原函式段，讓向後 near JMP 不發生 IP 下溢。
	for n := 0; n < 3; n++ {
		if _, err := o.Call(entry); err != nil {
			t.Fatal(err)
		}
		if o.Word(o.IDA(0x3B264)) != 73 || o.Word(o.IDA(0x2F972)) != 2 {
			t.Fatal("取消改變門檻或規劃位置")
		}
	}
}

// 有界執行初始化上界計算；輸入欄位由既有遭遇資料決定，非正常戰鬥啟動。
func TestBattlePlanningLimitOriginal(t *testing.T) {
	o := prepare(t)
	o.WriteU8(o.IDA(0x1E4C0), 0xCB)
	for count := uint16(1); count <= 12; count++ {
		o.WriteU16(o.IDA(0x30C6C), count)
		if _, err := o.Call(o.IDA(0x1E4AE)); err != nil {
			t.Fatal(err)
		}
		if o.Word(o.IDA(0x322F4)) != count-1 || o.Word(o.IDA(0x2FBC0)) != count+5 {
			t.Fatalf("敵人數 %d 上界錯誤", count)
		}
	}
}

// 有界原版規劃切片：略過互動 UI，保留排序位置、門檻、跳過及 AI 的原始控制流。
// 不是正常玩家路徑；兩處出口修改與完整輸入都在此明列。
func TestBattlePlanningInspectionOrderOriginal(t *testing.T) {
	o := prepare(t)
	// 隊員互動入口跳至原版完成該位置；整輪規劃末端返回測試呼叫者。
	o.WriteBytes(o.IDA(0x1F155), []byte{0xE9, 0x6C, 0x09})
	o.WriteU8(o.IDA(0x1FAEA), 0xCB)
	order := []uint16{0, 6, 1, 2, 7}
	o.WriteU16(o.IDA(0x2F972), 0)
	o.WriteU16(o.IDA(0x2FBC0), uint16(len(order)-1))
	for pos, slot := range order {
		o.WriteU8(o.DS(uint16(0x6712+pos*50+0x21)), byte(slot))
		side := byte(0)
		if slot >= 6 {
			side = 1
		}
		if slot == 2 {
			side = 2
		}
		o.WriteU8(o.DS(0x6712+50*slot+1), side)
	}
	var trace []string
	rolls := 0
	o.Stub(o.IDA(0x1B5A8), func(p *oracle.Oracle) uint32 {
		if p.Arg(0) != 100 {
			t.Fatal("門檻骰面不同")
		}
		rolls++
		trace = append(trace, fmt.Sprintf("roll:%d", p.Word(p.IDA(0x37D2A))))
		return uint32(rolls + 30)
	})
	o.Stub(o.IDA(0x22866), func(p *oracle.Oracle) uint32 {
		slot := p.Arg(0)
		trace = append(trace, fmt.Sprintf("check:%d", slot))
		if p.Word(p.IDA(0x3B264)) != uint16(rolls+30) {
			t.Fatal("未先保存門檻")
		}
		if slot == 1 {
			return 1
		}
		return 0
	})
	o.Stub(o.IDA(0x2374C), func(p *oracle.Oracle) uint32 {
		trace = append(trace, fmt.Sprintf("ai:%d", p.Arg(0)))
		return 0
	})
	if _, err := o.Call(o.IDA(0x1F05C)); err != nil {
		t.Fatal(err)
	}
	want := []string{"roll:0", "check:0", "roll:6", "check:6", "ai:6", "roll:1", "check:1", "roll:2", "check:2", "roll:7", "check:7", "ai:7"}
	if !reflect.DeepEqual(trace, want) || rolls != 5 || o.Word(o.IDA(0x2F972)) != 5 {
		t.Fatalf("規劃切片順序=%v rolls=%d", trace, rolls)
	}
	for _, slot := range order {
		if o.Byte(o.DS(0x6712+50*slot+0x22)) != 1 {
			t.Fatalf("slot %d 未完成規劃標記", slot)
		}
	}
}

func TestBattleCommandIntellectOriginal(t *testing.T) {
	o := prepare(t)
	o.Stub(o.IDA(0x1B5A8), func(p *oracle.Oracle) uint32 { t.Fatal("檢定本體不應另擲亂數"); return 0 })
	for _, slot := range []uint16{6, 17} {
		o.WriteU16(o.IDA(0x37D2A), slot)
		o.WriteU8(o.DS(0x6712+2), 9) // 行動者等級故意不同。
		for _, level := range []byte{1, 2, 3} {
			o.WriteU8(o.DS(0x6712+50*slot+2), level)
			for _, value := range []uint16{0, 23, 60} {
				for _, threshold := range []uint16{0, 7, 8, 15, 16, 20, 40, 60, 100} {
					got, err := o.Call(o.IDA(0x252CB), value, threshold)
					if err != nil {
						t.Fatal(err)
					}
					want := uint32(0)
					if int(value)*int(level)/3 > int(threshold) {
						want = 1
					}
					// 原版回傳 16 位 AX；DX 保留 DIV 餘數，不屬於布林回傳值。
					if uint16(got) != uint16(want) {
						t.Fatalf("slot=%d level=%d value=%d threshold=%d got=%d want=%d", slot, level, value, threshold, got, want)
					}
				}
			}
		}
	}
}

// 目標格只對敵人的命令圖示做檢定；上方裝備與人形不受此閘門影響。
func TestBattleCommandInspectionOriginal(t *testing.T) {
	for _, target := range []int{0, 6, 17} {
		for _, result := range []uint32{0, 1} {
			o := prepare(t)
			for _, ea := range []uint32{0x1449C, 0x11703, 0x11722, 0x1C683, 0x1C601, 0x1B90B, 0x223C0, 0x1BA9D} {
				o.StubValue(o.IDA(ea), 0)
			}
			o.WriteU16(o.IDA(0x37D2A), 0)
			o.WriteU16(o.IDA(0x2EA18), 0)
			o.WriteU16(o.IDA(0x2DF3A), 0)
			o.WriteU8(o.DS(0x6713), 0)
			o.WriteU8(o.DS(0x673A), 0)
			o.WriteU8(o.DS(0x6712+0xC), 23)
			o.WriteU16(o.IDA(0x3B264), 61)
			for i := 0; i < 81; i++ {
				code := byte(1)
				if target >= 6 {
					code = 7
				}
				o.WriteU8(o.DS(uint16(0x6AE4+i)), code)
				o.WriteU8(o.DS(uint16(0x1C90+i)), byte(target))
			}
			checks, upper, lower := 0, 0, 0
			o.Stub(o.IDA(0x252CB), func(p *oracle.Oracle) uint32 {
				checks++
				if p.Arg(0) != 23 || p.Arg(1) != 61 {
					t.Fatalf("檢定參數 %d %d", p.Arg(0), p.Arg(1))
				}
				return result
			})
			o.Stub(o.IDA(0x22EFC), func(p *oracle.Oracle) uint32 { upper++; return 0 })
			o.Stub(o.IDA(0x2324F), func(p *oracle.Oracle) uint32 {
				lower++
				want := uint16(1)
				if target >= 6 && result == 0 {
					want = 0
				}
				if p.Arg(0) != want || p.Word(p.IDA(0x37D2A)) != uint16(target) {
					t.Fatalf("target=%d result=%d 閘門=%d 人物=%d", target, result, p.Arg(0), p.Word(p.IDA(0x37D2A)))
				}
				return 0
			})
			o.StubValue(o.IDA(0x1B100), 0xD)
			if _, err := o.Call(o.IDA(0x24131), 1); err != nil {
				t.Fatal(err)
			}
			wantChecks := 0
			if target >= 6 {
				wantChecks = 1
			}
			if checks != wantChecks || upper != 1 || lower != 1 {
				t.Fatalf("target=%d checks=%d upper=%d lower=%d", target, checks, upper, lower)
			}
		}
	}
}

// 固定原版 KOLDATA 五筆內容，前 252 筆以空記錄隔開；不冒充正常載入文字庫。
func TestBattleCommandTextOriginal(t *testing.T) {
	o := prepare(t)
	bank := strings.Repeat("\r", 252) + "~V5B~H02OPT\r~V5B~H02ATK\r~V7B~H02LOC\r~V7B~H06DEF\r~V5B~H06DIR\r"
	o.WriteBytes(o.DS(0x9E5C), []byte(bank))
	o.StubValue(o.IDA(0x1103F), 0)
	for i, c := range []struct {
		text   string
		col, y uint16
	}{{"OPT", 2, 91}, {"ATK", 2, 91}, {"LOC", 2, 123}, {"DEF", 6, 123}, {"DIR", 6, 91}} {
		var chars []byte
		o.WriteU16(o.IDA(0x2DF14), 0xDEAD)
		o.WriteU16(o.IDA(0x2DF16), 0xDEAD)
		o.Stub(o.IDA(0x1C2DA), func(p *oracle.Oracle) uint32 {
			chars = append(chars, byte(p.Arg(0)))
			if p.Word(p.IDA(0x2DF14)) != c.col || p.Word(p.IDA(0x2DF16)) != c.y {
				t.Fatal("原版標籤座標不同")
			}
			// 字元出口被替換，不模擬字元寬度，也不聲稱已驗像素。
			return 0
		})
		if _, err := o.Call(o.IDA(0x23638), uint16(252+i), 99, 2, 7); err != nil {
			t.Fatal(err)
		}
		if string(chars) != c.text {
			t.Fatalf("標籤=%q 預期=%q", chars, c.text)
		}
	}
}

func TestBattleCommandDirectionOriginal(t *testing.T) {
	o := prepare(t)
	o.StubValue(o.IDA(0x1A8FD), 0)
	const cache = 0x7777
	o.WriteU16(o.IDA(0x39820), cache)
	for direction := uint16(0); direction < 8; direction++ {
		calls := 0
		o.Stub(o.IDA(0x11262), func(p *oracle.Oracle) uint32 {
			calls++
			if p.Arg(0) != 74+direction || p.Arg(1) != cache {
				t.Fatalf("方向圖示路由 %x %x", p.Arg(0), p.Arg(1))
			}
			rect := [4]uint16{p.Word(p.IDA(0x3499E)), p.Word(p.IDA(0x349A0)), p.Word(p.IDA(0x34A06)), p.Word(p.IDA(0x34A0E))}
			if rect != [4]uint16{6, 99, 9, 119} {
				t.Fatal(rect)
			}
			return 0
		})
		if _, err := o.Call(o.IDA(0x23638), 0x100, 99, 6, 0x47+direction); err != nil {
			t.Fatal(err)
		}
		if calls != 1 {
			t.Fatalf("方向圖示繪製次數 %d", calls)
		}
	}
}

// 僅替換繪圖出口；實際執行原版選擇、映射及分支，不作正常玩家路徑證據。
func TestBattleCommandPanelOriginal(t *testing.T) {
	o := prepare(t)
	for _, ea := range []uint32{0x1BA9D, 0x1B779, 0x27E20} {
		o.StubValue(o.IDA(ea), 0)
	}
	o.WriteU16(o.IDA(0x2DF1E), 0)
	o.WriteU16(o.IDA(0x2EA18), 0)
	t.Logf("DS:0B4C=%x DS:0B56=%x", o.Bytes(o.DS(0xB4C), 10), o.Bytes(o.DS(0xB56), 10))
	// 固定原版輸入的兩張十項 byte 表；期望值不從被測索引計算回讀。
	forms := []byte{0x2B, 0x2C, 0x2C, 0x2C, 0x2D, 0x2D, 0x2D, 0x2E, 0x2E, 0x2E}
	heights := []byte{0x41, 0x2F, 0x41, 0x42, 0x2F, 0x41, 0x42, 0x2F, 0x41, 0x42}
	if !reflect.DeepEqual(o.Bytes(o.DS(0xB4C), 10), forms) || !reflect.DeepEqual(o.Bytes(o.DS(0xB56), 10), heights) {
		t.Fatal("原版圖示映射表與固定輸入證據不同")
	}
	var got [][4]uint16
	o.Stub(o.IDA(0x23638), func(p *oracle.Oracle) uint32 {
		got = append(got, [4]uint16{p.Arg(0), p.Arg(1), p.Arg(2), p.Arg(3)})
		return 0
	})
	count := 0
	for _, slot := range []int{0, 6} {
		o.WriteU16(o.DS(0x9E5A), uint16(slot))
		o.WriteU16(o.DS(uint16(0x1C50+2*slot)), 100)
		o.WriteU16(o.DS(uint16(0x322E+2*slot)), 0)
		for _, enabled := range []uint16{0, 1} {
			for _, command := range []uint16{0, 0xF, 0x21, 0x27, 0x29, 0x2A, 0x3F, 7} {
				for _, style := range []uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10} {
					for _, guard := range []uint16{0, 1, 6} {
						for _, direction := range []byte{0, 7, 8} {
							// 另一來源故意設成不同值，證明當前人物取全域、其他人物取記錄。
							o.WriteU16(o.IDA(0x31122), 0)
							o.WriteU16(o.IDA(0x30A78), 0)
							o.WriteU16(o.IDA(0x30916), 0)
							base := uint16(0x6712 + 50*slot)
							o.WriteU8(o.DS(base+0x25), 0)
							o.WriteU8(o.DS(base+0x23), 0)
							o.WriteU8(o.DS(base+0x24), 0)
							if slot == 0 {
								o.WriteU16(o.IDA(0x31122), command)
								o.WriteU16(o.IDA(0x30A78), style)
								o.WriteU16(o.IDA(0x30916), guard)
							} else {
								o.WriteU8(o.DS(base+0x25), byte(command))
								o.WriteU8(o.DS(base+0x23), byte(style))
								o.WriteU8(o.DS(base+0x24), byte(guard))
							}
							o.WriteU8(o.DS(base+0x28), direction)
							var want [][4]uint16
							add := func(record, y, x, icon uint16) { want = append(want, [4]uint16{record, y, x, icon}) }
							face := uint16(0x47)
							if direction < 8 {
								face += uint16(direction)
							}
							block := uint16(0x2A)
							if guard != 0 {
								block = guard + 0x2F
								if block == 0x35 {
									block++
								}
							}
							if enabled != 0 {
								switch command {
								case 0, 0xF, 0x2A:
								case 0x21:
									if style == 0 {
										add(0xFD, 99, 2, 0x2A)
									} else {
										add(0xFD, 99, 2, uint16(forms[style-1]))
										add(0xFE, 131, 2, uint16(heights[style-1]))
									}
									add(0xFF, 131, 6, block)
									add(0x100, 99, 6, face)
								case 0x27, 0x3F:
									add(0xFC, 99, 2, style)
									add(0x100, 99, 6, face)
								case 0x29:
									icon := uint16(0x2A)
									if style != 0 {
										icon = style + 0x36
									}
									add(0xFD, 99, 2, icon)
									add(0xFF, 131, 6, block)
									add(0x100, 99, 6, face)
								default:
									add(0xFC, 99, 2, command)
								}
							}
							got = nil
							if _, err := o.Call(o.IDA(0x2324F), enabled); err != nil {
								t.Fatal(err)
							}
							if !reflect.DeepEqual(got, want) {
								t.Fatalf("slot=%d enabled=%d command=%x style=%d guard=%d direction=%d got=%v want=%v", slot, enabled, command, style, guard, direction, got, want)
							}
							// 同一組輸入直接比較正式面板，不只核對原版與手寫期望表。
							s := &battle.State{Stage: battle.StageField, Party: []battle.Member{{}}, Actors: make([]battle.Actor, 7), Round: battle.Round{Order: []int{0}, InspectionRoll: 1}, Plan: battle.Planning{BodyActor: slot + 1}}
							s.Actors[0] = battle.Actor{Side: battle.SideParty, Member: 0, Intellect: 100}
							if slot == 0 {
								s.Plan.Command, s.Plan.Style, s.Plan.Guard = int(command), int(style), int(guard)
							} else {
								s.Actors[slot] = battle.Actor{Side: battle.SideMonster, Level: 3, Command: int(command), Style: int(style), Guard: int(guard)}
							}
							s.Actors[slot].Facing = int(direction)
							if enabled == 0 {
								s.Plan.BodyActor = 0
							}
							var remake [][4]uint16
							labels := map[string]uint16{"OPT": 0xFC, "ATK": 0xFD, "LOC": 0xFE, "DEF": 0xFF, "DIR": 0x100}
							for _, icon := range s.CommandPanel() {
								record, ok := labels[icon.Label]
								if !ok {
									t.Fatalf("未知標籤 %q", icon.Label)
								}
								remake = append(remake, [4]uint16{record, uint16(icon.Y), uint16(icon.X / 8), uint16(icon.Code)})
							}
							if !reflect.DeepEqual(got, remake) {
								t.Fatalf("原版=%v remake=%v command=%x style=%d guard=%d", got, remake, command, style, guard)
							}
							count++
						}
					}
				}
			}
		}
	}
	t.Logf("%d 組命令圖示參數序列通過", count)
}

func TestBattleCommandSpriteOriginal(t *testing.T) {
	o := prepare(t)
	for _, c := range [][4]uint16{{0xFC, 99, 2, 7}, {0xFD, 99, 2, 0x2A}, {0xFE, 131, 2, 0x30}, {0xFF, 131, 6, 0x36}, {0x100, 99, 6, 0x4E}} {
		var calls []string
		o.Stub(o.IDA(0x1A8FD), func(p *oracle.Oracle) uint32 {
			calls = append(calls, "text")
			if p.Arg(0) != c[0] || p.Arg(1) != 0x9E5C {
				t.Fatalf("文字出口參數 %x %x", p.Arg(0), p.Arg(1))
			}
			return 0
		})
		o.Stub(o.IDA(0x1103F), func(p *oracle.Oracle) uint32 {
			calls = append(calls, "sprite")
			rect := [4]uint16{p.Word(p.IDA(0x3499E)), p.Word(p.IDA(0x349A0)), p.Word(p.IDA(0x34A06)), p.Word(p.IDA(0x34A0E))}
			if p.Arg(0) != c[3]+3 || rect != [4]uint16{c[2], c[1], c[2] + 3, c[1] + 20} {
				t.Fatalf("圖號=%x 矩形=%v", p.Arg(0), rect)
			}
			return 0
		})
		if _, err := o.Call(o.IDA(0x23638), c[0], c[1], c[2], c[3]); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(calls, []string{"text", "sprite"}) {
			t.Fatal(calls)
		}
	}
}
