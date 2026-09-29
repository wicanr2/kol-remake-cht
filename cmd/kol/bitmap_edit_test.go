package main

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/bitmapedit"
	"github.com/wicanr2/kol-remake-cht/internal/charcreate"
	"github.com/wicanr2/kol-remake-cht/internal/chardata"
	"github.com/wicanr2/kol-remake-cht/internal/flow"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/save"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

// 正常書本路徑：建一名角色 → EDIT SHIELD 從預設盾牌起編、填滿、離開答 y → 存檔 →
// 重讀名冊，那一名帶著改過的盾牌；圖像沒動，仍是 null（SPEC-036）。存檔失敗整筆退回。
func TestEditShieldSavesAndReloads(t *testing.T) {
	lib, err := assets.Load("../../assets/original-derived-png/ega")
	if err != nil {
		t.Skipf("需要抽取好的 PNG：%v", err)
	}
	data, err := chardata.Load("../../assets/original-derived-data/character-creation.json")
	if err != nil {
		t.Fatal(err)
	}
	start, err := loadCreationStart("../../assets/original-derived-data/creation-start.json")
	if err != nil {
		t.Fatal(err)
	}
	r := &roster.Roster{Slots: make([]roster.Slot, 16)}
	r.Layout.SlotCount = 16
	s := &town.Session{Outdoor: &town.Outdoor{}}
	w := &world.Screen{Travel: &world.Travel{}, Clock: world.NewClock()}
	path := filepath.Join(t.TempDir(), "complete.json")
	persist := func() error {
		out, err := buildCompleteSave(s, w, r, false)
		if err != nil {
			return err
		}
		return save.Write(path, out)
	}
	s.SaveGame = persist
	f := flow.New(r, charcreate.New(data, newRand(1)), roster.NewDeleteFlow(r), nil, s)
	f.CommitCreate = func(c *charcreate.State) error { return commitCreated(c, start, r, s) }
	attachBitmapEditor(f, lib, s)
	for _, key := range []string{"RET", "RET", "5", "D", "RET", "A", "RET", "RET", "RET", "RET", "D", "RET", "RET"} {
		f.Key(key)
	}
	if f.Scene != flow.SceneMenu || len(r.Occupied()) != 1 {
		t.Fatalf("建角沒有提交：%s", f.Scene)
	}
	for f.MenuCursor != flow.MenuEditShield {
		f.Key("D")
	}
	f.Key("RET") // 選角清單
	f.Key("RET") // 第一名 → 編輯器
	if f.Scene != flow.SceneEditor {
		t.Fatalf("應進編輯器，停在 %s（%v）", f.Scene, f.EditError)
	}
	def, err := lib.DecodeDefaultShield()
	if err != nil {
		t.Fatal(err)
	}
	want, err := bitmapedit.FromImage(def)
	if err != nil {
		t.Fatal(err)
	}
	if f.Editor.Pix != want {
		t.Fatal("新角色的盾牌應從預設盾牌起編")
	}
	edit := []string{">", ">", ">", "RET", "<", "<", "<", "RET"}
	// 先試一次存檔失敗：名冊與隊伍都不該留下改動。
	s.SaveGame = func() error { return errors.New("測試磁碟失敗") }
	for _, k := range append(edit, "y") {
		f.Key(k)
	}
	if f.EditError == nil || s.Roster[0].Shield != nil {
		t.Fatalf("存檔失敗應退回：錯誤 %v，盾牌 %v", f.EditError, s.Roster[0].Shield)
	}
	s.SaveGame = persist
	f.Key("RET")
	for _, k := range append(edit, "y") {
		f.Key(k)
	}
	if f.EditError != nil || f.Scene != flow.SceneEditPick {
		t.Fatalf("存檔：%v，停在 %s", f.EditError, f.Scene)
	}
	loaded, err := save.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	c := loaded.Characters.Characters[0]
	if c.Picture != nil || c.Shield == nil {
		t.Fatalf("存檔應只有盾牌：圖像 %v，盾牌 %v", c.Picture != nil, c.Shield != nil)
	}
	got, err := bitmapedit.ParseHex(*c.Shield)
	if err != nil || got.At(17, 23) != 3 || got.At(0, 0) != 3 {
		t.Fatalf("盾牌應整張填成顏色 3：%v", err)
	}
	rr := &roster.Roster{}
	ss := &town.Session{}
	if err = restoreRoster(loaded, rr, ss); err != nil {
		t.Fatal(err)
	}
	if ss.Roster[0].Shield == nil || *ss.Roster[0].Shield != *c.Shield {
		t.Fatal("重讀名冊應帶回改過的盾牌")
	}
	b, err := memberBitmap(lib, ss.Roster[0], true)
	if err != nil || b != got {
		t.Fatalf("重讀後再進編輯器應從改過的那一張起：%v", err)
	}
}
