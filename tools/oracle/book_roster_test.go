//go:build oracle

package oracle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

// 名冊探針只從 KOL.EXE 正常開機、送真實按鍵；不寫名字／槽位記憶體。
// 唯一測試注入是原版自然呼叫 _srand 時的 seed 參數，固定為 1。
// 定位沿用 KOLBOOK-rng-summary.json：IDA 線性 0x18922，near call 0x1015b。
func TestBookRosterBoundariesOriginal(t *testing.T) {
	root, out := os.Getenv("KOL_ORACLE_ROOT"), os.Getenv("KOL_ORACLE_CAPTURE")
	if root == "" || out == "" {
		t.Skip("需要玩家自備原版 KOL_ORACLE_ROOT 與 KOL_ORACLE_CAPTURE")
	}
	scenario := os.Getenv("KOL_ROSTER_SCENARIO")
	if scenario != "full" && scenario != "duplicate" && scenario != "charset" {
		t.Fatal("KOL_ROSTER_SCENARIO 必須是 full、duplicate 或 charset")
	}
	o, err := oracle.Load(filepath.Join(root, "KOL.EXE"), root)
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()
	scratch := filepath.Join(out, "saved")
	if err = os.MkdirAll(scratch, 0755); err != nil {
		t.Fatal(err)
	}
	o.SetScratch(scratch) // 原版 writer 使用 create；AllowFileWrites 只涵蓋既有檔開啟。
	readFile := func(name string) []byte {
		for _, dir := range []string{scratch, root} {
			entries, e := os.ReadDir(dir)
			if e != nil {
				t.Fatal(e)
			}
			for _, entry := range entries {
				if strings.EqualFold(entry.Name(), name) {
					b, e := os.ReadFile(filepath.Join(dir, entry.Name()))
					if e != nil {
						t.Fatal(e)
					}
					return b
				}
			}
		}
		t.Fatalf("沒有 %s", name)
		return nil
	}
	var rows []map[string]any
	var seedEvents []map[string]any
	files := func() map[string]string {
		result := map[string]string{}
		for _, name := range []string{"CHARDATA", "CHARPICS", "SHLDPICS"} {
			b := readFile(name)
			digest := sha256.Sum256(b)
			result[name] = hex.EncodeToString(digest[:])
		}
		return result
	}
	names := func() []string {
		b := readFile("CHARDATA")
		if len(b) != 16*768 {
			t.Fatalf("CHARDATA 大小：%d", len(b))
		}
		result := make([]string, 16)
		for i := range result {
			result[i] = string(bytes.SplitN(b[i*768:i*768+16], []byte{0}, 2)[0])
		}
		return result
	}
	before := files()
	defer func() {
		receipt := map[string]any{"schema": "kol-book-roster-observation-v1", "scenario": scenario,
			"status": "observation-requires-review", "source": "dosgolem-normal-KOL.EXE-player-input",
			"seed": 1, "seed_events": seedEvents, "steps": rows, "before": before, "after": files(),
			"names_after": names(), "console": o.Console(), "key_reads": o.KeyReads(), "missing": o.Missing(), "unimplemented": o.Unimplemented(), "exec_log": o.ExecLog()}
		b, e := json.MarshalIndent(receipt, "", "  ")
		if e != nil {
			t.Error(e)
			return
		}
		if e = os.WriteFile(filepath.Join(out, "roster-observation.json"), append(b, '\n'), 0644); e != nil {
			t.Error(e)
		}
	}()
	// KOL 以 AL=03h 覆疊載入書本；dosgolem 的 ExecRecord.PSP 此時保存的是
	// loadSeg（internal/dos/exec.go:loadOverlay），不是 AL=00h 子行程 PSP。
	loaded := false
	for i := 0; i < 20 && !loaded; i++ {
		if i >= 4 {
			o.Type("\r")
		}
		condition := oracle.NewCond("正常 EXEC 載入 KOLBOOK", func(o *oracle.Oracle) bool {
			for _, x := range o.ExecLog() {
				if strings.EqualFold(filepath.Base(x.Base), "KOLBOOK.EXE") {
					return true
				}
			}
			return false
		})
		err = o.RunUntil(condition, oracle.Budget(50_000_000))
		if _, ok := err.(*oracle.BudgetError); err != nil && !ok {
			t.Fatal(err)
		}
		for _, x := range o.ExecLog() {
			if !strings.EqualFold(filepath.Base(x.Base), "KOLBOOK.EXE") {
				continue
			}
			loaded = true
			seg := x.PSP
			o.OnCall(oracle.Far(seg, 0x8922), func(o *oracle.Oracle) {
				if o.NearCaller().Off != 0x015e {
					t.Fatalf("_srand 呼叫端不符：%s", o.NearCaller())
				}
				r := o.Regs()
				arg := oracle.Far(r.SS, r.SP+2)
				original := o.Word(arg)
				o.SetWord(arg, 1)
				seedEvents = append(seedEvents, map[string]any{"ida": "0x18922", "caller_ida": "0x1015b", "runtime": o.IP().String(), "step": o.Steps(), "original_arg": original, "fixed_arg": 1})
			})
		}
	}
	if !loaded {
		t.Fatal("正常啟動未載入 KOLBOOK")
	}
	// 覆疊載入尚未表示殼程式已呼叫書本；沿用既有正常開機的 DOS Enter。
	for i := 0; len(seedEvents) == 0 && i < 20; i++ {
		o.Type("\r")
		if err = o.Run(50_000_000); err != nil {
			t.Fatal(err)
		}
		t.Logf("書本進入等待 %d IP=%s 主控台=%q", i, o.IP(), o.Console())
	}
	if err = o.Run(5_000_000); err != nil {
		t.Fatal(err)
	}
	snap := func(label, key string) {
		t.Helper()
		o.SyncDGroupFromDS()
		frame := fmt.Sprintf("step-%03d.png", len(rows))
		captureBook(t, o, frame)
		buffer := o.Bytes(o.DS(0x1450), 17) // byte_1A690；既有名字輸入 helper 的暫存。
		row := map[string]any{"label": label, "key": key, "step": o.Steps(), "ip": o.IP().String(),
			"keys_pending": o.KeysPending(), "keys_consumed": o.KeysConsumed(), "frame": frame,
			"dos_pending": o.Pending(), "key_reads": o.KeyReads(),
			"name_buffer_hex": hex.EncodeToString(buffer), "files": files(), "names": names()}
		rows = append(rows, row)
		t.Logf("%s key=%q IP=%s 名字=%q", label, key, o.IP(), bytes.SplitN(buffer, []byte{0}, 2)[0])
	}
	fastCharset := false
	press := func(key, label string) {
		t.Helper()
		if len(key) == 1 {
			err = o.TypeKeys(key)
		} else {
			err = o.SendKeys(key)
		}
		if err != nil {
			t.Fatal(err)
		}
		budget, settle := uint64(10_000_000), uint64(20_000_000)
		if fastCharset {
			budget, settle = 1_000_000, 2_000_000
		}
		if err = o.Run(budget); err != nil {
			t.Fatal(err)
		}
		for i := uint64(0); o.KeysPending() > 0 && i < 400_000_000; i += budget {
			if err = o.Run(budget); err != nil {
				t.Fatal(err)
			}
		}
		if o.KeysPending() != 0 {
			t.Fatalf("%q 按鍵未被讀走", key)
		}
		if err = o.Run(settle); err != nil {
			t.Fatal(err)
		}
		snap(label, key)
	}
	snap("正常書本封面", "")
	press("Space", "書本目錄")
	press("5", "角色選項")
	press("Down", "選建立角色")
	press("Return", "名字輸入")
	if err = o.Run(60_000_000); err != nil {
		t.Fatal(err)
	}
	snap("名字頁繪製完成", "")
	if len(seedEvents) != 1 {
		t.Fatalf("預期一次固定種子，實際 %d", len(seedEvents))
	}
	create := func(name string) {
		t.Helper()
		for _, ch := range name {
			press(string(ch), "名字輸入")
		}
		press("Return", "接受名字")
		press("Return", "選人類")
		press("Return", "選男性")
		press("Return", "選第一職業")
		press("Down", "選接受屬性")
		press("Return", "接受屬性")
		press("Return", "接受肖像並提交")
	}
	switch scenario {
	case "full":
		create("NOVA")
		press("Down", "再次選建立角色")
		press("Return", "再次建立角色")
		create("LUNA")
		press("Down", "滿槽選建立角色")
		press("Return", "滿槽建立角色")
		create("THIRD")
		press("Space", "回應滿槽通知")
		press("Escape", "離開建立流程")
	case "duplicate":
		create("Theodoric")
		press("Space", "回應重名提交")
		press("Escape", "離開建立流程")
	case "charset":
		base := o.Save()
		fastCharset = true
		for ch := 32; ch <= 126; ch++ {
			o.Restore(base)
			key := string(rune(ch))
			label := fmt.Sprintf("ASCII U+%04X", ch)
			if strings.ContainsRune("!@#$%^&*()", rune(ch)) {
				// 舊版 BIOS 掃描碼表未列 Shift 數字鍵。字元走原版實際使用的
				// DOS 主控台，BIOS a 喚醒 BDA 輪詢，再 Backspace 移除該 a。
				// 未改 BDA、名字或角色資料；收據保留 int21 實際讀鍵紀錄。
				o.Type(key)
				press("a", label+"：DOS 字元與 BIOS 喚醒")
				press("Backspace", label+"：移除喚醒字母")
			} else {
				press(key, label)
			}
			press("Return", label+"：結束輸入並讀取名字")
			if o.Pending() != 0 {
				t.Fatalf("%s 的 DOS 字元仍未讀走", label)
			}
		}
		o.Restore(base)
		fastCharset = false
		for _, ch := range "MiXeD Name" {
			press(string(ch), "大小寫與空格")
		}
		press("Return", "接受混合名字")
		press("Return", "選人類")
		press("Return", "選男性")
		press("Return", "選第一職業")
		press("Down", "選接受屬性")
		press("Return", "接受屬性")
		press("Return", "接受肖像並提交")
	}
}
