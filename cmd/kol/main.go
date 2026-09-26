// kol 是《武士傳說》繁體中文重製版的可執行入口。
//
// 目前是 M4 的垂直切片：載入已抽取的 PNG 與 typed data，以原版解析度繪製書本畫面
// 與角色建立流程，並提供截圖模式讓對拍工具取得 remake 側的同狀態畫面。
package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/charcreate"
	"github.com/wicanr2/kol-remake-cht/internal/chardata"
	"github.com/wicanr2/kol-remake-cht/internal/ending"
	"github.com/wicanr2/kol-remake-cht/internal/flow"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/i18n"
	magicdata "github.com/wicanr2/kol-remake-cht/internal/magic"
	portraitdata "github.com/wicanr2/kol-remake-cht/internal/portrait"
	"github.com/wicanr2/kol-remake-cht/internal/rng"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/save"
	"github.com/wicanr2/kol-remake-cht/internal/screen"
	"github.com/wicanr2/kol-remake-cht/internal/title"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

type scene string

const (
	sceneCover      scene = "cover"
	sceneContents   scene = "contents"
	sceneOpenSpread scene = "spread"
	sceneNewChar    scene = "newchar"
	sceneScript     scene = "script"
	sceneTitle      scene = "title"
	sceneScriptCJK  scene = "script-cjk"
	sceneBookOpen   scene = "book-open"
	sceneMenu       scene = "menu"
	sceneRoster     scene = "roster"
	sceneDelete     scene = "delete"
	sceneHarbor     scene = "harbor"
	sceneInn        scene = "inn"
	sceneOutdoor    scene = "outdoor"
	sceneBuilding   scene = "building"
	scenePlay       scene = "play"
	sceneFlow       scene = "flow"
	sceneWorld      scene = "world"
	sceneEnding     scene = "ending"
	// sceneGame 是互動遊玩：一開視窗就從標題序列開始，鍵盤與滑鼠直接玩。
	// 它走的是 `flow` 那條路徑，只是 token 來自輸入裝置而不是腳本，所以與
	// `docs/verification/player-path/` 那幾份收據經過同一組狀態機。
	sceneGame scene = "game"
)

var scenes = []scene{sceneCover, sceneContents, sceneOpenSpread, sceneNewChar, sceneScript, sceneTitle, sceneScriptCJK, sceneBookOpen, sceneMenu, sceneRoster, sceneDelete, sceneHarbor, sceneInn, sceneOutdoor, sceneBuilding, scenePlay, sceneFlow, sceneWorld, sceneEnding}

// lcg 是可重播的亂數來源，內容就是原版那一支（`internal/rng`，RE-132）：
// 同一個種子在 remake 與原版會擲出同一串數，`-seed N` 等於原版的 `srand(N)`。
type lcg = rng.LCG

func newRand(seed int) *lcg { return rng.New(seed) }

// questBattle 把那一場戰鬥包起來，讓建構好的 game 與後面才設定的回呼共用同一份。
type questBattle struct {
	state *battle.State
	lines []string
	// bank 是 KOLDATA，執行段的訊息要照記錄號取文字（RE-086）。
	bank *town.UIText
	// trans 是譯文表；中文模式的戰鬥訊息從這裡取，取不到才退回原文。
	trans *town.Translations
	// english 為真時才補英文複數；中文的譯文自己處理掉了複數（RE-091）。
	english bool
	// tiles 是戰場磚圖，讀檔一次之後留著。
	tiles *ui.BattleTiles
	// awarded 記這一場的冒險點數與金冠發過了沒有——收尾那一段每按一次鍵都會
	// 走一遍，發放只能有一次。
	awarded bool
	// 戰利品清單只在勝利結算時開一次。玩家從物品頁取消或領完後，
	// 下一次按鍵要返回野外，不能把地上未領物品重新列出（RE-115）。
	spoilsOpened bool
}

// battleMember 把城鎮那一側的隊員換成戰鬥要用的欄位（RE-080）：戰鬥等級由
// 四列武器的攻防原始值算出來，規劃順序要洞察，昏迷門檻要體型與健康。
func battleMember(m town.Member) battle.Member {
	out := battle.Member{Name: m.Name, Wounds: m.Wounds, Female: m.Female, Portrait: m.Portrait}
	if m.Sheet == nil {
		return out
	}
	for _, w := range m.Sheet.Weapons {
		out.Skills = append(out.Skills, w.OffenceRaw, w.DefenceRaw)
		// 整份帶著：手上換一件別類的武器就換一列（`sub_26AC2` 每次取用都重查）。
		out.WeaponRows = append(out.WeaponRows, battle.WeaponSkillRow{
			Class: w.Weapon, Offence: w.OffenceRaw, Defence: w.DefenceRaw})
	}
	out.Foresight = m.Sheet.Stats[town.StatForesight]
	out.Size = m.Sheet.Stats[town.StatSize]
	out.Health = m.Sheet.Stats[town.StatHealth]
	out.Strength = m.Sheet.Stats[town.StatStrength]
	out.Quickness = m.Sheet.Stats[town.StatQuickness]
	out.Nutrition = m.Sheet.Nutrition
	out.Intellect = m.Sheet.Stats[town.StatIntellect]
	out.Ward = m.Sheet.Ward
	// 手上那件武器（角色記錄 0x30，不在穿戴欄裡）與三件護甲、盾牌的重量：
	// 攻擊值、負重與骰子都要用。
	if w := m.InHand; w != nil {
		out.HandItem = battleItem(w)
		out.HasWeapon = true
		out.Weapon = battle.Weapon{
			Kind: w.Class, Base: w.Dice.Bonus, Code: dieCode(w.Record),
			Weight: w.Encumbrance, TwoHand: w.Hands == 1,
		}
		out.Weapon.Count, out.Weapon.Sides = battle.DamageCode(out.Weapon.Code)
		out.WeaponName = w.Name
		// 手上武器的種類決定用哪一列武器技能（RE-080）。
		for _, row := range m.Sheet.Weapons {
			if row.Weapon == w.Class {
				out.Attack, out.Parry = row.OffenceRaw, row.DefenceRaw
				break
			}
		}
	}
	// 任務旗標（角色記錄 0x1E2 起那六個 byte）：打贏任務戰鬥要原地標成完成
	// （RE-101）。這裡共用同一個 slice，改動因此會回到城鎮那一側的隊員身上。
	out.Quests = m.Quests
	// 法術（角色記錄 0x200 起那十六筆，RE-100）：施法要用，指令列上那一格也看它。
	for _, sp := range m.Spells {
		out.Spells = append(out.Spells, battle.Spell{Name: sp.Name,
			Effect: sp.Effect, TurnCode: sp.TurnCode, PowerCode: sp.PowerCode,
			Guards: spellGuards(sp.Guards)})
	}
	out.PackCount = len(out.Spells)
	// 背包六格（角色記錄 0xA0 起，RE-075）：戰利品優先往這裡放。
	for i, g := range m.Backpack {
		if i >= battle.BackpackSlots {
			break
		}
		out.Pack[i] = g.Name
		out.PackItems[i] = battleItem(&g)
	}
	// 腰帶是角色記錄 +0x190；交換時必須連同武器數值一起搬移（RE-094／127）。
	if w := m.Worn[9]; w != nil {
		out.BeltItem = battleItem(w)
		out.HasSecondItem, out.BeltName = true, w.Name
		out.BeltWeapon = battle.Weapon{Kind: w.Class, Base: w.Dice.Bonus,
			Code: dieCode(w.Record), Weight: w.Encumbrance, TwoHand: w.Hands == 1}
		out.BeltWeapon.Count, out.BeltWeapon.Sides = battle.DamageCode(out.BeltWeapon.Code)
		out.Worn[5] = w.Name
	}
	for _, pos := range battle.ArmourSlots {
		if g := m.Worn[pos]; g != nil {
			out.Armour[pos] = battleItem(g)
			out.ArmourWeight += g.Encumbrance
		}
	}
	// 徒手四式各拿一格穿戴品當武器（`UnarmedSlots`，RE-131）：踢用 FEET、
	// 盾擊用 SHIELD、頭槌用 HEAD、拳用 HANDS。那一格空著就是空的武器，
	// 傷害只剩力量加成——與原版讀到一筆全零的記錄一樣。
	for style, pos := range battle.UnarmedSlots {
		if pos >= len(m.Worn) {
			continue
		}
		g := m.Worn[pos]
		if g == nil {
			continue
		}
		w := battle.Weapon{Kind: g.Class, Base: g.Dice.Bonus, Code: dieCode(g.Record),
			Weight: g.Encumbrance, TwoHand: g.Hands == 1}
		w.Count, w.Sides = battle.DamageCode(w.Code)
		out.UnarmedWeapon[style] = w
	}
	if sh := m.Worn[battle.ShieldSlot]; sh != nil {
		out.Worn[4] = sh.Name
		out.ShieldWeight = sh.Encumbrance
		out.ShieldGrade = sh.SizeClass
	} else {
		out.ShieldGrade = battle.NoShield
	}
	return out
}

// dieCode 從物品記錄的 hex 取回 +0x0C 的傷害碼。
func dieCode(record string) int {
	if len(record) < 26 {
		return 0
	}
	v, err := strconv.ParseUint(record[24:26], 16, 8)
	if err != nil {
		return 0
	}
	return int(v)
}

// nonEmptyLines 去掉記錄拆行之後的空列。
func nonEmptyLines(lines []string) []string {
	out := lines[:0:0]
	for _, line := range lines {
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// lastLine 是一筆記錄拆行之後最後那一行非空的字。問句記錄前面常常是空列
// （`0xE3` 開頭那個 `@` 就是敘述與問句之間的空白），要的是後面那一句。
func lastLine(lines []string) string {
	out := ""
	for _, line := range lines {
		if line != "" {
			out = line
		}
	}
	return out
}

// partyHasQuest 回報隊上有沒有人帶著這個任務地點的任務（RE-079）。
func partyHasQuest(party []town.Member, site *battle.Site) bool {
	for _, m := range party {
		if site.Taken(m.Quests) {
			return true
		}
	}
	return false
}

// partyMovement 是隊伍裡最慢的那個移動率（`sub_29004:0x290A6`…`0x290D7`）。
// 數字越大越慢，0 當成 20。沒有資料頁的隊員不算。
func partyMovement(party []town.Member) int {
	slowest := 0
	for _, m := range party {
		if m.Sheet == nil {
			continue
		}
		rate := m.Sheet.MovementRate
		if rate == 0 {
			rate = world.DefaultMovement
		}
		if rate > slowest {
			slowest = rate
		}
	}
	return slowest
}

// fixedChoice 固定取同一個選項，用來窮舉標題序列兩處亂數的每一條分支。
// 對拍的判準是集合包含，因此兩次固定取法的聯集就涵蓋所有會出現的畫面。
type fixedChoice int

func (c fixedChoice) Choose(n int) int {
	if int(c) >= n {
		return n - 1
	}
	return int(c)
}

// canvasFor 依語言決定畫布大小：中文模式在放大後的畫布上作畫。
func canvasFor(lang string) *screen.Canvas {
	if lang == "en" {
		return screen.NewCanvas()
	}
	return screen.NewCanvasScaled(book.CJKScale)
}

// rgbaFrom 把抽出的 PNG 解成 RGBA，供 CPU 端的動畫重播使用。
func rgbaFrom(lib *assets.Library, archive string, index int) (*image.RGBA, error) {
	src, _, err := lib.Decode(archive, index)
	if err != nil {
		return nil, err
	}
	out := image.NewRGBA(src.Bounds())
	draw.Draw(out, out.Bounds(), src, src.Bounds().Min, draw.Src)
	return out, nil
}

// writeRGBA 把畫面寫成 PNG；對拍工具讀的就是這個檔案。
func writeRGBA(path string, img *image.RGBA) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// runCharacterWrite 用 remake 的 writer 產出一筆槽位記錄，供與原版實際寫入逐 byte 對拍。
// 樣板是本機的原版 CHARDATA，不入庫；輸出也只作本機比對。
func runCharacterWrite(outPath, templatePath string, slot int, contractPath,
	itemPath, chardataOut string, c chardata.NewCharacter, stats string) error {
	if templatePath == "" {
		return fmt.Errorf("需要 -chardata-template")
	}
	raw, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}
	if len(raw) != chardata.RecordBytes*chardata.SlotCount {
		return fmt.Errorf("樣板長度 %d 不是 %d", len(raw),
			chardata.RecordBytes*chardata.SlotCount)
	}
	if slot < 0 || slot >= chardata.SlotCount {
		return fmt.Errorf("槽位 %d 超出範圍", slot)
	}
	parts := strings.Split(stats, ",")
	if len(parts) != 7 {
		return fmt.Errorf("-chardata-stats 需要七個值")
	}
	for i, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return fmt.Errorf("屬性 %d 不是整數：%w", i, err)
		}
		c.Primary[i] = v
	}
	contracts, err := chardata.LoadContracts(contractPath)
	if err != nil {
		return err
	}
	contract, ok := contracts.Get(c.ProfessionIndex)
	if !ok {
		return fmt.Errorf("沒有職業記錄 %d 的起始契約", c.ProfessionIndex)
	}
	items, err := chardata.LoadWeaponTemplates(itemPath)
	if err != nil {
		return err
	}
	base := slot * chardata.RecordBytes
	record, err := chardata.WriteRecord(raw[base:base+chardata.RecordBytes], c,
		contract, items)
	if err != nil {
		return err
	}
	doc := map[string]any{
		"schema": "kol-chardata-write-v1", "slot": slot,
		"character": map[string]any{"name": c.Name, "race": c.Race,
			"secondary": c.Secondary, "profession_index": c.ProfessionIndex,
			"primary_internal": c.Primary},
		"contract":        contract,
		"written_offsets": chardata.WrittenOffsets(),
		"record_hex":      hex.EncodeToString(record),
	}
	blob, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(outPath, append(blob, '\n'), 0o644); err != nil {
		return err
	}
	if chardataOut != "" {
		updated := make([]byte, len(raw))
		copy(updated, raw)
		copy(updated[base:base+chardata.RecordBytes], record)
		if err := os.WriteFile(chardataOut, updated, 0o644); err != nil {
			return err
		}
	}
	return nil
}

type spriteCache struct {
	lib *assets.Library
	m   map[string]*ebiten.Image
}

func (c *spriteCache) Get(archive string, index int) (*ebiten.Image, error) {
	key := fmt.Sprintf("%s/%d", archive, index)
	if img, ok := c.m[key]; ok {
		return img, nil
	}
	src, _, err := c.lib.Decode(archive, index)
	if err != nil {
		return nil, err
	}
	img := ebiten.NewImageFromImage(src)
	c.m[key] = img
	return img, nil
}

// applyPlayToken 把腳本 token 套到城鎮狀態機；token 對應原版的鍵盤操作。
// battleSteps 是戰場上方向鍵的位移：列是原版的 X、欄是 Y，
// 斜走鍵沿用參考卡的 A／S／Z／X（SPEC-010）。
var battleSteps = map[string][2]int{
	"U": {-1, 0}, "D": {1, 0}, "L": {0, -1}, "R": {0, 1},
	"A": {-1, -1}, "S": {-1, 1}, "Z": {1, -1}, "X": {1, 1},
}

// applyBattleToken 把腳本 token 交給戰鬥。遭遇那一張任何鍵進戰場；戰場上依
// 規劃段停在哪一層決定方向鍵是移游標還是沒有作用，`<`／`>` 移圖示列的白框，
// Return 是確定（SPEC-010）。整輪決定完停在「Begin next round?」，那一層沿用
// 既有的是／否；答應之後由呼叫端給的回呼推進執行段。
// loadTranslations 一次載入城鎮譯文與 UI 文字庫譯文。兩份分開存檔——城鎮譯文以
// 「城／門／欄位」索引，UI 譯文以文字庫的記錄號索引，混在一起會讓兩邊都難查。
func loadTranslations(townPath, uiPath string) (*town.Translations, error) {
	trans, err := town.LoadTranslations(townPath)
	if err != nil {
		return nil, err
	}
	ui, err := town.LoadUITranslations(uiPath)
	if err != nil {
		return nil, err
	}
	trans.UI = ui
	return trans, nil
}

func (g *game) applyBattleToken(s *battle.State, token string,
	step func(*battle.State)) error {
	if s.Stage == battle.StageEncounter {
		s.Continue(s.Rand)
		return nil
	}
	if s.Stage == battle.StageSpoils {
		switch token {
		case "U":
			s.SpoilMove(-1)
		case "D":
			s.SpoilMove(1)
		case "<":
			s.SpoilIconMove(-1)
		case ">":
			s.SpoilIconMove(1)
		case "RET", "SPACE":
			// 白框停在 U-TURN 就是取消，與原版的回傳值判讀相同（RE-115）。
			if _, _, done := s.SpoilAccept(); done {
				s.Stage = battle.StageDone
			}
		case "N", "ESC":
			// 原版另外看 `word_2DF20`（ESC），與圖示無關。
			if s.SpoilCancel() {
				s.Stage = battle.StageDone
			}
		}
		return nil
	}
	if s.Plan.Prompt == battle.PromptNextRound {
		switch token {
		case "RET", "SPACE":
			s.Answer(true)
			step(s)
		case "N":
			s.Answer(false)
		}
		return nil
	}
	if s.Plan.Prompt == battle.PromptFlee {
		// 原版直接比字母鍵（`0x21158`…`0x21190`），別的鍵重問。
		switch token {
		case "Y":
			s.Answer(true)
			step(s)
		case "N":
			s.Answer(false)
		}
		return nil
	}
	// 撿那一問同樣只認 y／n（`0x2036E` 的 `sub_1B100(5)`，RE-113）。
	if s.Plan.Prompt == battle.PromptPickUp {
		switch token {
		case "Y":
			s.AnswerPickUp(true)
			step(s)
		case "N":
			s.AnswerPickUp(false)
			step(s)
		}
		return nil
	}
	// 挑部位那一問只認 1…6（`0x284D8` 的 `sub_1B100(6)`，RE-100）。
	if s.Plan.Prompt == battle.PromptArea {
		if len(token) == 1 && token[0] >= '1' && token[0] <= '6' {
			if res, caster, ok := s.AnswerArea(int(token[0] - '1')); ok {
				g.battleAreaLine(s, res, caster)
			}
		}
		return nil
	}
	if s.Plan.Prompt == battle.PromptExecuting {
		switch token {
		case "RET", "SPACE":
			step(s)
		}
		return nil
	}
	if step, ok := battleSteps[token]; ok {
		s.MoveCursor(step[0], step[1], false)
		return nil
	}
	// 滑鼠指格（fieldCellTargets）：換成相對於游標的位移，照原版走同一段夾限。
	if cell, ok := strings.CutPrefix(token, "CELL:"); ok {
		var r, c int
		if _, err := fmt.Sscanf(cell, "%d,%d", &r, &c); err == nil && s.Plan.Prompt == battle.PromptTarget {
			s.MoveCursor(r-s.Plan.Cursor.Row, c-s.Plan.Cursor.Col, true)
		}
		return nil
	}
	switch token {
	case "<":
		s.MoveSelection(-1)
	case ">":
		s.MoveSelection(1)
	case "RET", "SPACE":
		if token == "SPACE" && (s.Plan.Prompt == battle.PromptCommand || s.Plan.Prompt == battle.PromptTarget) {
			s.ToggleEquipmentPanel()
		} else {
			s.Choose()
		}
	}
	return nil
}

func applyPlayToken(s *town.Session, token string) error {
	if !s.Key(token) {
		return fmt.Errorf("未知的城鎮操作 token：%q", token)
	}
	return nil
}

// playToken 把腳本 token 交給目前掌管畫面的那個狀態機：人在城裡就是城鎮，
// 走出城門之後就是野外。
func (g *game) playToken(token string) error {
	// 結局蓋在最上面：回報滿 0x17 件之後按鍵是翻結局的段落，不是走路（RE-106）。
	if g.endingReached() {
		g.ending.Advance()
		return nil
	}
	// 戰鬥蓋在最上面。遭遇那一張按一個鍵進戰場，接著就是規劃段的指令列。
	if g.quest != nil && g.quest.state != nil {
		return g.applyBattleToken(g.quest.state, token,
			func(*battle.State) { g.battleStep() })
	}
	// 競技場的對話走城鎮那一層的鍵位（RE-108）；談完就開那一場。
	if g.play != nil && g.play.InArena() {
		if err := applyPlayToken(g.play, token); err != nil {
			return err
		}
		g.startArenaFight()
		return nil
	}
	if g.play != nil && g.play.LeftTown && g.worldSession != nil {
		return applyWorldToken(g.worldSession, token)
	}
	return applyPlayToken(g.play, token)
}

// awardBattle 把打贏那一場的冒險點數與金冠寫回隊伍（`0x2141C`，RE-096）：
// 對隊伍那六格裡有人的每一名各加一份——兩個數字本來就是「每人分到多少」。
// 競技場與訓練戰整段跳過（`s.RankFight`），一場只發一次。
func (g *game) awardBattle() {
	s := g.quest.state
	if g.quest.awarded || s.Verdict != battle.VerdictWon || s.RankFight {
		return
	}
	g.quest.awarded = true
	for i := range g.play.Party {
		g.play.Party[i].Gold += s.Gold
		if sh := g.play.Party[i].Sheet; sh != nil {
			sh.AdventurePoints += s.Adventure
		}
	}
}

// startArenaFight 把競技場談完的那一場開起來（RE-108）：一隻、等級是上場那一名的
// 實得位階加一、陣型與種類各擲一次。站位不擲也不查表——競技場走的是
// `sub_1E3A8` 對 `word_345DE >= 0` 的那一條（RE-098），只有上場那一名參戰。
// 地形依 SPEC-025 載入第 24 號戰場圖，再走共同的樹叢散布。
func (g *game) startArenaFight() {
	f := g.play.PendingFight()
	if f == nil || g.quest == nil || g.arena == nil {
		return
	}
	g.play.ClearPendingFight()
	rng := g.arena.rand
	// 擲的順序照原版（`0x2AFF9` 先陣型、`0x2B012` 才是種類）。陣型這一場用不到
	// ——站位是固定的——但那一次擲骰要擲，否則之後的亂數序列會偏掉。
	rng.Roll(battle.AmbushGroups)
	kind := rng.Roll(battle.ArenaMonsterKinds) - 1
	members := make([]battle.Member, 0, len(g.play.Party))
	for _, m := range g.play.Party {
		members = append(members, battleMember(m))
	}
	mon := g.arena.monsters.At(kind)
	field := &battle.Field{Tiles: append([]int(nil), g.arena.mapField.Tiles...)}
	field.Scatter(g.arena.field)
	st := battle.NewArena(f.Member, kind, f.Level, mon, members, field)
	if st == nil {
		return
	}
	st.Rand = g.arena.rand
	st.SpoilItems = g.arena.spoils
	g.quest.tiles = g.arena.tiles
	g.quest.state = st
	g.quest.bank = g.arena.bank
	g.quest.trans = g.arena.trans
	g.quest.english = g.arena.english
	// 遭遇那一張的文字：原版在這裡把 `word_34968` 設成 1，走的是任務地點那一條，
	// 所以用同一句（記錄 0xE9）。**這一句本身沒有從程式碼讀出來**（RE-108 沒有
	// 讀到競技場的遭遇訊息），是照那個旗標推的。
	g.quest.lines = nonEmptyLines(g.arena.bank.Lines(town.BankUI,
		battle.QuestCreaturesRecord))
	g.arenaFighter = f.Member
	g.arenaBets = f.Bets
	g.play.LeaveArena()
}

// savedParty 把存檔裡的隊伍換成城鎮這一側的型別。角色的靜態資料（頭像、屬性）
// 仍然來自 typed data，存檔只帶會變的那些欄位（SPEC-006）。
func savedParty(s *save.Save) []town.Member {
	if s.Characters != nil {
		return completeParty(s)
	}
	out := make([]town.Member, 0, len(s.Party))
	for _, m := range s.Party {
		out = append(out, town.Member{
			Name: m.Name, Portrait: m.Portrait, Female: m.Female,
			Wounds: m.Wounds, Gold: m.Gold, Place: m.Place,
			TownCode: m.TownCode, BuildingID: m.BuildingID,
			Renown: m.Renown, Spellcaster: m.Spellcaster, Speaker: m.Speaker,
			Quests:   append([]int(nil), m.Quests...),
			Progress: append([]int(nil), m.Progress...),
			Carried:  decodeCarried(m.Carried),
			Steed:    townSteed(m.Steed),
			Spells:   townSavedSpells(m.Spells),
			Order:    townOrder(m.Order),
		})
	}
	return out
}

// townSavedSpells／savedSpells 換算法術那一欄。存檔只帶規則層用得到的欄位。
func townSavedSpells(rows []save.Spell) []town.Spell {
	if len(rows) == 0 {
		return nil
	}
	out := make([]town.Spell, 0, len(rows))
	for _, r := range rows {
		out = append(out, town.Spell{Name: r.Name, Effect: r.Effect,
			TurnCode: r.TurnCode, PowerCode: r.PowerCode,
			Guards: append([]int(nil), r.Guards...)})
	}
	return out
}

func savedSpells(rows []town.Spell) []save.Spell {
	if len(rows) == 0 {
		return nil
	}
	out := make([]save.Spell, 0, len(rows))
	for _, r := range rows {
		out = append(out, save.Spell{Name: r.Name, Effect: r.Effect,
			TurnCode: r.TurnCode, PowerCode: r.PowerCode,
			Guards: append([]int(nil), r.Guards...)})
	}
	return out
}

// townOrder／savedOrder 換算所屬教派那一欄（RE-121）。
func townOrder(o *save.Order) *town.Order {
	if o == nil {
		return nil
	}
	return &town.Order{Name: o.Name, Title: o.Title, Intellect: o.Intellect,
		Allow:  append([]int(nil), o.Allow...),
		Record: o.Record, Token: o.Token, TokenRecord: o.TokenRecord}
}

func savedOrder(o *town.Order) *save.Order {
	if o == nil {
		return nil
	}
	return &save.Order{Name: o.Name, Title: o.Title, Intellect: o.Intellect,
		Allow:  append([]int(nil), o.Allow...),
		Record: o.Record, Token: o.Token, TokenRecord: o.TokenRecord}
}

// townSteed／savedSteed 換算坐騎那一欄。城鎮那一側用的是商品記錄，存檔只留
// 讀得出來的那三欄（RE-119）。
func townSteed(s *save.Steed) *town.Good {
	if s == nil {
		return nil
	}
	return &town.Good{Index: s.Index, Name: s.Name, Price: s.Price,
		Picture: s.Picture, Movement: s.Movement}
}

func savedSteed(g *town.Good) *save.Steed {
	if g == nil {
		return nil
	}
	return &save.Steed{Index: g.Index, Name: g.Name, Price: g.Price,
		Picture: g.Picture, Movement: g.Movement}
}

// decodeCarried 把存檔裡那十七格 hex 換回位元組；長度不對的一格當成空的。
func decodeCarried(rows []string) [][]byte {
	if len(rows) == 0 {
		return nil
	}
	out := make([][]byte, 0, len(rows))
	for _, row := range rows {
		raw, err := hex.DecodeString(row)
		if err != nil || len(raw) != town.CarriedStride {
			raw = make([]byte, town.CarriedStride)
		}
		out = append(out, raw)
	}
	return out
}

// encodeCarried 是 decodeCarried 的反向。
func encodeCarried(rows [][]byte) []string {
	if len(rows) == 0 {
		return nil
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, hex.EncodeToString(row))
	}
	return out
}

// townSheet 把角色檔抽出來的資料頁欄位換成城鎮這一側的型別。
func townSheet(s *roster.Sheet) *town.Sheet {
	if s == nil || len(s.Stats) != town.StatCount {
		return nil
	}
	out := &town.Sheet{
		Race: s.Race, Profession: s.Profession, Rank: s.Rank,
		AdventurePoints: s.AdventurePoints, MovementRate: s.MovementRate,
		Nutrition: s.Nutrition, Ward: s.Ward, HeightFeet: s.HeightFeet,
		HeightInches: s.HeightInches, Weight: s.Weight,
		Order: s.Order, Title: s.Title,
	}
	copy(out.Stats[:], s.Stats)
	for _, w := range s.Weapons {
		out.Weapons = append(out.Weapons, town.WeaponRow{
			Row: w.Row, Name: w.Name, Weapon: w.Weapon,
			Offence: w.Offence, Defence: w.Defence,
			OffenceRaw: w.OffenceRaw, DefenceRaw: w.DefenceRaw})
	}
	return out
}

// spellGuards 把擋招那四格湊成固定長度的陣列；缺的補零。
func spellGuards(g []int) [4]int {
	var out [4]int
	for i := 0; i < len(g) && i < len(out); i++ {
		out[i] = g[i]
	}
	return out
}

// townSpells 把角色槽位裡那十六格法術換成城鎮那一側的型別（RE-100）。
// 空的格子不收進來，所以清單的長度就是「會幾個法術」。
func townSpells(items []*roster.Item) []town.Spell {
	var out []town.Spell
	for _, it := range items {
		if it == nil {
			continue
		}
		out = append(out, town.Spell{Name: it.Name, Effect: it.Effect,
			TurnCode: it.TurnCode, PowerCode: it.PowerCode, Guards: it.Guards})
	}
	return out
}

// townGood 把角色槽位裡的一件東西換成城鎮那一側的型別（RE-080）。
func townGood(it *roster.Item) *town.Good {
	if it == nil {
		return nil
	}
	g := town.Good{
		Name: it.Name, Price: it.Price, Kind: it.Kind, Damage: it.Damage,
		Hands: it.Hands, Class: it.Class, Position: it.Position,
		Encumbrance: it.Encumbrance, Condition: it.Condition,
		Percent: it.Percent, SizeClass: it.SizeClass, Record: it.Record,
	}
	g.Dice = town.Dice{Count: it.Dice.Count, Sides: it.Dice.Sides,
		Bonus: it.Dice.Bonus, Low: it.Dice.Low, High: it.Dice.High}
	return &g
}

// townWorn 把穿戴欄十格換過來。第十格是腰帶，手上那件另外走 townGood（RE-127）。
func townWorn(items []*roster.Item) [town.WornSlots]*town.Good {
	var out [town.WornSlots]*town.Good
	for i, it := range items {
		if i >= town.WornSlots {
			break
		}
		out[i] = townGood(it)
	}
	return out
}

// buildSave 把目前的遊玩狀態組成一份存檔（SPEC-006）。人在城裡就記城鎮座標，
// 走到野外就記世界座標；兩邊都只放有 RE 證據的欄位。
func buildSave(s *town.Session, w *world.Screen, faithful bool) *save.Save {
	out := save.New()
	out.SavedAt = time.Now().UTC().Format(time.RFC3339)
	out.Settings.FaithfulInnStay = faithful
	if w != nil {
		t := w.Travel
		out.World = save.World{X: t.X, Y: t.Y, SubX: t.SubX, SubY: t.SubY}
	}
	if s.LeftTown {
		out.Location = save.Location{Kind: save.LocationWorld}
	} else {
		x, y := s.Outdoor.PartyPosition()
		out.Location = save.Location{Kind: save.LocationTown,
			Town: s.Outdoor.Town, EntryPoint: s.Outdoor.EntryPoint, X: x, Y: y}
	}
	for _, m := range s.Party {
		out.Party = append(out.Party, save.Member{
			Name: m.Name, Portrait: m.Portrait, Female: m.Female,
			Wounds: m.Wounds, Gold: m.Gold, Place: m.Place,
			TownCode: m.TownCode, BuildingID: m.BuildingID,
			Renown: m.Renown, Spellcaster: m.Spellcaster, Speaker: m.Speaker,
			Quests:          append([]int(nil), m.Quests...),
			Progress:        append([]int(nil), m.Progress...),
			Carried:         encodeCarried(m.Carried),
			Steed:           savedSteed(m.Steed),
			Spells:          savedSpells(m.Spells),
			Order:           savedOrder(m.Order),
			Rank:            savedRank(m),
			AdventurePoints: savedPoints(m),
		})
	}
	// 存檔裡的發言者要明確：畫面上沒有標記時發言者就是隊長，寫進去也照這個規則，
	// 讀回來才不會變成「沒有人在說話」。
	if i := out.Speaker(); i >= 0 && !out.Party[i].Speaker {
		out.Party[i].Speaker = true
	}
	return out
}

// parseSteps 解析 `-shot-steps` 的步號清單。空字串回 nil，代表每一步都截。
func parseSteps(spec string) (map[int]bool, error) {
	if strings.TrimSpace(spec) == "" {
		return nil, nil
	}
	out := map[int]bool{}
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("-shot-steps 的 %q 不是步號：%w", part, err)
		}
		out[n] = true
	}
	return out, nil
}

// savedRank 取那一名的帳面位階；沒有資料頁就不寫這一欄。
func savedRank(m town.Member) *int {
	if m.Sheet == nil {
		return nil
	}
	rank := m.Sheet.Rank
	return &rank
}

// savedPoints 同上，取的是冒險點數。
func savedPoints(m town.Member) *int {
	if m.Sheet == nil {
		return nil
	}
	points := m.Sheet.AdventurePoints
	return &points
}

// parsePair 解析 `X,Y` 形式的參數。
func parsePair(spec string) (int, int, error) {
	var a, b int
	if _, err := fmt.Sscanf(spec, "%d,%d", &a, &b); err != nil {
		return 0, 0, err
	}
	return a, b, nil
}

// enterWorld 把隊伍放到世界地圖上的某一格。城門選 DOOR 之後就是走這一條。
func enterWorld(s *world.Screen, p town.Point) {
	if s == nil {
		return
	}
	s.Travel = world.New(s.Travel.Data, s.Travel.Grid, p.X, p.Y)
}

// applyWorldToken 把腳本 token 套到野外狀態機。
func applyWorldToken(s *world.Session, token string) error {
	// 換日那一張停在原地等一個鍵，按什麼都一樣（`sub_297F2:0x298E2` 的 `sub_144C3`）。
	if s.Camping && token != "TICK" {
		s.EndCamping()
		return nil
	}
	// 騎走成功／失敗那一句同樣只等一個鍵（`0x2A6EF` 與 `0x2A70A` 的 `sub_144C3`）。
	if s.Prompt != nil && s.Prompt.Kind == world.PromptNotice && token != "TICK" {
		s.DismissNotice()
		return nil
	}
	switch token {
	case "L":
		s.Walk(-1, 0)
	case "R":
		s.Walk(1, 0)
	case "U":
		s.Walk(0, -1)
	case "D":
		s.Walk(0, 1)
	case "A":
		s.Walk(-1, -1)
	case "S":
		s.Walk(1, -1)
	case "Z":
		s.Walk(-1, 1)
	case "X":
		s.Walk(1, 1)
	case "<":
		s.MoveCursor(-1)
	case ">":
		s.MoveCursor(1)
	case "RET":
		s.Accept()
	case "TICK":
		s.Screen.Phase++
	default:
		return fmt.Errorf("未知的野外操作 token：%q", token)
	}
	return nil
}

// parseIcons 解析 -outdoor-icons：`none` 是空列，其餘是逗號分隔的 KOLMPICS 圖號。
func parseIcons(spec string) ([]int, error) {
	if spec == "none" {
		return nil, nil
	}
	var out []int
	for _, field := range strings.Split(spec, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(field))
		if err != nil {
			return nil, fmt.Errorf("-outdoor-icons 的 %q 不是圖號：%w", field, err)
		}
		out = append(out, n)
	}
	return out, nil
}

// loadGrid 從資產目錄讀出某城鎮的地圖格子陣列（8-bit 灰階 PNG，像素值即磚塊編號）。
func loadGrid(lib *assets.Library, archive string) (*town.Grid, error) {
	src, _, err := lib.Decode(archive, 0)
	if err != nil {
		return nil, err
	}
	return town.LoadGrid(src)
}

// loadBattlePNG 讀戰場用的 PNG。這一批是 `tools/extract_battle_tiles.py` 的輸出，
// 有自己的 manifest，不在主 manifest 的 archive 索引裡（RE-079）。
func loadBattlePNG(path string) (image.Image, error) {
	body, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("讀取戰場素材：%w", err)
	}
	defer body.Close()
	img, err := png.Decode(body)
	if err != nil {
		return nil, fmt.Errorf("解碼戰場素材 %s：%w", path, err)
	}
	return img, nil
}

type shotJob struct {
	name  string
	path  string
	apply func()
	// skip 為真時這一步不寫 PNG，只跑 apply 並記一筆軌跡。長路徑不需要
	// 每一步都留一張圖，留下的是檢查點那幾張。
	skip bool
}

type game struct {
	notice       []string
	worldState   *world.Screen
	worldSession *world.Session
	// quest 是走到任務地點之後開起來的那一場戰鬥；用指標是因為建構之後才會被填。
	quest *questBattle
	// battleAtlas 是戰場磚圖。
	battleAtlas *ebiten.Image
	worldParty  []town.Member
	sprites     *spriteCache
	small       *font.Font
	large       *font.Font
	cjk         *font.Font
	// titleSubs 是中文模式的開場字幕與封面副標；英文模式為 nil。
	titleSubs  *titleSubtitles
	strings    *i18n.Table
	roster     *roster.Roster
	remove     *roster.DeleteFlow
	inn        *town.Inn
	outdoor    *town.Outdoor
	play       *town.Session
	flow       *flow.Flow
	walkTick   int
	markerStep int
	canvas     *screen.Canvas
	state      *charcreate.State
	sequence   *title.Sequence
	player     *title.Player
	book       *book.OpenPlayer
	bookAnim   *book.OpenAnimation
	bookWait   float64
	bookFrom   string // 動畫的起始畫面
	bookTarget string // 動畫的目標畫面
	bookLib    *assets.Library

	// ending 是結局播到第幾段（RE-106）；endBank／endTrans 是敘事的文字庫與譯文。
	ending   *ending.State
	endBank  *town.UIText
	endTrans *town.Translations

	// interactive 打開時 Update 讀鍵盤與滑鼠，而不是逐步跑 jobs 截圖。
	// logInput 把收到的 token 印出來——輸入沒反應時要看得到是「沒收到鍵」
	// 還是「收到了但狀態機不理它」，這兩者的修法完全不同。
	interactive bool
	logInput    bool

	scene    scene
	portrait int
	selected int
	frame    title.Frame

	jobs  []shotJob
	jobAt int
	drawn bool

	// tracePath 是狀態軌跡的輸出檔；每截一張圖就記一筆目前站在哪一層、位置與階段。
	// 截圖只證明畫面長什麼樣，軌跡才證明這一步走的是哪一個狀態機。
	tracePath string
	// battleCanvas 保留原版戰場解析度，再配合目前畫布縮放。
	battleCanvas *ebiten.Image
	battleBody   ui.BattleBody
	battleSheet  ui.BattleSheet
	traceRows    []map[string]any
	// lastParty 是上一步寫進軌跡的隊伍摘要，用來省掉沒有變化的那幾行；
	// lastStage 是上一步的戰鬥階段，只有它變了才記一次戰場。
	lastParty []string
	lastStage int
	// traceDetail 打開時軌跡多記指令列、游標與逐步的戰場——產生腳本用，
	// 收據不需要。
	traceDetail bool

	// endingArmed 是「該播結局了」。任務戰鬥打完那一刻才可能設起來（RE-106）。
	endingArmed bool

	// arena 是競技場開一場要用的東西；arenaFighter 是上場的那一名（−1 表示
	// 這一場不是競技場打的）。
	arena        *arenaDeps
	arenaFighter int
	// arenaBets 是那一場的賭注，索引同隊伍。
	arenaBets []int
}

// arenaDeps 是競技場開一場戰鬥要用到的東西，建構時湊齊。
type arenaDeps struct {
	mapField *battle.Field
	monsters *battle.Monsters
	spoils   []battle.SpoilItem
	tiles    *ui.BattleTiles
	bank     *town.UIText
	trans    *town.Translations
	rand     *lcg
	field    *lcg
	english  bool
}

// 軌跡裡要取的那兩欄。`snapshot` 裡的區域變數叫 `town`，會把套件名遮起來，
// 所以先在這裡取出來。
var (
	spellNameField = town.SpellFieldName
	spellStatField = town.SpellFieldStat
)

// snapshot 記下這一步之後遊戲停在哪。欄位只取「能證明走到哪一層」的那幾個：
// 場景、城鎮座標、世界座標與時刻、戰鬥階段、結局段號。
func (g *game) snapshot() map[string]any {
	row := map[string]any{}
	switch {
	case g.endingReached():
		row["scene"] = string(sceneEnding)
	case g.scene == sceneFlow && g.flow != nil:
		row["scene"] = "flow:" + string(g.flow.Scene)
	default:
		row["scene"] = string(g.scene)
	}
	if g.play != nil {
		x, y := g.play.Outdoor.PartyPosition()
		town := map[string]any{"mode": int(g.play.Mode), "x": x, "y": y,
			"town": g.play.Outdoor.Town, "left_town": g.play.LeftTown}
		if g.play.Outdoor.AtDoorCell {
			town["at_door"] = true
		}
		if g.play.Outdoor.AtEdge {
			town["at_edge"] = true
		}
		if len(g.play.TurnedIn) > 0 {
			town["turned_in"] = len(g.play.TurnedIn)
		}
		if q := g.play.TookQuest; q != nil {
			town["took_quest"] = map[string]any{"member": q.Member,
				"site": q.Site, "final": q.Final}
		}
		// 隊伍與戰場的摘要壓成一行一名的字串，而且只在與上一步不同時才寫。
		// 上千步的路徑逐步展開成物件會把軌跡撐到兩百萬 byte，其中九成是重複的。
		if party := partyLines(g.play.Party); !sameLines(party, g.lastParty) {
			town["party"] = party
			g.lastParty = party
		}
		// 法術編輯器停在哪一欄、造價算到多少（RE-128）。長路徑不會經過這一頁，
		// 所以只在開著的時候多這一塊。
		if in := g.play.Inside; in != nil && in.SpellEdit != nil {
			e := in.SpellEdit
			edit := map[string]any{"cursor": e.Cursor,
				"adventure": e.Adventure, "gold": e.Gold,
				"name": e.Fields[spellNameField].Text,
				"stat": e.Fields[spellStatField].Text}
			if e.Menu != nil {
				edit["menu"] = e.Menu.Total
			}
			if e.Typing != nil {
				edit["typing"] = *e.Typing
			}
			town["spell_edit"] = edit
		}
		if d := g.play.Door(); d != nil {
			door := map[string]any{"index": d.Index, "name": d.Name}
			if q := d.Quest(); q != nil {
				door["quest_site"] = q.Site
				door["quest_final"] = q.Final
				door["quest_wants"] = d.QuestWants
			}
			town["door"] = door
		}
		row["town"] = town
	}
	if g.worldState != nil && g.play != nil && g.play.LeftTown {
		t := g.worldState.Travel
		w := map[string]any{"x": t.X, "y": t.Y, "sub_x": t.SubX, "sub_y": t.SubY,
			"time_index": g.worldState.TimeIndex}
		if c := g.worldState.Clock; c != nil {
			w["day"] = c.Day + 1
			w["month"] = c.Month + 1
			w["ticks"] = c.Ticks
		}
		if g.worldSession != nil {
			w["mode"] = int(g.worldSession.Mode)
			if g.worldSession.Camping {
				w["camping"] = true
			}
			if p := g.worldSession.Prompt; p != nil {
				switch p.Kind {
				case world.PromptQuest:
					w["quest_prompt"] = true
				case world.PromptRideAway:
					w["ride_prompt"] = true
				case world.PromptNotice:
					w["notice"] = true
				case world.PromptHealing:
					w["heal_prompt"] = true
				}
			}
		}
		row["world"] = w
	}
	if g.quest != nil && g.quest.state != nil {
		st := g.quest.state
		b := map[string]any{"stage": int(st.Stage), "round": st.Round.Number,
			"prompt": int(st.Plan.Prompt), "verdict": int(st.Verdict),
			"turn": st.Round.Turn, "quest_done": st.QuestDone}
		if len(st.Exec.Text) > 0 {
			b["exec_text"] = st.Exec.Text
		}
		if st.Stage == battle.StageSpoils {
			b["spoil_phase"] = int(st.Spoil.Phase)
			b["spoil_lines"] = st.SpoilLines()
			b["spoil_cursor"] = st.Spoil.Cursor
			b["spoil_icon"] = st.Spoil.Icon
		}
		// 指令列、游標與參戰者只有產生腳本的時候要看；收據裡逐步記會把軌跡
		// 撐大好幾倍，`-trace-detail` 才寫。
		if g.traceDetail {
			b["order"] = len(st.Round.Order)
			b["cursor"] = st.Cursor
			b["bar"] = st.Plan.Bar
			b["selected"] = st.Plan.Selected
			b["actors"] = len(st.Actors)
			b["cursor_cell"] = []int{st.Plan.Cursor.Row, st.Plan.Cursor.Col,
				st.Plan.Cursor.Mode, st.Plan.Cursor.Icon}
		}
		// 戰場上的位置與傷勢每一步都在動，逐步記會把軌跡撐大好幾倍而且沒有人讀。
		// 只在**階段改變**時記一次：進戰場、分勝負、收尾各留一份。
		if g.traceDetail || int(st.Stage) != g.lastStage {
			g.lastStage = int(st.Stage)
			b["roster"] = actorLines(st.Actors)
		}
		if st.Site != nil {
			b["site"] = st.Site.Index
			b["monster"] = st.Site.MonsterName
		}
		row["battle"] = b
	}
	if g.ending != nil {
		row["ending"] = map[string]any{"panel": g.ending.Panel,
			"armed": g.endingArmed}
	}
	return row
}

// partyLines 把隊伍壓成一名一行：名字、金幣、傷勢、已回報數、任務旗標、
// 身上帶著幾件東西。
func partyLines(party []town.Member) []string {
	out := make([]string, 0, len(party))
	for _, m := range party {
		reported := 0
		if len(m.Progress) > 0 {
			reported = m.Progress[0]
		}
		carried := 0
		for _, item := range m.Carried {
			for _, b := range item {
				if b != 0 {
					carried++
					break
				}
			}
		}
		flags := make([]string, 0, len(m.Quests))
		for _, q := range m.Quests {
			flags = append(flags, strconv.Itoa(q))
		}
		out = append(out, fmt.Sprintf("%s gold=%d wounds=%d reported=%d quests=%s carried=%d",
			m.Name, m.Gold, m.Wounds, reported, strings.Join(flags, ","), carried))
	}
	return out
}

// actorLines 把戰場上的參戰者壓成一名一行：索引、名字、哪一邊、座標與頭胸的現值。
func actorLines(actors []battle.Actor) []string {
	var out []string
	for i := range actors {
		a := &actors[i]
		if a.Name == "" {
			continue
		}
		side := "party"
		switch a.Side {
		case battle.SideMonster:
			side = "foe"
		case battle.SideEmpty:
			// 競技場那一場沒上場的人：記錄 `+0x01` 是 2，站在觀眾席不參戰。
			side = "watch"
		}
		state := ""
		if a.Down() {
			state = " down"
		}
		out = append(out, fmt.Sprintf("%d %s %s %d,%d head=%d chest=%d%s",
			i, a.Name, side, a.X, a.Y, a.Head, a.Chest, state))
	}
	return out
}

// sameLines 比兩份摘要一不一樣。
func sameLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// writeTrace 把整條軌跡寫出去。沒有指定 -trace 就什麼都不做。
func (g *game) writeTrace() {
	if g.tracePath == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(g.tracePath), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	// 一步一行：整份縮排會把上千步的軌跡撐大三倍，而一行一步照樣讀得下去。
	var out bytes.Buffer
	out.WriteString("{\n \"schema\": \"kol-play-trace-v1\",\n \"steps\": [\n")
	for i, row := range g.traceRows {
		line, err := json.Marshal(row)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		out.WriteString("  ")
		out.Write(line)
		if i < len(g.traceRows)-1 {
			out.WriteString(",")
		}
		out.WriteString("\n")
	}
	out.WriteString(" ]\n}\n")
	if err := os.WriteFile(g.tracePath, out.Bytes(), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// renderScene 用指定的場景跑一次既有的繪製分派，供 flow 借用。
func (g *game) renderScene(sc scene) error {
	saved := g.scene
	g.scene = sc
	err := g.render()
	g.scene = saved
	return err
}

// drawPage 畫一張整頁美術。中文模式的畫布是原版的 `book.CJKScale` 倍，所以美術要
// 以最近鄰放大同樣的倍率，座標也跟著乘——否則只會畫在左上角那一塊原版大小的區域。
//
// 這幾張（封面、目錄頁、攤開的書頁、角色立繪）是**純美術**，沒有文字要換，所以
// 沒有各自的 CJK 版本；其餘畫面都有自己的 `…CJK` 繪製函式，那些自己處理倍率。
func (g *game) drawPage(img *ebiten.Image, x, y int) {
	if g.strings == nil {
		g.canvas.DrawAt(img, x, y)
		return
	}
	s := book.CJKScale
	g.canvas.DrawScaled(img, x*s, y*s, s)
}

func (g *game) render() error {
	if g.play != nil {
		g.worldParty = g.play.Party
	}
	g.canvas.Clear()
	switch g.scene {
	case sceneCover, sceneContents, sceneOpenSpread:
		index := map[scene]int{
			sceneCover: book.Cover, sceneContents: book.Contents,
			sceneOpenSpread: book.OpenSpread,
		}[g.scene]
		img, err := g.sprites.Get("BOOKPICS", index)
		if err != nil {
			return err
		}
		g.drawPage(img, 0, 0)
		if g.scene == sceneCover {
			drawCoverSubtitle(g.canvas.Image(), g.cjk, g.titleSubs)
		}
	case sceneNewChar:
		bg, err := g.sprites.Get("BOOKPICS", book.OpenSpread)
		if err != nil {
			return err
		}
		g.drawPage(bg, 0, 0)
		portrait, err := g.sprites.Get("CHARPICS", g.portrait)
		if err != nil {
			return err
		}
		g.drawPage(portrait, book.PortraitOrigin.X, book.PortraitOrigin.Y)
	case sceneScript:
		return ui.NewCharacter(g.canvas.Image(), g.sprites, g.small, g.large, g.state)
	case sceneScriptCJK:
		return ui.NewCharacterCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings, g.state)
	case sceneRoster:
		if g.strings != nil {
			return ui.RosterCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings,
				g.roster.Occupied())
		}
		return ui.Roster(g.canvas.Image(), g.sprites, g.small, g.roster.Occupied())
	case sceneDelete:
		if g.strings != nil {
			return ui.DeleteListCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings,
				g.remove.List(), g.remove.Cursor, g.remove.Confirm)
		}
		return ui.DeleteList(g.canvas.Image(), g.sprites, g.small,
			g.remove.List(), g.remove.Cursor, g.remove.Confirm)
	case sceneBuilding, sceneInn:
		if g.strings != nil {
			return ui.InnCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings, g.inn)
		}
		return ui.Inn(g.canvas.Image(), g.sprites, g.small, g.large, g.inn)
	case sceneFlow:
		return g.renderFlow()
	case scenePlay:
		// 最後一件任務回報掉之後畫的是結局（RE-106）。
		if g.endingReached() {
			return g.renderScene(sceneEnding)
		}
		// 競技場的對話蓋在野外上面（RE-108）。
		if g.play != nil && g.play.InArena() {
			if g.strings != nil {
				return ui.InnCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings, g.play.Inside)
			}
			return ui.Inn(g.canvas.Image(), g.sprites, g.small, g.large, g.play.Inside)
		}
		if g.quest != nil && g.quest.state != nil {
			return g.renderBattle()
		}
		if g.play.LeftTown && g.worldState != nil {
			if g.strings != nil {
				return ui.WorldCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings,
					g.worldParty, g.worldState, g.worldSession)
			}
			return ui.WorldSession(g.canvas.Image(), g.sprites, g.small, g.large,
				g.worldParty, g.worldState, g.worldSession)
		}
		if g.play.Mode == town.ModeInside {
			if g.strings != nil {
				return ui.InnCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings, g.play.Inside)
			}
			return ui.Inn(g.canvas.Image(), g.sprites, g.small, g.large, g.play.Inside)
		}
		if g.strings != nil {
			return ui.OutdoorCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings, g.play.Outdoor)
		}
		return ui.Outdoor(g.canvas.Image(), g.sprites, g.small, g.play.Outdoor)
	case sceneWorld:
		if g.strings != nil {
			return ui.WorldCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings,
				g.worldParty, g.worldState, g.worldSession)
		}
		return ui.WorldSession(g.canvas.Image(), g.sprites, g.small, g.large,
			g.worldParty, g.worldState, g.worldSession)
	case sceneOutdoor:
		if g.strings != nil {
			return ui.OutdoorCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings, g.outdoor)
		}
		return ui.Outdoor(g.canvas.Image(), g.sprites, g.small, g.outdoor)
	case sceneHarbor:
		if g.strings != nil {
			return ui.HarborMasterCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings)
		}
		return ui.HarborMaster(g.canvas.Image(), g.sprites, g.small)
	case sceneMenu:
		if g.strings != nil {
			return ui.CharacterOptionsCJK(g.canvas.Image(), g.sprites, g.cjk,
				g.strings, g.selected)
		}
		return ui.CharacterOptions(g.canvas.Image(), g.sprites, g.small, g.selected)
	case sceneBookOpen:
		if err := g.ensureBookPlayer(); err != nil {
			return err
		}
		g.canvas.Image().WritePixels(g.book.Image().Pix)
		return nil
	case sceneEnding:
		// 結局：一段一張圖加兩三筆敘事，按一次鍵推一段（RE-106）。
		var segs []town.Segment
		if p := g.ending.Current(); p != nil {
			// 一段的兩三筆要**接成一整段再拆座標**：第二筆開頭那幾個字接在第一筆
			// 的最後位置，分開解析會把它當成新的一段畫到左上角（RE-106）。
			// 英文模式不查譯文——小字型畫不出中文。
			tr := g.endTrans
			if g.cjk == nil {
				tr = nil
			}
			segs = town.JoinedSegments(g.endBank, tr, ending.Bank, p.Records)
		}
		// 英文模式沒有 CJK 字型，敘事用小字型畫。
		f := g.small
		if g.cjk != nil {
			f = g.cjk
		}
		return ui.Ending(g.canvas.Image(), g.sprites, f, g.ending, segs)
	case sceneTitle:
		bg, err := g.sprites.Get("BACKPICS", g.frame.Back)
		if err != nil {
			return err
		}
		g.drawPage(bg, 0, 0)
		if g.frame.Archive == "" || g.frame.Overlay < 0 {
			drawTitleSubtitle(g.canvas.Image(), g.cjk, g.titleSubs, g.frame.Back)
			return nil
		}
		rect, ok := g.sequence.Rect(g.frame.Archive)
		if !ok {
			return fmt.Errorf("沒有 %s 的疊圖矩形", g.frame.Archive)
		}
		overlay, err := g.sprites.Get(g.frame.Archive, g.frame.Overlay)
		if err != nil {
			return err
		}
		g.drawPage(overlay, rect.X, rect.Y)
	default:
		return fmt.Errorf("未知場景 %q", g.scene)
	}
	return nil
}

// ensureBookPlayer 在遊戲迴圈內建立翻書動畫的播放器。
//
// 動畫的來源是「畫好的目標頁」，因此要先把目標畫面畫進離屏畫布再讀回；這一步只能在
// 繪圖環境就緒之後做，所以延後到第一次繪製時。
func (g *game) ensureBookPlayer() error {
	if g.book != nil {
		return nil
	}
	start, err := g.renderPage(g.bookFrom)
	if err != nil {
		return err
	}
	source, err := g.renderPage(g.bookTarget)
	if err != nil {
		return err
	}
	player, err := book.NewOpenPlayer(g.bookAnim, start, source)
	if err != nil {
		return err
	}
	g.book = player
	return nil
}

// renderPage 把一頁畫進離屏畫布再讀回。翻書動畫的起點與來源都是「畫好的那一頁」，
// 而這一步只能在繪圖環境就緒之後做，所以延後到第一次繪製時。
func (g *game) renderPage(name string) (*image.RGBA, error) {
	aux := screen.NewCanvas()
	switch name {
	case "cover", "contents":
		index := book.Cover
		if name == "contents" {
			index = book.Contents
		}
		img, err := g.sprites.Get("BOOKPICS", index)
		if err != nil {
			return nil, err
		}
		aux.DrawAt(img, 0, 0)
	case "menu":
		// 動畫期間沒有任何項目反白——反白是在動畫結束之後才畫上去的。
		// 這一點由原版 accept-and-save-run1 的 frame-036、039–043 逐像素確認。
		if err := ui.CharacterOptions(aux.Image(), g.sprites, g.small, -1); err != nil {
			return nil, err
		}
	case "menu-selected":
		if err := ui.CharacterOptions(aux.Image(), g.sprites, g.small, g.selected); err != nil {
			return nil, err
		}
	case "roster":
		if err := ui.Roster(aux.Image(), g.sprites, g.small, g.roster.Occupied()); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("未知的書頁 %q", name)
	}
	return aux.Snapshot(), nil
}

func (g *game) Update() error {
	if len(g.jobs) > 0 {
		if !g.drawn {
			return nil
		}
		job := g.jobs[g.jobAt]
		if !job.skip {
			if err := g.canvas.WritePNG(job.path); err != nil {
				return err
			}
			fmt.Printf("%s %s\n", job.name, job.path)
		}
		if g.tracePath != "" {
			row := g.snapshot()
			row["step"] = g.jobAt
			row["name"] = job.name
			g.traceRows = append(g.traceRows, row)
		}
		g.jobAt++
		if g.jobAt >= len(g.jobs) {
			g.writeTrace()
			return ebiten.Termination
		}
		if apply := g.jobs[g.jobAt].apply; apply != nil {
			apply()
		}
		g.drawn = false
		return nil
	}
	if g.interactive {
		return g.updateInteractive()
	}
	// 截圖模式沒有離開流程，Escape 直接關掉。互動遊玩的 Escape 是「回上一頁」，
	// 所以那條路走上面那一支。
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if g.scene == sceneBookOpen && g.book != nil {
		// 每一步固定等 delay_ticks_per_step 個 BIOS tick；Ebitengine 固定 60 TPS。
		g.bookWait += title.TickHz / 60
		for g.bookWait >= float64(g.bookAnim.DelayTicks) && !g.book.Done() {
			g.bookWait -= float64(g.bookAnim.DelayTicks)
			g.book.Step()
		}
	}
	if g.scene == scenePlay && g.play != nil {
		g.updatePlay()
	}
	if g.scene == sceneEnding && g.ending != nil {
		// 每一段等一個鍵（原版 `sub_1B100(0)`，RE-106）。
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) ||
			inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.ending.Advance()
		}
	}
	if g.scene == sceneTitle && g.player != nil {
		// 原版的等待單位是 BIOS tick；Ebitengine 固定 60 TPS，換算後推進播放器。
		g.player.Advance(title.TickHz / 60)
		g.frame = g.player.Current()
	}
	return nil
}

// endingReached 回報要不要播結局。**觸發點是任務戰鬥打完那一刻**（RE-106）：
// 原版在 `sub_2B0FF` 的尾端先看 `word_34968`——那是「這一場真的把任務標成完成了」
// 的旗標（`sub_28F5D` 設的，RE-101）——成立才逐名檢查 `[0x1DD] == 0x17`。
// 交任務物品那一步不檢查，走在路上也不檢查。
func (g *game) endingReached() bool { return g.endingArmed }

// armEnding 是戰鬥收尾時的那一次檢查。標到任務完成才看，沒標到就不看。
func (g *game) armEnding(s *battle.State) {
	if g.play == nil || g.endBank == nil || s == nil || s.QuestDone == 0 {
		return
	}
	party := make([][]int, 0, len(g.play.Party))
	for _, m := range g.play.Party {
		party = append(party, m.Progress)
	}
	if ending.AnyTriggered(party, 0) {
		g.endingArmed = true
	}
}

// leftTown 回報這一次遊玩已經走出城門了——野外與戰鬥都算。flow 走到城鎮那一層
// 之後就靠它把畫面與按鍵交給 `scenePlay` 那一條，整條路徑才貫得通。
func (g *game) leftTown() bool {
	if g.quest != nil && g.quest.state != nil {
		return true
	}
	return g.play != nil && g.play.LeftTown
}

// renderFlow 依 flow 目前的畫面分派到既有的繪製函式。每一張都是已經對拍過的畫面，
// flow 只決定現在畫哪一張。
func (g *game) renderFlow() error {
	dst := g.canvas.Image()
	switch g.flow.Scene {
	case flow.SceneTitle:
		return g.renderScene(sceneTitle)
	case flow.SceneCover:
		return g.renderScene(sceneCover)
	case flow.SceneContents:
		return g.renderScene(sceneContents)
	case flow.SceneMenu:
		if g.strings != nil {
			return ui.CharacterOptionsCJK(dst, g.sprites, g.cjk, g.strings, g.flow.MenuCursor)
		}
		return ui.CharacterOptions(dst, g.sprites, g.small, g.flow.MenuCursor)
	case flow.SceneRoster:
		return g.renderScene(sceneRoster)
	case flow.SceneNewChar:
		if g.strings != nil {
			return ui.NewCharacterCJK(dst, g.sprites, g.cjk, g.strings, g.flow.Create)
		}
		return ui.NewCharacter(dst, g.sprites, g.small, g.large, g.flow.Create)
	case flow.SceneDelete:
		return g.renderScene(sceneDelete)
	case flow.SceneHarbor:
		return g.renderScene(sceneHarbor)
	case flow.SceneSelect:
		if g.strings != nil {
			return ui.DeleteListCJK(dst, g.sprites, g.cjk, g.strings,
				g.roster.Occupied(), g.flow.SelectCursor, false)
		}
		return ui.SelectCharacter(dst, g.sprites, g.small,
			g.roster.Occupied(), g.flow.SelectCursor)
	case flow.SceneInn:
		if g.strings != nil {
			return ui.InnCJK(dst, g.sprites, g.cjk, g.strings, g.flow.Inn)
		}
		return ui.Inn(dst, g.sprites, g.small, g.large, g.flow.Inn)
	case flow.SceneTown:
		// 最後一件任務回報掉之後畫的是結局，不是城鎮（RE-106）。
		if g.endingReached() {
			return g.renderScene(sceneEnding)
		}
		// 走出城門之後就不是城鎮那一層了：野外與戰鬥借 `scenePlay` 的繪製，
		// 那幾張都已經對拍過（RE-066…RE-079）。
		if g.leftTown() {
			return g.renderScene(scenePlay)
		}
		if g.flow.Town.Mode == town.ModeInside {
			if g.strings != nil {
				return ui.InnCJK(dst, g.sprites, g.cjk, g.strings, g.flow.Town.Inside)
			}
			return ui.Inn(dst, g.sprites, g.small, g.large, g.flow.Town.Inside)
		}
		if g.strings != nil {
			return ui.OutdoorCJK(dst, g.sprites, g.cjk, g.strings, g.flow.Town.Outdoor)
		}
		return ui.Outdoor(dst, g.sprites, g.small, g.flow.Town.Outdoor)
	}
	return fmt.Errorf("未知的 flow 畫面：%s", g.flow.Scene)
}

// walkTicksPerFrame 是隊伍標記的動畫節拍。原版用 `word_2DF28` 當延遲計數，
// 每三次繪製推進一格；Ebitengine 固定 60 TPS，取相近的節拍。
const walkTicksPerFrame = 8

// updateInteractive 是互動遊玩的那一幀：讀鍵盤與滑鼠，把 token 交給目前
// 掌管畫面的狀態機。
//
// **分派與 `-flow-script` 那條完全相同**——人在城裡就是 `flow.Key`，走出城門
// 之後交給 `playToken`（野外、戰鬥與結局）。收據跑的是同一段程式碼，所以互動
// 遊玩不是另一條沒被驗過的路徑。
func (g *game) updateInteractive() error {
	if g.flow == nil {
		return ebiten.Termination
	}
	// Ctrl+Q 是**額外**的退路，不是唯一的。原版的離開是在旅店住宿——住宿即存檔
	// 即離隊，隊伍空了問記錄 0x36「Do you wish to quit?」，答 YES 交棒回書本
	// （RE-073）；remake 那一條接的是回目錄頁，收據見
	// `internal/flow` 的 `TestInnDisbandReturnsToContents`。
	// 留著 Ctrl+Q 是因為人不一定停在旅店裡，而視窗總得關得掉。
	if (ebiten.IsKeyPressed(ebiten.KeyControlLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyControlRight)) &&
		inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}
	// 標題序列自己會播，按任何鍵才離開（RE-047）。
	if g.flow.Scene == flow.SceneTitle && g.player != nil {
		g.player.Advance(title.TickHz / 60)
		g.frame = g.player.Current()
	}
	// 隊伍標記的走路動畫。腳本模式逐步截圖不需要它自己動，互動模式要。
	if g.flow.Scene == flow.SceneTown && g.leftTown() && !g.endingReached() &&
		g.play != nil && !g.play.InArena() &&
		(g.quest == nil || g.quest.state == nil) && g.worldSession != nil {
		g.worldSession.Screen.AdvanceWeather(title.TickHz / 60)
	}
	if g.flow.Scene == flow.SceneTown && g.play != nil && !g.leftTown() &&
		g.play.Mode != town.ModeInside {
		g.walkTick++
		if g.walkTick >= walkTicksPerFrame {
			g.walkTick = 0
			g.play.Outdoor.AdvanceMarker(&g.markerStep)
		}
	}
	targets := g.flowTargets()
	typingName := g.flow.Scene == flow.SceneNewChar && g.flow.Create.Stage == charcreate.StageName
	tokens := append(keyboardTokens(typingName), mouseTokens(targets, "RET")...)
	if len(g.notice) > 0 {
		if len(tokens) > 0 {
			g.notice = nil
			if g.logInput {
				fmt.Fprintln(os.Stderr, "input scene=notice token=ack")
			}
		}
		return nil
	}
	if g.logInput && len(tokens) > 0 {
		mx, my := ebiten.CursorPosition()
		fmt.Fprintf(os.Stderr, "input scene=%s cursor=%d 游標=(%d,%d) 可點=%d token=%v\n",
			g.flow.Scene, g.flow.MenuCursor, mx, my, len(targets), tokens)
	}
	for _, tok := range tokens {
		if g.leftTown() || g.endingReached() {
			// 這一層認不得的鍵在原版也不做事；玩家亂按不該中斷遊戲。
			_ = g.playToken(tok)
			continue
		}
		full := g.flow.Scene == flow.SceneMenu && g.flow.MenuCursor == flow.MenuCreate && g.roster.Full() && tok == "RET"
		g.flow.Key(tok)
		if full {
			g.notice = g.saveNotice("full")
		}
	}
	return nil
}

// updatePlay 是城鎮這一段的輸入迴圈。鍵位照參考卡：方向鍵移動、
// `<`／`>` 移動圖示列白框、Return 執行。
func (g *game) updatePlay() {
	// 戰鬥蓋在最上面：規劃段的指令列吃掉全部按鍵。
	if g.quest != nil && g.quest.state != nil {
		g.updateBattle()
		return
	}
	g.walkTick++
	if g.walkTick >= walkTicksPerFrame {
		g.walkTick = 0
		g.play.Outdoor.AdvanceMarker(&g.markerStep)
	}
	// MOUTH 打字中：鍵盤是輸入框，方向鍵與圖示列都先讓開（RE-067）。
	if g.play.Asking() {
		for _, r := range ebiten.AppendInputChars(nil) {
			g.play.TypeAsk(r)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			g.play.BackspaceAsk()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
			g.play.SubmitAsk()
		}
		return
	}
	for _, k := range []struct {
		key    ebiten.Key
		dx, dy int
	}{
		{ebiten.KeyArrowLeft, -1, 0}, {ebiten.KeyArrowRight, 1, 0},
		{ebiten.KeyArrowUp, 0, -1}, {ebiten.KeyArrowDown, 0, 1},
		// 參考卡的斜走鍵。
		{ebiten.KeyA, -1, -1}, {ebiten.KeyS, 1, -1},
		{ebiten.KeyZ, -1, 1}, {ebiten.KeyX, 1, 1},
	} {
		if inpututil.IsKeyJustPressed(k.key) {
			g.play.Walk(k.dx, k.dy)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyComma) {
		g.play.MoveCursor(-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPeriod) {
		g.play.MoveCursor(1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.play.Accept()
	}
}

// ambushLines 是遭遇那一句：`YOU HAVE BEEN AMBUSHED BY n, LEVEL m NAMES!`
// （`0x2A886`，RE-097）。名字的複數與打贏那一句同一套。
func ambushLines(bank *town.UIText, trans *town.Translations, a *battle.Ambush,
	name string, english bool) []string {
	countRec, levelRec := battle.AmbushLines()
	// 記錄 0xE7 自己帶一個 `@`（換行），所以取 Segment 而不是剝掉版面碼的 UIRaw：
	// 原版印出來是「YOU HAVE BEEN AMBUSHED」「BY n, LEVEL m NAMES!」兩行。
	segs := town.UISegments(bank, trans, town.BankUI, countRec)
	out := make([]string, 0, len(segs))
	for i, seg := range segs {
		if i < len(segs)-1 {
			out = append(out, seg.Text)
			continue
		}
		last := seg.Text + strconv.Itoa(a.Count) +
			town.UIRaw(bank, trans, town.BankUI, levelRec) +
			strconv.Itoa(a.Level) + " " + name
		if english {
			last += battle.PluralSuffix(name, a.Count)
		}
		out = append(out, last+"!")
	}
	return nonEmptyLines(out)
}

// battleSpellRows 是法術那兩張清單要顯示的幾列。正負那一張的文字是 KOLDATA 記錄，
// 規則層只留位置不碰文字（RE-100）。
func (g *game) battleSpellRows() []string {
	s := g.quest.state
	if s.Plan.Prompt != battle.PromptAffect {
		return s.Plan.List
	}
	out := make([]string, 0, len(battle.AffectRecords()))
	for _, rec := range battle.AffectRecords() {
		out = append(out, g.battleRecord(rec))
	}
	return out
}

// battleWoreOffLines 是這一回合法術剛好失效的那幾句（記錄 0x173，RE-100）：
// 一位一句，各等一次鍵。
func (g *game) battleWoreOffLines() [][]string {
	s := g.quest.state
	if len(s.WoreOff) == 0 {
		return nil
	}
	out := make([][]string, 0, len(s.WoreOff))
	for _, i := range s.WoreOff {
		ctx := battle.MessageContext{Actor: g.battleSpeaker(i), Lookup: g.battleRecord}
		out = append(out,
			battle.WrapMessage(ctx.Line(g.battleRecord(battle.WearsOffRecord))))
	}
	return out
}

// battleTakeSpoil 在戰利品清單按下確認：白框停在 U-TURN 就取消，否則領走一件。
// 背包塞不下的時候原版會說一句（記錄 0x183），把它排進訊息區。
func (g *game) battleTakeSpoil() {
	s := g.quest.state
	_, _, done := s.SpoilAccept()
	if s.Spoil.Say != 0 {
		ctx := battle.MessageContext{Lookup: g.battleRecord}
		s.Exec.Text = battle.WrapMessage(
			ctx.Substitute(g.battleRecords([]int{s.Spoil.Say})))
	}
	if done {
		s.Stage = battle.StageDone
	}
}

// battleSpeaker 把一名參戰者換成訊息代換要用的資料（RE-086）。
func (g *game) battleSpeaker(i int) battle.Speaker {
	s := g.quest.state
	if i < 0 || i >= len(s.Actors) {
		return battle.Speaker{}
	}
	a := s.Actors[i]
	out := battle.Speaker{Name: a.Name, Monster: a.Member < 0}
	if a.Member >= 0 && a.Member < len(s.Party) {
		m := s.Party[a.Member]
		out.Female = m.Female
		out.Weapon = m.WeaponName
		out.Belt = m.BeltName
		out.Weapon = g.strings.DisplayName(out.Weapon)
		out.Belt = g.strings.DisplayName(out.Belt)
		return out
	}
	// 敵人手上那件武器的名字在怪物記錄 +0x20（RE-080）。少了它 `|E` 會展開成空，
	// 訊息變成 `THE BINDERAK CUTS HIS  AT POONTANG`——實跑抓到的。
	if s.Monster.Weapon != nil {
		out.Weapon = s.Monster.Weapon.Name
	}
	out.Name = g.strings.DisplayName(out.Name)
	out.Weapon = g.strings.DisplayName(out.Weapon)
	return out
}

// battleRecord 取一筆 KOLDATA 訊息。
func (g *game) battleRecord(record int) string {
	if g.quest == nil || g.quest.bank == nil {
		return ""
	}
	// 用 UIRaw 而不是 Lines：`]` 要保留成空白（`The]` ＋ 名字），Lines 會把兩端的
	// 空白 trim 掉。`|X` 兩者都原樣留著，由 Substitute 展開。中文模式優先用譯文。
	return town.UIRaw(g.quest.bank, g.quest.trans, town.BankUI, record)
}

// battleAreaLine 是挑完部位之後那一句效果（`0x287E4`，RE-100）。問句本身已經印過，
// 這裡直接換成效果那一句。
func (g *game) battleAreaLine(s *battle.State, res battle.CastResult, caster int) {
	ctx := battle.MessageContext{Lookup: g.battleRecord}
	negative := caster >= 0 && caster < len(s.Actors) && s.Actors[caster].Negative
	before, after := battle.SpellEffectLine(res.Power, negative)
	s.Exec.Text = battle.WrapMessage(ctx.Substitute(
		g.battleRecords(before) + " " + g.battleRecords(after)))
}

// battleStep 做出執行段的下一手，並把那一句訊息展開、斷行放進畫面
// （RE-086／RE-087；攻擊那一串的索引規則見 RE-088）。
func (g *game) battleStep() {
	if g.quest == nil || g.quest.state == nil {
		return
	}
	s := g.quest.state
	// 同一手還沒印完的下一句：攻擊命中分兩句，擊退成立還會再多一到兩句，
	// 各等一次鍵（RE-088／RE-095）。
	if len(s.Exec.Pending) > 0 {
		s.Exec.Text = s.Exec.Pending[0]
		s.Exec.Pending = s.Exec.Pending[1:]
		return
	}
	// 這一場已經結束、勝負的兩句也印完了，就停在最後一句上，不要再往下走——
	// 再呼叫一次 Step 會把勝負訊息重新排一遍，變成按不完的迴圈。
	if s.Stage == battle.StageDone {
		g.awardBattle()
		// 逃走那一句（RE-103）：`Flee()` 把訊息清空，所以第一次進來先印
		// 「You flee for your life!」，按一次鍵才真的離開戰場。
		if s.Verdict == battle.VerdictFled && len(s.Exec.Text) == 0 {
			s.Exec.Text = g.battleVerdictLines()
			return
		}
		// 打贏的話還有一張戰利品清單（RE-096）。地上什麼都沒有就跳過。
		if s.Verdict == battle.VerdictWon && !g.quest.spoilsOpened {
			g.quest.spoilsOpened = true
			if s.BeginSpoils() {
				s.Stage = battle.StageSpoils
				return
			}
		}
		// 競技場那一場的收尾（`0x21607`，RE-103／RE-108）：打贏升一階（一次一級），
		// 接著結算賭金——輸掉那一場時整排先取負。
		if g.arenaFighter >= 0 {
			won := s.Verdict == battle.VerdictWon
			var lines []string
			if won {
				lines = g.play.ArenaWin(g.arenaFighter)
			}
			lines = append(lines, g.play.SettleBets(g.arenaBets, won)...)
			g.arenaFighter, g.arenaBets = -1, nil
			if len(lines) > 0 {
				s.Exec.Text = lines
				return
			}
		}
		// 這一場標到了任務完成，而且有人的已回報數正好是 0x17，就接結局
		// （RE-106）。檢查在放開畫面之前，順序與原版相同。
		g.armEnding(s)
		g.writeBattleParty(s)
		// 勝負那兩句與戰利品都處理完了，這一場真的結束——放開畫面回野外，
		// 玩家才能繼續走。傷勢與既有裝備修改已依 SPEC-027 回寫；
		// 敗北的額外掉落及完整原版返回路徑仍待驗證。
		g.quest.state = nil
		return
	}
	// 沒有訊息的那幾手不停下來等鍵，停在空白畫面只會讓人以為當掉了。原版每一手
	// 都印一句（SPEC-008「已知差異」），哪些手會落到這裡還沒盤點。
	var r battle.ActResult
	for {
		var ok bool
		r, ok = s.Step(s.Rand)
		if !ok {
			// 新回合開始前，身上的法術先倒數一次；剛好過期的各印一句
			// （記錄 0x173，RE-100）。
			if lines := g.battleWoreOffLines(); len(lines) > 0 {
				s.WoreOff = nil
				s.Exec.Text, s.Exec.Pending = lines[0], lines[1:]
				return
			}
			// 還沒分勝負但站到了戰場邊上：問「要逃嗎」（記錄 0x109，RE-103）。
			if s.Plan.Prompt == battle.PromptFlee {
				s.Exec.Text = battle.WrapMessage(
					g.battleRecord(battle.FleeAskRecord))
				return
			}
			s.Exec.Text = g.battleVerdictLines()
			return
		}
		if r.Message != 0 || r.Outcome != nil || len(r.Records) > 0 ||
			r.Exhausted != 0 {
			break
		}
	}
	ctx := battle.MessageContext{
		Actor:  g.battleSpeaker(r.Actor),
		Target: g.battleSpeaker(r.Target),
		Lookup: g.battleRecord,
	}
	// 力竭那一句接在這一手的訊息後面（`0x20D9A`，RE-131）：原版是這一手做完、
	// 疲勞算完才另起一段。這裡用 defer 是因為下面每一種訊息各自 return。
	defer func() {
		if r.Exhausted == 0 {
			return
		}
		line := battle.WrapMessage(ctx.Line(g.battleRecord(r.Exhausted)))
		if s.Exec.Text == nil {
			s.Exec.Text = line
			return
		}
		s.Exec.Pending = append(s.Exec.Pending, line)
	}()
	// 丟與撿的訊息在動作**做完之後**才展開，東西已經不在原處了，所以名字由
	// ActResult 帶出來（RE-094）：`|E` 是丟掉的那件、`|H` 是撿到的那件。
	if r.Item != "" {
		ctx.Item = g.strings.DisplayName(r.Item)
		if r.Command == battle.CmdDrop {
			ctx.Actor.Weapon = g.strings.DisplayName(r.Item)
		}
	}
	// 施法那一手：`|S` 是法術的名字，生效之後再接一句「The spell has a … effect.」
	// （RE-100）。
	if r.Cast != nil {
		if m := s.Member(r.Actor); m != nil && r.Actor >= 0 {
			if a := s.Actors[r.Actor]; a.SpellPick >= 0 && a.SpellPick < len(m.Spells) {
				ctx.Scratch = m.Spells[a.SpellPick].Name
			}
		}
		s.Exec.Text = battle.WrapMessage(ctx.Line(g.battleRecord(r.Message)))
		// 效果 6 而且不用選目標格：接的不是效果那一句，是「挑一個部位」那一問
		// （記錄 0x184 ＋ hurt／heal ＋ 0x16C，RE-100）。答完才輪到效果那一句。
		if r.Cast.NeedsArea {
			s.Exec.Pending = [][]string{battle.WrapMessage(ctx.Substitute(
				g.battleRecords(battle.AreaPrompt(s.Actors[r.Actor].Negative))))}
			return
		}
		before, after := battle.SpellEffectLine(r.Cast.Power,
			s.Actors[r.Actor].Negative)
		s.Exec.Pending = [][]string{battle.WrapMessage(ctx.Substitute(
			g.battleRecords(before) + " " + g.battleRecords(after)))}
		return
	}
	if len(r.Records) > 0 {
		// 一句要接兩筆記錄的（射擊撞在樹上或牆上，RE-102）。
		s.Exec.Text = battle.WrapMessage(ctx.Line(g.battleRecords(r.Records)))
		return
	}
	// 撿那三句前面不加名字：問句自己就是一句完整的話，0x112／0x113 開頭就是
	// `|G`（RE-113）。
	if r.Plain {
		s.Exec.Text = battle.WrapMessage(ctx.Substitute(g.battleRecord(r.Message)))
		return
	}
	if r.Outcome == nil {
		s.Exec.Text = battle.WrapMessage(ctx.Line(g.battleRecord(r.Message)))
		// 移動成功那一句原版要按兩次：印完等一次、重畫戰場再等一次（RE-085）。
		if r.Redraw {
			s.Exec.Pending = [][]string{s.Exec.Text}
		}
		return
	}
	// 攻擊那一手：第一句是動作加結果，命中的話第二句是傷害（RE-088）。
	a := s.Actors[r.Actor]
	heavy := a.Weapon.Kind > 0x7F
	s.Exec.Text = battle.WrapMessage(ctx.Line(
		g.battleRecords(battle.AttackLine(r.Command, a.Style, *r.Outcome, heavy))))
	if r.Outcome.Miss || r.Outcome.Parried {
		return
	}
	monster := r.Target >= 0 && s.Actors[r.Target].Member < 0
	s.Exec.Pending = [][]string{battle.WrapMessage(ctx.Substitute(
		g.battleRecords(battle.DamageLine(*r.Outcome, 0, monster))))}
	// 擊退或打暈那一句接在傷害之後（RE-095）。`|F` 指的還是被打的那一位。
	if len(r.Knock) > 0 {
		s.Exec.Pending = append(s.Exec.Pending, battle.WrapMessage(
			ctx.Substitute(g.battleRecords(r.Knock))))
	}
}

// townCarried 把角色檔裡身上那十七格的十六進位換成位元組，順序照抽取器給的
// 位移由小到大——`town.CarriedOffsets` 是同一個順序（RE-109）。
func townCarried(slots []roster.CarriedSlot) [][]byte {
	if len(slots) == 0 {
		return nil
	}
	out := make([][]byte, 0, len(slots))
	for _, s := range slots {
		raw, err := hex.DecodeString(s.Bytes)
		if err != nil || len(raw) != town.CarriedStride {
			raw = make([]byte, town.CarriedStride)
		}
		out = append(out, raw)
	}
	return out
}

// battleVerdictLines 是這一場結束時要印的那幾句（RE-091）。沒分勝負就回 nil，
// 執行段照常排下一回合。
//
// 打贏分兩句、各等一次鍵：第一句報戰果，第二句報金冠。**冒險點數沒有報**——
// 那一句的數字要 `sub_26C5F(0x17)` 才算得出來，那一支還沒解讀，寧可不報。
func (g *game) battleVerdictLines() []string {
	s := g.quest.state
	enemies := s.EnemyCount()
	name := s.Monster.Name
	if !g.quest.english {
		name = g.strings.DisplayName(name)
	}
	won, lost := "", ""
	if g.quest.english {
		won = battle.PluralSuffix(name, enemies)
		lost = battle.DefeatSuffix(name, enemies)
	}
	ctx := battle.MessageContext{Lookup: g.battleRecord}
	switch s.Verdict {
	case battle.VerdictWon:
		// `|F` 展開要有目標；原版是把 word_3237E 指到第一個敵人（`0x211D5`）。
		ctx.Target = g.battleSpeaker(battle.PartySlots)
		first := ctx.Substitute(g.battleRecords(battle.VictoryRecords())) + won + "!"
		// 競技場與訓練戰整段獎勵都跳過（`0x2124E`，RE-096），所以那一句不印，
		// 接著的是升階與賭金結算。
		if s.RankFight {
			return battle.WrapMessage(first)
		}
		// 兩個 `|N` 各讀一次，所以分兩段展開：前半是冒險點數、後半是金冠。
		before, after := battle.RewardRecords(s.PartyCount())
		ctx.Number = s.Adventure
		reward := ""
		// 隊伍只有一個人的時候原版不印 `Each player`，改逐一印在籍角色的名字
		// （`0x212F4`）。名字不是記錄，所以插在前面。
		if s.PartyCount() == 1 {
			for i := range s.Party {
				reward += s.Party[i].Name + " "
			}
		}
		reward += ctx.Substitute(g.battleRecords(before))
		ctx.Number = s.Gold
		reward += ctx.Substitute(g.battleRecords(after))
		s.Exec.Pending = [][]string{battle.WrapMessage(reward)}
		return battle.WrapMessage(first)
	case battle.VerdictLost:
		before, after := battle.DefeatRecords(enemies)
		line := g.battleRecords(before) + name + lost + g.battleRecords(after)
		return battle.WrapMessage(ctx.Substitute(line))
	case battle.VerdictFled:
		// 「You flee for your life!」——單複數看隊伍人數（RE-103）。
		return battle.WrapMessage(ctx.Substitute(
			g.battleRecords(battle.FleeRecords(s.PartyCount()))))
	}
	return nil
}

// battleRecords 把一串記錄號接成一段文字。原版是逐筆送進同一個緩衝區，
// 接起來再展開代換碼等價（RE-088）。
func (g *game) battleRecords(records []int) string {
	out := ""
	for _, r := range records {
		out += g.battleRecord(r)
	}
	return out
}

// updateBattle 是戰場的輸入迴圈（SPEC-010）。鍵位與城鎮一致：方向鍵與四個
// 斜走鍵移游標、`<`／`>` 移圖示列白框、Return 確定。
func (g *game) updateBattle() {
	s := g.quest.state
	if s.Stage == battle.StageEncounter {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) ||
			inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			s.Continue(s.Rand)
		}
		return
	}
	if s.Stage == battle.StageSpoils {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
			s.SpoilMove(-1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
			s.SpoilMove(1)
		}
		// `,`／`.`（不按 Shift 的 `<`／`>`）移圖示列的白框，與城鎮同一組鍵。
		if inpututil.IsKeyJustPressed(ebiten.KeyComma) {
			s.SpoilIconMove(-1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyPeriod) {
			s.SpoilIconMove(1)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) ||
			inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.battleTakeSpoil()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyN) ||
			inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			if s.SpoilCancel() {
				s.Stage = battle.StageDone
			}
		}
		return
	}
	for _, k := range []struct {
		key      ebiten.Key
		row, col int
	}{
		{ebiten.KeyArrowUp, -1, 0}, {ebiten.KeyArrowDown, 1, 0},
		{ebiten.KeyArrowLeft, 0, -1}, {ebiten.KeyArrowRight, 0, 1},
		{ebiten.KeyA, -1, -1}, {ebiten.KeyS, -1, 1},
		{ebiten.KeyZ, 1, -1}, {ebiten.KeyX, 1, 1},
	} {
		if inpututil.IsKeyJustPressed(k.key) {
			s.MoveCursor(k.row, k.col, false)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyComma) {
		s.MoveSelection(-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPeriod) {
		s.MoveSelection(1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) &&
		(s.Plan.Prompt == battle.PromptCommand || s.Plan.Prompt == battle.PromptTarget) {
		s.ToggleEquipmentPanel()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		switch s.Plan.Prompt {
		case battle.PromptNextRound:
			s.Answer(true)
			g.battleStep()
		case battle.PromptExecuting:
			g.battleStep()
		case battle.PromptFlee, battle.PromptPickUp, battle.PromptArea:
			// 逃走與撿只認 y／n，挑部位只認 1…6，這兩個鍵在這三問沒有作用。
		default:
			s.Choose()
		}
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		switch s.Plan.Prompt {
		case battle.PromptNextRound, battle.PromptFlee:
			s.Answer(false)
		case battle.PromptPickUp:
			s.AnswerPickUp(false)
			g.battleStep()
		}
	}
	// 挑部位那一問只認 1…6（`0x284D8` 的 `sub_1B100(6)`）。
	if s.Plan.Prompt == battle.PromptArea {
		for k := 0; k < 6; k++ {
			if !inpututil.IsKeyJustPressed(ebiten.Key1 + ebiten.Key(k)) {
				continue
			}
			if res, caster, ok := s.AnswerArea(k); ok {
				g.battleAreaLine(s, res, caster)
			}
			break
		}
	}
	// 逃走與撿那兩問只認 y／n（`0x21158`／`0x2036E`），Return 與空白鍵都不算數。
	if inpututil.IsKeyJustPressed(ebiten.KeyY) {
		switch s.Plan.Prompt {
		case battle.PromptFlee:
			s.Answer(true)
			g.battleStep()
		case battle.PromptPickUp:
			s.AnswerPickUp(true)
			g.battleStep()
		}
	}
}

func (g *game) Draw(dst *ebiten.Image) {
	if err := g.render(); err != nil {
		fmt.Fprintln(os.Stderr, "繪製失敗：", err)
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(g.canvas.Image(), op)
	if len(g.notice) > 0 {
		g.drawNotice(dst)
	}
	g.drawn = true
}

func (g *game) Layout(int, int) (int, int) {
	b := g.canvas.Image().Bounds()
	return b.Dx(), b.Dy()
}

// applyToken 把腳本 token 套到狀態機；token 對應原版的鍵盤操作。
func applyToken(st *charcreate.State, token string) error {
	switch token {
	case "RET":
		st.Accept()
	case "ESC":
		st.Cancel()
	case "UP":
		st.MoveCursor(-1)
		st.MovePortrait(-1)
	case "DOWN":
		st.MoveCursor(1)
		st.MovePortrait(1)
	case "LEFT":
		st.MovePortrait(-1)
	case "RIGHT":
		st.MovePortrait(1)
	case "BS":
		st.Backspace()
	case "SPACE":
		st.TypeRune(' ')
	default:
		r := []rune(token)
		if len(r) != 1 {
			return fmt.Errorf("未知腳本 token %q", token)
		}
		st.TypeRune(r[0])
	}
	return nil
}

// applyDeleteToken 把腳本 token 套到刪除流程；token 對應原版的鍵盤操作。
func applyDeleteToken(f *roster.DeleteFlow, token string) error {
	switch token {
	case "RET":
		f.Accept()
	case "UP":
		f.MoveCursor(-1)
	case "DOWN":
		f.MoveCursor(1)
	case "y", "Y":
		f.Answer(true)
	case "n", "N":
		f.Answer(false)
	default:
		return fmt.Errorf("未知刪除腳本 token %q", token)
	}
	return nil
}

// batchFlags 是「產出檔案然後結束」的那些旗標。設了其中任何一個就不是互動遊玩。
var batchFlags = map[string]bool{
	"shot": true, "shot-dir": true, "shot-steps": true,
	"script": true, "flow-script": true, "play-script": true,
	"world-script": true, "delete-script": true, "chardata-script": true,
	"title-shot-dir": true, "book-shot-dir": true, "title-timeline": true,
	"title-frame": true, "list-assets": true, "dump-cjk-layout": true,
	"chardata-write": true, "chardata-delete-slot": true, "chardata-out": true,
	"route-to": true, "trace": true,
}

// useResourceDir 讓程式在自己旁邊找資源。
//
// 全部的資源旗標都是相對路徑（`assets/…`、`docs/…`），從儲存庫根目錄跑沒問題，
// 但發行包不是這樣被打開的：AppImage 掛在暫時目錄、macOS 的 .app 由 Finder
// 啟動時工作目錄是 `/`、Windows 的捷徑也可能指到別處。工作目錄找得到資源就
// 不動，找不到才換到執行檔旁邊，所以既有的對拍與截圖流程完全不受影響。
func useResourceDir(marker string) {
	if _, err := os.Stat(marker); err == nil {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return
	}
	dir := filepath.Dir(exe)
	// macOS 的 .app 把執行檔放在 Contents/MacOS，資源放在 Contents/Resources。
	for _, cand := range []string{dir, filepath.Join(dir, "..", "Resources")} {
		if _, err := os.Stat(filepath.Join(cand, marker)); err == nil {
			_ = os.Chdir(cand)
			return
		}
	}
}

// defaultSavePath 是互動遊玩的存檔位置：使用者的設定目錄底下。
// Linux 是 ~/.config、macOS 是 ~/Library/Application Support、Windows 是 %AppData%，
// 三個平台都由 os.UserConfigDir 決定，不自己拼路徑。
func defaultSavePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "kol-remake-cht")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "save.dat"), nil
}

// buildVersion 由正式封包的連結參數注入；開發建置不冒充正式版本。
var buildVersion = "development"

func main() {
	showVersion := flag.Bool("version", false, "顯示程式版本後結束")
	assetRoot := flag.String("assets", "assets/original-derived-png/ega", "抽取資產根目錄")
	fontRoot := flag.String("fonts", "assets/original-derived-png/fonts", "字型資產目錄")
	dataPath := flag.String("data", "assets/original-derived-data/character-creation.json",
		"角色建立 typed data")
	sceneName := flag.String("scene", string(sceneGame),
		"要繪製的場景；預設 game 是互動遊玩，其餘場景名是對拍與截圖用的單張畫面")
	portrait := flag.Int("portrait", 0, "newchar 場景使用的角色圖像圖號")
	menuSelected := flag.Int("menu-selected", 0, "menu 場景反白的項目索引")
	innPath := flag.String("inn", "docs/re/town-inn-sample.json",
		"旅店畫面的對拍樣本")
	innSelected := flag.Int("inn-selected", -1, "覆寫旅店圖示列的白框位置（-1 為沿用樣本）")
	outdoorPath := flag.String("outdoor", "docs/re/town-outdoor-sample.json",
		"戶外畫面樣本狀態 JSON")
	outdoorCamera := flag.String("outdoor-camera", "", "覆寫戶外鏡頭座標，格式 X,Y")
	outdoorFrame := flag.Int("outdoor-frame", -1, "覆寫隊伍標記的動畫畫格（-1 為沿用樣本）")
	outdoorMarker := flag.String("outdoor-marker", "", "覆寫隊伍標記在視窗內的偏移，格式 列,欄")
	uiI18nPath := flag.String("ui-i18n", "docs/i18n/ui-zh-Hant.json",
		"UI 文字庫（KOLDATA／BOOKDATA／ENDTXT）的繁體中文譯文")
	displayNamesPath := flag.String("display-names", "docs/i18n/display-names-zh-Hant.tsv",
		"資料名稱的繁體中文顯示譯名")
	titleSubsPath := flag.String("title-i18n", "docs/i18n/title-zh-Hant.json",
		"開場字幕卡與書本封面的中文字幕")
	spellWordsPath := flag.String("spell-words", "docs/i18n/spell-words-zh-Hant.json",
		"法術名精靈語音節的中文注解表（SPEC-034）")
	townI18nPath := flag.String("town-i18n", "docs/i18n/town-zh-Hant.json",
		"城鎮建築的譯文表")
	townTextPath := flag.String("town-text",
		"assets/original-derived-data/town-text.json", "城鎮文字 typed data")
	buildingTown := flag.Int("building-town", 7, "建築內部畫面的城鎮編號")
	buildingDoor := flag.Int("building-door", 5, "建築內部畫面的門序號")
	worldPath := flag.String("world-map",
		"assets/original-derived-data/world-map.json", "野外地圖 typed data")
	battleTiles := flag.String("battle-tiles",
		"assets/original-derived-png/ega/BATTLE/battle-tiles-08.png", "戰場磚圖")
	battleMaps := flag.String("battle-maps",
		"assets/original-derived-png/ega/BATTLE/battle-map-%02d.png", "戰場格子陣列")
	questPath := flag.String("quest-sites",
		"assets/original-derived-data/quest-sites.json", "任務地點 typed data")
	monsterPath := flag.String("monsters",
		"assets/original-derived-data/monsters.json", "怪物 typed data")
	battleItemsPath := flag.String("battle-items",
		"assets/original-derived-data/items.json", "戰後掉落物品表 typed data")
	worldPos := flag.String("world-pos", "", "覆寫野外世界座標，格式 X,Y")
	worldSub := flag.String("world-sub", "", "覆寫野外次格座標，格式 次格X,次格Y")
	worldWeather := flag.Int("world-weather", 0, "天候組 0…5")
	worldPhase := flag.Int("world-phase", 0, "雲層旋轉過的像素數")
	worldTime := flag.Int("world-time", 0, "時段 0…4")
	worldScript := flag.String("world-script", "", "野外的操作腳本；每一步後截一張圖")
	topicsPath := flag.String("town-topics",
		"assets/original-derived-data/town-keywords.json", "城鎮話題 typed data")
	uiTextPath := flag.String("ui-text",
		"assets/original-derived-data/ui-text.json", "UI 文字庫 typed data")
	buildingsPath := flag.String("town-buildings",
		"assets/original-derived-data/town-buildings.json", "城鎮建築 typed data")
	portraitPath := flag.String("portrait-layout",
		"assets/original-derived-data/portrait-layout.json", "立繪版面 typed data")
	magicPath := flag.String("magic-dictionary",
		"assets/original-derived-data/magic.json", "法術參數字典 typed data")
	scenesPath := flag.String("wilderness-scenes",
		"assets/original-derived-data/wilderness-scenes.json",
		"野外治療那一條的六段敘述 typed data")
	flowScript := flag.String("flow-script", "",
		"整條遊玩路徑的操作腳本，空白分隔；從標題序列開始，經書本、角色選項、"+
			"角色選擇清單、旅店走到戶外地圖")
	playScript := flag.String("play-script", "",
		"城鎮遊玩的操作腳本，空白分隔：L R U D 是方向鍵、"+
			"A S Z X 是斜走、< > 移動圖示白框、RET 執行")
	outdoorAtEdge := flag.Bool("outdoor-at-edge", false,
		"戶外畫面畫成走到地圖邊界的城門提示：地圖視窗換成林地圖，圖示列 DOOR 與 U-TURN")
	outdoorSelected := flag.Int("outdoor-selected", 0, "戶外圖示列白框的位置")
	outdoorAtDoor := flag.Bool("outdoor-at-door", false,
		"強制進入門口狀態（原版是移動到門的座標時才觸發）")
	outdoorIcons := flag.String("outdoor-icons", "", "覆寫戶外圖示列，逗號分隔的 KOLMPICS 圖號；`none` 代表整列淨空")
	deleteScript := flag.String("delete-script", "",
		"delete 場景的按鍵腳本，token 為 UP／DOWN／RET／y／n，逐步輸出截圖")
	rosterPath := flag.String("roster",
		"assets/original-derived-data/character-roster.json", "角色槽位 typed data")
	shot := flag.String("shot", "", "單張截圖輸出路徑")
	shotDir := flag.String("shot-dir", "", "批次截圖輸出目錄")
	script := flag.String("script", "", "以空白分隔的操作腳本；每一步後截一張圖")
	seed := flag.Int("seed", 1, "亂數種子，供可重播的屬性擲骰與標題序列使用")
	ambushRolls := flag.String("ambush-rolls", "",
		"測試用受控亂數：以逗號分隔，依序當作野外遭遇擲點（RollAmbush）的骰值，用完回到種子亂數")
	titlePath := flag.String("title-sequence", "docs/re/title-sequence.json",
		"標題序列腳本（由 RE-047 從 TITLE.EXE 的 _main 轉寫）")
	titleShots := flag.String("title-shot-dir", "", "把標題序列一輪的每個畫面輸出到這個目錄")
	titleFrame := flag.Int("title-frame", -1,
		"只輸出標題序列展開後的第 N 個畫面到 -shot，供展示圖使用")
	bookAnimPath := flag.String("book-animation",
		"assets/original-derived-data/book-open-animation.json", "翻書動畫參數表")
	bookShots := flag.String("book-shot-dir", "",
		"把翻書動畫的每一步輸出到這個目錄")
	bookTransition := flag.String("book-transition", "open",
		"翻書動畫的段落：open 是封面翻到目錄頁，menu 是目錄頁翻到角色選項頁")
	titleTimeline := flag.String("title-timeline", "",
		"把一輪標題序列的時間軸寫成 JSON 後結束")
	titleChoice := flag.String("title-choice", "rng",
		"標題序列兩處亂數的取法：rng 依 seed 重播，first／second 固定取第一或第二個選項")
	scaleFlag := flag.Int("scale", 3, "視窗放大倍率")
	listAssets := flag.Bool("list-assets", false, "只列出 manifest 摘要後結束")
	writeRecord := flag.String("chardata-write", "",
		"把一筆角色記錄寫成 JSON 後結束，供 tools/verify_chardata_write.py 對拍")
	writeTemplate := flag.String("chardata-template", "",
		"槽位樣板檔（本機的原版 CHARDATA），與 -chardata-slot 搭配")
	writeSlot := flag.Int("chardata-slot", 14, "要寫入的槽位編號")
	writeName := flag.String("chardata-name", "zed", "要寫入的角色名字")
	writeRace := flag.Int("chardata-race", 0, "種族 enum")
	writeSecondary := flag.Int("chardata-secondary", 0, "secondary selector")
	writeProfession := flag.Int("chardata-profession", 0, "職業記錄索引")
	writeStats := flag.String("chardata-stats", "", "七個逗號分隔的內部屬性值")
	writeScript := flag.String("chardata-script", "",
		"改用操作腳本跑完整個建立流程，把結果寫進槽位；與 -chardata-stats 二選一")
	writeOut := flag.String("chardata-out", "",
		"把寫入後的完整 CHARDATA 存到這個檔案（含原版樣板 bytes，只作本機比對）")
	deleteSlot := flag.Int("chardata-delete-slot", -1,
		"以 remake 的刪除規則清掉這個槽位，把結果寫到 -chardata-out（本機比對用）")
	itemTemplates := flag.String("chardata-weapons",
		"assets/original-derived-data/weapon-templates.json", "起始武器樣板表")
	writerSummary := flag.String("chardata-contract",
		"docs/re/ida-fingerprints/KOLBOOK-character-writer-summary.json",
		"職業起始契約（IDA 安全摘要）")
	dumpLayout := flag.String("dump-cjk-layout", "",
		"把中文模式的版面座標寫成 JSON 後結束，供文字安全矩形檢查使用")
	lang := flag.String("lang", "en", "介面語言：en 走原版英文字模的嚴格對拍模式，"+
		"zh-Hant 走中文模式（畫布放大三倍、16×18 中文字格，不對拍）")
	uiFontRoot := flag.String("ui-fonts", "assets/ui-font", "中文 UI 字模表目錄")
	savePath := flag.String("save", "", "存檔路徑；客棧住宿時寫進這個檔（SPEC-006）")
	creationPath := flag.String("creation-start", "assets/original-derived-data/creation-start.json", "建角起始 typed 資料")
	loadPath := flag.String("load", "", "開場先讀這份存檔：隊伍、金幣、傷勢與所在位置"+
		"都從存檔來（SPEC-006）")
	faithfulInn := flag.Bool("faithful-inn-stay", false,
		"客棧住宿完全照原版：問「存到硬碟嗎」、角色離隊、隊伍空了問是否離開遊戲。"+
			"與原版對拍時要打開")
	tracePath := flag.String("trace", "",
		"把每一步的狀態軌跡寫成 JSON：場景、城鎮／世界座標、戰鬥階段與結局段號")
	textLog := flag.String("text-log", "",
		"把每個第一次畫出的（字型, 字串）以 JSON 行寫進這個檔，供找中文模式的漏譯")
	traceDetail := flag.Bool("trace-detail", false,
		"軌跡多記指令列、游標與逐步的戰場；產生腳本用，收據不需要")
	routeTo := flag.String("route-to", "",
		"印出從開場位置走到某一格的方向鍵序列（`x,y`）就離開；產生對拍腳本用")
	shotSteps := flag.String("shot-steps", "",
		"只在這幾步截圖（逗號分隔的步號）；長路徑用它留檢查點，其餘只記軌跡")
	stringsPath := flag.String("strings", "docs/i18n/newchar-zh-Hant.json",
		"中文 UI 字串表")
	logInput := flag.Bool("log-input", false,
		"互動遊玩時把每個輸入 token 與當時的畫面印到 stderr，供輸入驗收判讀")
	flag.Parse()
	if *showVersion {
		fmt.Println(buildVersion)
		return
	}
	// 發行包被雙擊時工作目錄不是程式所在地，先把它擺正再解析任何相對路徑。
	useResourceDir(*assetRoot)
	if *textLog != "" {
		if err := startTextLog(*textLog); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	// `-scene game` 是互動遊玩。它借 `flow` 那條路徑的全部初始化，差別只在 token
	// 來自鍵盤與滑鼠而不是腳本，所以這裡改寫成 flow，另外記一個旗標給 Update。
	//
	// 但**帶了產檔旗標就不是互動遊玩**。那些流程（截圖、腳本、傾印）本來就靠
	// `-scene` 的預設值，預設值從封面改成互動之後要把它們留在封面上，否則
	// `docs/screenshots/shots.spec` 裡不寫 `-scene` 的那幾行會開起視窗等人按鍵，
	// 一張圖也產不出來。列旗標名而不是列條件，是因為新增一個產檔旗標時漏掉
	// 這裡的代價是「批次流程靜靜地卡住」，很難查。
	interactive := scene(*sceneName) == sceneGame
	if interactive {
		flag.Visit(func(f *flag.Flag) {
			if batchFlags[f.Name] {
				interactive = false
			}
		})
	}
	if !interactive && scene(*sceneName) == sceneGame {
		*sceneName = string(sceneCover)
	}
	if interactive {
		*sceneName = string(sceneFlow)
		// 這是繁體中文重製版，玩的時候預設就走中文模式；`-lang en` 切回原版
		// 字模的英文版面。對拍腳本一律明寫 `-scene`，不會走到這裡。
		langSet := false
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "lang" {
				langSet = true
			}
		})
		if !langSet {
			*lang = "zh-Hant"
		}
		// 中文模式的畫布是 960×600（`book.CJKScale` 取 3 才放得下 18 px 的字模格），
		// 預設三倍就是 2880×1800——放不進一般螢幕。沒指定倍率時挑一個放得下的：
		// 960×600 再乘一倍就是 1920×1200，所以中文模式預設不再放大。
		scaleSet := false
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "scale" {
				scaleSet = true
			}
		})
		if !scaleSet && *lang != "en" {
			*scaleFlag = 1
		}
		// 存檔要寫在使用者自己的資料目錄。互動模式沒有這一段的話住宿存檔會靜默
		// 跳過（`-save` 空字串就不寫），而 AppImage 的工作目錄是唯讀掛載點，
		// 寫回程式旁邊也不行。
		if *savePath == "" {
			if p, err := defaultSavePath(); err == nil {
				*savePath = p
			} else {
				fmt.Fprintln(os.Stderr, "找不到可寫的存檔目錄，無法開始可保存的遊戲：", err)
				os.Exit(1)
			}
		}
	}

	// 旗標預設維持 `en`：沒寫 `-lang` 的對拍計畫要照舊產出 320×200 英文畫格。
	// 互動遊玩在上面已切成繁中，所以這裡才檢查譯表，雙擊啟動也涵蓋得到。
	if *lang != "en" {
		for _, path := range []string{*townI18nPath, *uiI18nPath, *displayNamesPath, *spellWordsPath, *titleSubsPath} {
			info, err := os.Stat(path)
			if err != nil || !info.Mode().IsRegular() {
				fmt.Fprintf(os.Stderr, "繁體中文譯表無法讀取：%s（%v）\n", path, err)
				os.Exit(1)
			}
		}
	}

	lib, err := assets.Load(*assetRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *deleteSlot >= 0 {
		if *writeTemplate == "" || *writeOut == "" {
			fmt.Fprintln(os.Stderr, "-chardata-delete-slot 需要 -chardata-template 與 -chardata-out")
			os.Exit(2)
		}
		raw, err := os.ReadFile(*writeTemplate)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := chardata.DeleteSlot(raw, *deleteSlot); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.WriteFile(*writeOut, raw, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *writeRecord != "" {
		spec := chardata.NewCharacter{
			Name: *writeName, Race: *writeRace, Secondary: *writeSecondary,
			ProfessionIndex: *writeProfession,
		}
		stats := *writeStats
		if *writeScript != "" {
			createData, err := chardata.Load(*dataPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			st := charcreate.New(createData, newRand(*seed))
			for _, token := range strings.Fields(*writeScript) {
				if err := applyToken(st, token); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(2)
				}
			}
			if st.Stage != charcreate.StageDone {
				fmt.Fprintf(os.Stderr, "腳本沒有走到完成，停在第 %d 層\n", st.Stage)
				os.Exit(2)
			}
			spec = chardata.NewCharacter{Name: st.Name, Race: st.Race,
				Secondary:       st.Secondary,
				ProfessionIndex: st.ProfessionRecordIndex()}
			values := make([]string, 0, 7)
			for _, v := range st.Primary {
				values = append(values, strconv.Itoa(v))
			}
			stats = strings.Join(values, ",")
		}
		if err := runCharacterWrite(*writeRecord, *writeTemplate, *writeSlot,
			*writerSummary, *itemTemplates, *writeOut, spec, stats); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *dumpLayout != "" {
		raw, err := json.MarshalIndent(map[string]any{
			"schema": "kol-cjk-layout-v1", "scale": book.CJKScale,
			"cell": book.CJKCell, "layout": book.CJKLayout, "text_box": book.CJKTextBox,
			"confirm": book.CJKConfirm, "town_inn": town.CJKInn,
			"town_outdoor": town.CJKOutdoor, "world": world.CJK,
		}, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.WriteFile(*dumpLayout, append(raw, '\n'), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *listAssets {
		m := lib.Manifest()
		fmt.Printf("source_zip=%s sha256=%s archives=%d pngs=%d\n",
			m.SourceZip.Name, m.SourceZip.SHA256, m.ArchiveNum, m.PNGCount)
		for _, a := range m.Archives {
			fmt.Printf("  %-10s %d\n", a.Name, len(a.Images))
		}
		return
	}
	small, err := font.Load(*fontRoot, "book-small")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	large, err := font.Load(*fontRoot, "book-large")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var cjk *font.Font
	var table *i18n.Table
	if *lang != "en" {
		cjk, err = font.Load(*uiFontRoot, "zh-hant-16")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		table, err = i18n.Load(*stringsPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// 書本畫面的標籤走 BOOKDATA 那一份譯文：那份文字庫逐筆翻過
		// 也驗過，`-strings` 只留 remake 自己加的字串。
		if _, err := table.LoadBank(*uiI18nPath, town.BankBook); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := table.LoadDisplayNames(*displayNamesPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		dict, err := magicdata.Load(*magicPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := table.LoadSpellWords(*spellWordsPath, dict.SubclassName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	var slots *roster.Roster
	if scene(*sceneName) == sceneRoster || scene(*sceneName) == sceneDelete ||
		scene(*sceneName) == sceneFlow || *deleteScript != "" ||
		(*bookShots != "" && *bookTransition == "roster") {
		if slots, err = roster.Load(*rosterPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	var innState *town.Inn
	if scene(*sceneName) == sceneInn {
		if innState, err = town.LoadInn(*innPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if *innSelected >= 0 {
			if *innSelected >= len(innState.Icons) {
				fmt.Fprintf(os.Stderr, "-inn-selected %d 超出範圍\n", *innSelected)
				os.Exit(2)
			}
			innState.Selected = *innSelected
		}
	}

	var worldScreenState *world.Screen
	var worldSessionState *world.Session
	var worldPartyState []town.Member
	// quest 是走到任務地點、答應之後開起來的那一場戰鬥。
	quest := &questBattle{}
	var outdoorState *town.Outdoor
	if scene(*sceneName) == sceneOutdoor {
		if outdoorState, err = town.LoadOutdoor(*outdoorPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if *outdoorCamera != "" {
			var cx, cy int
			if _, err := fmt.Sscanf(*outdoorCamera, "%d,%d", &cx, &cy); err != nil {
				fmt.Fprintf(os.Stderr, "-outdoor-camera 格式應為 X,Y：%v\n", err)
				os.Exit(2)
			}
			outdoorState.Camera.X, outdoorState.Camera.Y = cx, cy
		}
		if *outdoorFrame >= 0 {
			if *outdoorFrame >= town.MarkerFrames {
				fmt.Fprintf(os.Stderr, "-outdoor-frame %d 超出範圍\n", *outdoorFrame)
				os.Exit(2)
			}
			outdoorState.Marker.Frame = *outdoorFrame
		}
		if *outdoorMarker != "" {
			var mr, mc int
			if _, err := fmt.Sscanf(*outdoorMarker, "%d,%d", &mr, &mc); err != nil {
				fmt.Fprintf(os.Stderr, "-outdoor-marker 格式應為 列,欄：%v\n", err)
				os.Exit(2)
			}
			outdoorState.Marker.Row, outdoorState.Marker.Col = mr, mc
		}
		if *outdoorIcons != "" {
			icons, err := parseIcons(*outdoorIcons)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
			outdoorState.Icons = icons
			outdoorState.Selected = 0
		}
		if *outdoorSelected != 0 {
			outdoorState.Selected = *outdoorSelected
			outdoorState.DoorCursor = *outdoorSelected
		}
		grid, err := loadGrid(lib, outdoorState.Archive())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		outdoorState.SetGrid(grid)
		markerSrc, _, err := lib.Decode(town.MarkerArchive, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		markers, err := town.LoadMarkers(markerSrc)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		outdoorState.SetMarkers(markers)
		buildings, err := town.LoadBuildings(*buildingsPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		info := buildings.Town(outdoorState.Town)
		if info == nil {
			fmt.Fprintf(os.Stderr, "城鎮建築資料沒有城鎮 %d\n", outdoorState.Town)
			os.Exit(1)
		}
		outdoorState.SetBuildings(info)
		if *outdoorAtDoor {
			if outdoorState.DoorUnderParty() == nil {
				fmt.Fprintln(os.Stderr, "-outdoor-at-door：隊伍座標上沒有門")
				os.Exit(2)
			}
			outdoorState.AtDoorCell = true
		}
		if *outdoorAtEdge {
			outdoorState.AtEdge = true
		}
	}

	// 野外那一整套（地圖資料、天候、時刻、任務地點）也是 flow 走完城鎮之後要用的：
	// 城門選 DOOR 之後接手畫面的就是它，少建起來 flow 就只會停在城牆邊。
	if scene(*sceneName) == sceneWorld || scene(*sceneName) == scenePlay ||
		scene(*sceneName) == sceneFlow {
		data, err := world.LoadData(*worldPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		src, _, err := lib.Decode("TOWN0", 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		grid, err := world.LoadGrid(src)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		sample, err := town.LoadOutdoor(*outdoorPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		px, py := 40, 23
		if *worldPos != "" {
			if px, py, err = parsePair(*worldPos); err != nil {
				fmt.Fprintf(os.Stderr, "-world-pos 格式應為 X,Y：%v\n", err)
				os.Exit(2)
			}
		}
		travel := world.New(data, grid, px, py)
		if *worldSub != "" {
			sx, sy, err := parsePair(*worldSub)
			if err != nil {
				fmt.Fprintf(os.Stderr, "-world-sub 格式應為 次格X,次格Y：%v\n", err)
				os.Exit(2)
			}
			travel.SubX, travel.SubY = sx, sy
		}
		weather, err := world.LoadWeatherGroups(func(index int) (image.Image, error) {
			img, _, err := lib.Decode(world.WeatherArchive, index)
			return img, err
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		bank, err := town.LoadUIText(*uiTextPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		trans, err := loadTranslations(*townI18nPath, *uiI18nPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		names := world.BankNamer{
			Padded: func(index int) string { return bank.RecordPadded(town.BankUI, index) }}
		if *lang != "en" {
			names.Zh = trans.UIRecord
		}
		worldPartyState = sample.Party
		worldScreenState = &world.Screen{Travel: travel, Weather: weather,
			Group: *worldWeather, Phase: *worldPhase, TimeIndex: *worldTime,
			Date: sample.Date, DateZh: sample.DateZh, Names: names,
			CampText: bank.RecordPadded(town.BankUI, world.CampRecord)}
		worldSessionState = world.NewSession(worldScreenState)
	}

	if scene(*sceneName) == sceneBuilding {
		buildings, err := town.LoadBuildings(*buildingsPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		info := buildings.Town(*buildingTown)
		if info == nil {
			fmt.Fprintf(os.Stderr, "城鎮建築資料沒有城鎮 %d\n", *buildingTown)
			os.Exit(1)
		}
		var door *town.Door
		for i := range info.Doors {
			if info.Doors[i].Index == *buildingDoor {
				door = &info.Doors[i]
			}
		}
		if door == nil {
			fmt.Fprintf(os.Stderr, "城鎮 %d 沒有門 %d\n", *buildingTown, *buildingDoor)
			os.Exit(1)
		}
		texts, err := town.LoadText(*townTextPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		sample, err := town.LoadInn(*innPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		innState = town.EnterBuilding(fmt.Sprintf("TWN%dPICS", *buildingTown), door,
			texts.Building(*buildingTown, *buildingDoor), sample.Party)
		if *lang != "en" {
			tr, err := loadTranslations(*townI18nPath, *uiI18nPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			innState.Apply(tr.Door(*buildingTown, *buildingDoor))
		}
	}

	var playSession *town.Session
	importedLegacy := false
	var arenaDep *arenaDeps
	if scene(*sceneName) == scenePlay || scene(*sceneName) == sceneFlow {
		state, err := town.LoadOutdoor(*outdoorPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// 讀檔要在載入格子陣列之前決定城鎮編號——地圖檔是逐城鎮的（SPEC-006）。
		var loaded *save.Save
		if *loadPath != "" {
			if loaded, err = save.Read(*loadPath); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			state.Party = savedParty(loaded)
			if loaded.Schema == save.Schema {
				importedLegacy = true
				fmt.Fprintln(os.Stderr, "正在匯入 v1：舊格式未保存的能力、裝備變更與世界時間無法還原，缺項使用原始名冊與起始時間。原始 v1 檔案保留。")
				if *savePath == "" {
					*savePath = *loadPath + ".v2.json"
				}
				a, _ := filepath.Abs(*savePath)
				b, _ := filepath.Abs(*loadPath)
				sameFile := false
				if src, e := os.Stat(*loadPath); e == nil {
					if dst, e := os.Stat(*savePath); e == nil {
						sameFile = os.SameFile(src, dst)
					}
				}
				if a == b || sameFile {
					*savePath = *loadPath + ".v2.json"
				}
				if src, e := os.Stat(*loadPath); e == nil {
					if dst, e := os.Stat(*savePath); e == nil && os.SameFile(src, dst) {
						fmt.Fprintln(os.Stderr, "v2 目的地指向原始 v1，拒絕覆寫")
						os.Exit(1)
					}
				}
			}
			if loaded.Location.Kind == save.LocationTown {
				state.Town = loaded.Location.Town
				state.EntryPoint = loaded.Location.EntryPoint
			}
		}
		grid, err := loadGrid(lib, state.Archive())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		state.SetGrid(grid)
		markerSrc, _, err := lib.Decode(town.MarkerArchive, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		markers, err := town.LoadMarkers(markerSrc)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		state.SetMarkers(markers)
		buildings, err := town.LoadBuildings(*buildingsPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		info := buildings.Town(state.Town)
		if info == nil {
			fmt.Fprintf(os.Stderr, "城鎮建築資料沒有城鎮 %d\n", state.Town)
			os.Exit(1)
		}
		texts, err := town.LoadText(*townTextPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		trans, err := loadTranslations(*townI18nPath, *uiI18nPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if *lang == "en" {
			trans = nil
		}
		playSession = town.NewSession(state, info, texts, trans, state.Party)
		if trans != nil && table != nil {
			playSession.DisplayName = table.DisplayName
			playSession.SpellGloss = table.SpellGloss
		}
		topics, err := town.LoadTopics(*topicsPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		playSession.Topics = topics
		playSession.CJK = *lang != "en"
		if rosterState, err := roster.Load(*rosterPath); err == nil {
			if slots == nil {
				slots = rosterState
			}
			for _, slot := range rosterState.Occupied() {
				playSession.Roster = append(playSession.Roster, town.Candidate{
					Name: slot.Name, Portrait: slot.Portrait,
					Place: slot.Place, Town: slot.Town,
					Female: slot.Female, Wounds: slot.Wounds,
					Gold: slot.Gold, BuildingID: slot.Building,
					Quests: slot.Quests, Sheet: townSheet(slot.Sheet),
					Worn:     townWorn(slot.Equipment),
					InHand:   townGood(slot.InHand),
					Backpack: townBackpack(slot.Backpack),
					Spells:   townSpells(slot.Spells),
					// 任務交付要逐 byte 比，所以帶原始位元組（RE-109）。
					Progress: slot.Progress,
					Carried:  townCarried(slot.Carried)})
			}
			if loaded != nil && loaded.Characters != nil {
				if err := restoreRoster(loaded, slots, playSession); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
			// 畫面樣本只寫隊伍有誰，錢包與所在地在角色檔裡——住宿與治療要花的
			// 就是這一份錢（RE-073），所以照名字補齊。
			for i := range playSession.Party {
				if loaded != nil && loaded.Characters != nil {
					continue
				}
				for _, c := range playSession.Roster {
					if c.Name != playSession.Party[i].Name {
						continue
					}
					if loaded == nil && playSession.Party[i].Gold == 0 {
						playSession.Party[i].Gold = c.Gold
					}
					if playSession.Party[i].Place == "" {
						playSession.Party[i].Place = c.Place
						playSession.Party[i].TownCode = c.Town
						playSession.Party[i].BuildingID = c.BuildingID
					}
					if playSession.Party[i].Sheet == nil {
						playSession.Party[i].Sheet = c.Sheet
					}
					if playSession.Party[i].Quests == nil {
						playSession.Party[i].Quests = c.Quests
					}
					// 帳面位階與冒險點數存檔帶了就以存檔為準：前者競技場打贏會
					// 加一（RE-108），後者打贏一場加一份、訓練花掉一份
					// （RE-096／RE-076）。角色檔那兩份是遊玩之前的值。
					if loaded != nil && i < len(loaded.Party) &&
						playSession.Party[i].Sheet != nil &&
						(loaded.Party[i].Rank != nil ||
							loaded.Party[i].AdventurePoints != nil ||
							loaded.Party[i].Steed != nil ||
							loaded.Party[i].Order != nil) {
						// 複製一份再改：資料頁那份是與名冊共用的指標，
						// 直接改會連招募清單上的數字一起動到。
						sheet := *playSession.Party[i].Sheet
						if loaded.Party[i].Rank != nil {
							sheet.Rank = *loaded.Party[i].Rank
						}
						if loaded.Party[i].AdventurePoints != nil {
							sheet.AdventurePoints = *loaded.Party[i].AdventurePoints
						}
						// 買來的坐騎帶著移動率（角色記錄 +0x1BA 落在坐騎那
						// 十六個 byte 裡，RE-119）；沒帶就維持角色檔的值。
						if st := loaded.Party[i].Steed; st != nil {
							sheet.MovementRate = st.Movement
						}
						// 入會之後資料頁那一列印的是教派與頭銜（RE-121）。
						if o := loaded.Party[i].Order; o != nil {
							sheet.Order, sheet.Title = o.Name, o.Title
						}
						playSession.Party[i].Sheet = &sheet
					}
					// 已回報的任務數同理：存檔帶了就用存檔的，
					// 沒帶才照角色檔補（RE-107）。
					if playSession.Party[i].Progress == nil {
						playSession.Party[i].Progress = c.Progress
					}
					if playSession.Party[i].Carried == nil {
						playSession.Party[i].Carried = c.Carried
					}
					// 身上穿的東西：存檔沒帶就照角色檔補（RE-080）。
					empty := true
					for _, g := range playSession.Party[i].Worn {
						if g != nil {
							empty = false
							break
						}
					}
					if empty {
						playSession.Party[i].Worn = c.Worn
					}
					// 手上那件不在穿戴欄裡（RE-127），要另外補。
					if playSession.Party[i].InHand == nil {
						playSession.Party[i].InHand = c.InHand
					}
					if playSession.Party[i].Backpack == nil {
						playSession.Party[i].Backpack = append([]town.Good(nil), c.Backpack...)
					}
					// 法術同理：存檔沒帶就照角色檔補（RE-100）。
					if playSession.Party[i].Spells == nil {
						playSession.Party[i].Spells = c.Spells
					}
					break
				}
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		bank, err := town.LoadUIText(*uiTextPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		playSession.UIText = bank
		// 立繪的版面表：讀不到就只是資料頁那兩頁畫不出人，其他照常（RE-126）。
		if layout, err := portraitdata.Load(*portraitPath); err == nil {
			playSession.Portraits = layout
		}
		// 法術的參數字典：讀不到就只是法術編輯器那一頁每一欄都空著（RE-128）。
		if dict, err := magicdata.Load(*magicPath); err == nil {
			playSession.Magic = dict
		}
		if worldScreenState != nil {
			playSession.LoadStartingTown = func(code string) ([]town.TownEntry, *town.TownBuildings, *town.Grid, error) {
				recs := worldScreenState.Travel.Data.TownsByCode(code)
				if len(recs) == 0 {
					return nil, nil, nil, fmt.Errorf("找不到城鎮代碼：%s", code)
				}
				var entries []town.TownEntry
				for _, rec := range recs {
					if rec.Town != recs[0].Town {
						return nil, nil, nil, fmt.Errorf("城鎮代碼 %s 對到不同的城鎮檔", code)
					}
					entries = append(entries, town.TownEntry{Code: rec.Code, Town: rec.Town,
						Camera: town.Point{X: rec.Camera.X, Y: rec.Camera.Y},
						MinX:   rec.Bounds.MinX, MaxX: rec.Bounds.MaxX,
						MinY: rec.Bounds.MinY, MaxY: rec.Bounds.MaxY,
						EntryPoint: rec.EntryPoint})
				}
				info := buildings.Town(recs[0].Town)
				if info == nil {
					return nil, nil, nil, fmt.Errorf("缺少城鎮建築：%d", recs[0].Town)
				}
				grid, err := loadGrid(lib, fmt.Sprintf("TOWN%d", recs[0].Town))
				return entries, info, grid, err
			}
			playSession.TownCode = worldScreenState.Travel.Data.TownCode(
				state.Town, state.EntryPoint)
			playSession.World = worldScreenState
			playSession.OnLeaveTown = func(p town.Point) {
				enterWorld(worldScreenState, p)
				worldSessionState.Reset()
			}
			// 野外選 DOOR 就回到城鎮：鏡頭、夾限、入城點都由世界地圖的記錄給
			// （RE-069），格子陣列與建築資料則換成該城鎮那一份。
			worldSessionState.OnEnterTown = func(rec *world.Town) {
				info := buildings.Town(rec.Town)
				if info == nil {
					fmt.Fprintf(os.Stderr, "城鎮建築資料沒有城鎮 %d\n", rec.Town)
					os.Exit(1)
				}
				grid, err := loadGrid(lib, fmt.Sprintf("TOWN%d", rec.Town))
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				playSession.EnterFromWorld(town.TownEntry{
					Code: rec.Code,
					Town: rec.Town, Camera: town.Point{X: rec.Camera.X, Y: rec.Camera.Y},
					MinX: rec.Bounds.MinX, MaxX: rec.Bounds.MaxX,
					MinY: rec.Bounds.MinY, MaxY: rec.Bounds.MaxY,
					EntryPoint: rec.EntryPoint,
				}, info, grid)
			}
		}
		// 讀檔的落點：鏡頭夾限用該入城點的世界地圖記錄，隊伍放回存檔記的那一格。
		if loaded != nil && loaded.Location.Kind == save.LocationTown &&
			worldScreenState != nil {
			if rec := worldScreenState.Travel.Data.TownEntry(
				state.Town, state.EntryPoint); rec != nil {
				state.Bounds.MinX, state.Bounds.MaxX = rec.Bounds.MinX, rec.Bounds.MaxX
				state.Bounds.MinY, state.Bounds.MaxY = rec.Bounds.MinY, rec.Bounds.MaxY
			}
			state.PlaceParty(loaded.Location.X, loaded.Location.Y)
			playSession.Mode = town.ModeMap
			playSession.SetBarForMode()
		}
		// 產生腳本用：印出走到某一格要按哪幾個方向鍵就離開。對拍的走路段因此
		// 由地圖算出來，不必人工試。
		if *routeTo != "" {
			var tx, ty int
			if _, err := fmt.Sscanf(*routeTo, "%d,%d", &tx, &ty); err != nil {
				fmt.Fprintf(os.Stderr, "-route-to 要寫成 x,y：%v\n", err)
				os.Exit(2)
			}
			path := state.RouteTo(tx, ty)
			if path == nil {
				fmt.Fprintf(os.Stderr, "從 (%d,%d) 走不到 (%d,%d)\n",
					state.Camera.X+state.Marker.Col,
					state.Camera.Y+state.Marker.Row, tx, ty)
				os.Exit(1)
			}
			fmt.Println(strings.Join(path, " "))
			return
		}
		// 住宿存檔（SPEC-006）。沒有給 -save 就不寫檔，流程其餘部分照跑——
		// 截圖與對拍不需要真的產生存檔。
		if loaded != nil {
			requestedFaithful := *faithfulInn
			*faithfulInn = loaded.Settings.FaithfulInnStay
			flag.Visit(func(f *flag.Flag) {
				if f.Name == "faithful-inn-stay" {
					*faithfulInn = requestedFaithful
				}
			})
		}
		playSession.FaithfulInnStay = *faithfulInn
		if *savePath != "" {
			playSession.SaveGame = func() error {
				snapshot, err := buildCompleteSave(playSession, worldScreenState, slots, *faithfulInn)
				if err != nil {
					return err
				}
				if loaded != nil {
					snapshot.Settings.Extra = loaded.Settings.Extra
					snapshot.Origin = loaded.Origin
					if loaded.Schema == save.Schema {
						snapshot.Origin = save.Origin{Kind: "imported", Tool: "kol-v1-to-v2", Note: "由 v1 匯入；舊格式未保存的變更無法還原，缺項以原始名冊與起始時間補齊。"}
					}
				}
				return save.Write(*savePath, snapshot)
			}
		}
		// 野外畫面的隊員格畫的是實際隊伍，不是戶外樣本裡那一份。
		worldPartyState = playSession.Party
		// 任務地點：走進去而且隊上有人帶著那個任務，就蓋出敘述與「Will you partake?」
		// （RE-079）。答應之後進戰場，規劃順序照 RE-080 排，指令與結算走
		// SPEC-008／010 的既有實作。
		if worldSessionState != nil {
			sites, err := battle.LoadSites(*questPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			monsters, err := battle.LoadMonsters(*monsterPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			spoilItems, err := battle.LoadSpoilItems(*battleItemsPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			// 治療那一條的六段敘述（RE-112）。讀不到就只是擲到那一條時當作
			// 沒遇到，其他遭遇照跑。
			healScenes, _ := battle.LoadWildernessScenes(*scenesPath)
			worldSessionState.OnArrive = func(x, y int) *world.Prompt {
				site := sites.At(x, y)
				if site == nil || !partyHasQuest(playSession.Party, site) {
					return nil
				}
				ask := ""
				for _, line := range bank.Lines(town.BankUI, battle.PartakeRecord) {
					if line != "" {
						ask = line
					}
				}
				return &world.Prompt{Kind: world.PromptQuest,
					Lines: site.Narrative, Ask: ask}
			}
			// 騎走那一段的狀態（RE-112）：`ride` 是還沒開打的那一次遭遇，
			// `rideNext` 記著訊息那一頁收掉之後要不要接著開打。
			var ride *battle.Ambush
			var rideNext bool
			// 治療那一條（怪物編號 0xFF）：`heal` 是那一次遭遇，走完騎走那一問
			// 之後接著印走近的敘述再問要不要治療。它不會開打。
			var heal *battle.Ambush
			startAmbush := func(a *battle.Ambush) {
				ranks := make([]int, 0, len(playSession.Party))
				for _, m := range playSession.Party {
					if m.Sheet != nil {
						ranks = append(ranks, m.Sheet.EarnedRank())
					}
				}
				party := len(playSession.Party)
				a.RollForce(battle.PartyRank(ranks, party), party,
					worldSessionState.Rand)
				atlas, err := loadBattlePNG(*battleTiles)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
				members := make([]battle.Member, 0, len(playSession.Party))
				for _, m := range playSession.Party {
					members = append(members, battleMember(m))
				}
				mon := monsters.At(a.Monster)
				st := battle.NewAmbush(a, mon, members,
					battle.Generate(newRand(*seed)))
				if st == nil {
					return
				}
				st.Rand = newRand(*seed)
				st.SpoilItems = spoilItems
				quest.tiles = ui.NewBattleTiles(atlas)
				quest.state = st
				quest.bank = bank
				quest.trans = trans
				quest.english = *lang == "en"
				name := mon.Name
				if table != nil {
					name = table.DisplayName(name)
				}
				quest.lines = ambushLines(bank, trans, a, name, *lang == "en")
			}
			// notice 是只等一個鍵那一頁（`sub_144C3`）：騎走成功或失敗的那一句。
			notice := func(record int) {
				worldSessionState.SetPrompt(&world.Prompt{
					Kind:  world.PromptNotice,
					Lines: nonEmptyLines(bank.Lines(town.BankUI, record))})
			}
			// healScene 是治療那一條走近之後那一頁（`0x2A719`）：印第
			// `Scene + 1` 段，第 4 段那一組只等一個鍵，其餘接記錄 0xE6 的問句。
			healScene := func(a *battle.Ambush) {
				lines := healScenes.Scene(a.Scene + 1)
				if a.Scene == battle.HealerNoAskScene {
					heal = nil
					worldSessionState.SetPrompt(&world.Prompt{
						Kind: world.PromptNotice, Lines: lines})
					return
				}
				worldSessionState.SetPrompt(&world.Prompt{
					Kind: world.PromptHealing, Lines: lines,
					Ask: lastLine(bank.Lines(town.BankUI, battle.HealAskRecord))})
			}
			worldSessionState.OnPromptAnswer = func(yes bool) {
				// 治療那一條：走近那一頁答應就清全隊的傷勢與飢餓，答不答應
				// 都不會開打（`0x2A759`…`0x2A794`）。
				if heal != nil && ride == nil {
					a := heal
					heal = nil
					if yes {
						town.HealParty(playSession.Party)
					}
					_ = a
					return
				}
				// 遭遇那一頁：問要不要騎走，或是騎走結果那一句收掉了。
				if ride != nil {
					a := ride
					if rideNext {
						ride, rideNext = nil, false
						if heal != nil {
							healScene(a)
							return
						}
						startAmbush(a)
						return
					}
					if !yes {
						ride = nil
						if heal != nil {
							healScene(a)
							return
						}
						startAmbush(a)
						return
					}
					if battle.RideAway(a.Monster, worldScreenState.Movement,
						worldSessionState.Rand) {
						ride, heal = nil, nil
						notice(battle.RodeAwayRecord)
						return
					}
					rideNext = true
					notice(battle.StoodGroundRec)
					return
				}
				if !yes {
					return
				}
				site := sites.At(worldScreenState.Travel.X, worldScreenState.Travel.Y)
				if site == nil {
					return
				}
				members := make([]battle.Member, 0, len(playSession.Party))
				for _, m := range playSession.Party {
					members = append(members, battleMember(m))
				}
				atlas, err := loadBattlePNG(*battleTiles)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
				mapImg, err := loadBattlePNG(fmt.Sprintf(*battleMaps, site.Index))
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
				field, err := battle.LoadField(mapImg)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
				// 任務地圖不重生地形，但照樣灑一次樹叢（`sub_1E3A8:0x1E5E5`）。
				field.Scatter(newRand(*seed))
				quest.tiles = ui.NewBattleTiles(atlas)
				st := battle.NewQuest(site, members, monsters.At(site.Monster), field)
				st.Rand = newRand(*seed)
				st.SpoilItems = spoilItems
				st.Stage = battle.StageEncounter
				quest.state = st
				quest.bank = bank
				quest.trans = trans
				quest.english = *lang == "en"
				quest.lines = nonEmptyLines(bank.Lines(town.BankUI,
					battle.QuestCreaturesRecord))
			}
			// 競技場（RE-108）：走到世界地圖 (55,13) 選 DOOR 就進去。
			// 畫面借建築那一套，但沒有肖像——`sub_2B454` 畫的是什麼還沒解。
			if atlas, err := loadBattlePNG(*battleTiles); err == nil {
				arenaImage, err := loadBattlePNG(fmt.Sprintf(*battleMaps, 24))
				if err != nil {
					fmt.Fprintln(os.Stderr, "競技場地圖載入失敗：", err)
					os.Exit(1)
				}
				arenaField, err := battle.LoadField(arenaImage)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				arenaDep = &arenaDeps{monsters: monsters, spoils: spoilItems,
					mapField: arenaField,
					tiles:    ui.NewBattleTiles(atlas), bank: bank, trans: trans,
					rand:    newRand(*seed),
					field:   newRand(*seed),
					english: *lang == "en"}
			}
			worldSessionState.OnEnterArena = func() {
				in := &town.Inn{Party: playSession.Party}
				playSession.EnterArena(in)
			}
			// 野外的隨機遭遇：時段一變就擲一次（SPEC-011）。走進某一格不會擲，
			// 所以一天最多四次。
			ambushScript, err := parseScriptedRolls(*ambushRolls)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
			worldSessionState.OnTimePart = func() {
				if quest.state != nil || worldScreenState == nil {
					return
				}
				c := worldScreenState.Clock
				if c == nil || worldSessionState.Rand == nil {
					return
				}
				rng := battle.Rand(worldSessionState.Rand)
				if ambushScript != nil {
					ambushScript.next = rng
					rng = ambushScript
				}
				ranks := make([]int, 0, len(playSession.Party))
				for _, m := range playSession.Party {
					if m.Sheet != nil {
						ranks = append(ranks, m.Sheet.EarnedRank())
					}
				}
				party := len(playSession.Party)
				rank := battle.PartyRank(ranks, party)
				season := battle.MonthSeason[c.Month%len(battle.MonthSeason)]
				night := worldScreenState.Group == battle.NightWeather
				a := monsters.RollAmbush(worldScreenState.TerrainCode, season,
					night, rank, rng)
				if a == nil {
					return
				}
				// 治療那一條（編號 0xFF）：先印遠景那一段，接下來與一般遭遇
				// 共用騎走那一問；沒騎走就印走近那一段再問要不要治療。
				// 敘述讀不到就當作沒遇到——空白的一頁比沒有那一頁還糟。
				if a.Monster == battle.AmbushHealer {
					if healScenes == nil || len(healScenes.Scene(a.Scene)) == 0 {
						return
					}
					heal = a
					if !battle.CanRideAway(worldScreenState.Movement) {
						healScene(a)
						return
					}
					ride, rideNext = a, false
					worldSessionState.SetPrompt(&world.Prompt{
						Kind:  world.PromptRideAway,
						Lines: healScenes.Scene(a.Scene),
						Ask: lastLine(bank.Lines(town.BankUI,
							battle.RideAwayRecord))})
					return
				}
				// 全隊都有坐騎才問要不要騎走（`0x2A667`）：那一頁印那一隻記錄
				// 末端的四行敘述（RE-112）加記錄 0xE3 的問句。有人用走的就整段
				// 跳過，連那四行都看不到——原版就是這樣。
				if battle.CanRideAway(worldScreenState.Movement) {
					ride, rideNext = a, false
					worldSessionState.SetPrompt(&world.Prompt{
						Kind:  world.PromptRideAway,
						Lines: monsters.At(a.Monster).Encounter,
						Ask: lastLine(bank.Lines(town.BankUI,
							battle.RideAwayRecord))})
					return
				}
				startAmbush(a)
			}
		}
		// 野外的時刻：隊伍最慢的那個移動率決定每步扣多少，換日補營養（RE-077）。
		if worldScreenState != nil {
			worldScreenState.Clock = world.NewClock()
			if loaded != nil {
				restoreWorld(loaded, worldScreenState, playSession)
				if playSession.LeftTown {
					worldSessionState.Sync()
				}
			}
			worldScreenState.Movement = partyMovement(playSession.Party)
			worldScreenState.Date = worldScreenState.Clock.DateLines()
			worldSessionState.Rand = newRand(*seed)
			worldSessionState.OnNewDay = func(hunger int) {
				for _, m := range playSession.Party {
					if m.Sheet == nil {
						continue
					}
					m.Sheet.Nutrition += hunger
					if m.Sheet.Nutrition > 0xFF {
						m.Sheet.Nutrition = 0xFF
					}
				}
			}
		}
		if loaded != nil && loaded.Schema == save.Schema && playSession.SaveGame != nil {
			if err := playSession.SaveGame(); err != nil {
				fmt.Fprintln(os.Stderr, "v1 匯入失敗：", err)
				os.Exit(1)
			}
			fmt.Fprintln(os.Stderr, "v2 匯入完成：", *savePath)
		}
		// LISTEN 的挑槽在截圖與玩家路徑裡必須可重播，因此用固定序列而不是真亂數。
		listenPick := 0
		playSession.Rand = func(n int) int {
			v := listenPick % n
			listenPick++
			return v
		}
	}

	var flowState *flow.Flow
	if scene(*sceneName) == sceneFlow {
		if innState == nil {
			if innState, err = town.LoadInn(*innPath); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	}

	data, err := chardata.Load(*dataPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sequence, err := title.Load(*titlePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 結局場景要自己載文字庫與譯文——它不經過城鎮那一條路（RE-106）。
	// 遊玩那兩條也要載：回報滿 0x17 件任務的那一刻要直接接上結局。
	var endBank *town.UIText
	var endTrans *town.Translations
	if scene(*sceneName) == sceneEnding || scene(*sceneName) == sceneFlow ||
		scene(*sceneName) == scenePlay {
		var err error
		if endBank, err = town.LoadUIText(*uiTextPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if endTrans, err = loadTranslations(*townI18nPath, *uiI18nPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	var titleSubs *titleSubtitles
	if table != nil {
		if titleSubs, err = loadTitleSubtitles(*titleSubsPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	g := &game{
		ending:       &ending.State{},
		endBank:      endBank,
		endTrans:     endTrans,
		sprites:      &spriteCache{lib: lib, m: map[string]*ebiten.Image{}},
		sequence:     sequence,
		small:        small,
		large:        large,
		cjk:          cjk,
		titleSubs:    titleSubs,
		strings:      table,
		canvas:       canvasFor(*lang),
		state:        charcreate.New(data, newRand(*seed)),
		scene:        scene(*sceneName),
		portrait:     *portrait,
		selected:     *menuSelected,
		roster:       slots,
		inn:          innState,
		outdoor:      outdoorState,
		play:         playSession,
		flow:         flowState,
		worldState:   worldScreenState,
		worldSession: worldSessionState,
		worldParty:   worldPartyState,
		quest:        quest,
		tracePath:    *tracePath,
		traceDetail:  *traceDetail,
		arena:        arenaDep,
		arenaFighter: -1,
		interactive:  interactive,
		logInput:     *logInput,
	}
	if slots != nil {
		g.remove = roster.NewDeleteFlow(slots)
	}
	if importedLegacy {
		g.notice = g.saveNotice("import")
	}
	var creationData *creationStart
	if playSession != nil && slots != nil {
		if *loadPath == "" && scene(*sceneName) == sceneFlow {
			playSession.Party = nil
			playSession.Outdoor.Party = nil
			g.worldParty = nil
		}
		creationData, err = loadCreationStart(*creationPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		g.remove.Commit = func() (err error) {
			defer func() {
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					g.notice = g.saveNotice("delete")
				}
			}()
			present := map[string]bool{}
			for _, slot := range slots.Occupied() {
				present[slot.Name] = true
			}
			for _, m := range playSession.Party {
				if !present[m.Name] {
					return fmt.Errorf("角色仍在隊伍中，請先住宿離隊")
				}
			}
			if playSession.SaveGame != nil {
				if err := playSession.SaveGame(); err != nil {
					return err
				}
			}
			var kept []town.Candidate
			for _, c := range playSession.Roster {
				if present[c.Name] {
					kept = append(kept, c)
				}
			}
			playSession.Roster = kept
			return nil
		}
	}
	attachCreation := func(f *flow.Flow) {
		if creationData != nil {
			f.CommitCreate = func(c *charcreate.State) error {
				err := commitCreated(c, creationData, slots, playSession)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					g.notice = g.saveNotice("create")
				}
				return err
			}
		}
	}

	if scene(*sceneName) == sceneBookOpen || *bookShots != "" {
		anim, err := book.LoadOpenAnimation(*bookAnimPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		g.bookAnim = anim
		g.bookLib = lib
		switch *bookTransition {
		case "open":
			g.bookFrom, g.bookTarget = "cover", "contents"
		case "menu":
			g.bookFrom, g.bookTarget = "contents", "menu"
		case "roster":
			// 進入清單時，原版先把反白拿掉再開始捲動：frame-045 還有反白，
			// frame-046 起就沒有了。
			g.bookFrom, g.bookTarget = "menu", "roster"
		default:
			fmt.Fprintf(os.Stderr, "未知的 -book-transition %q\n", *bookTransition)
			os.Exit(2)
		}
	}
	if scene(*sceneName) == sceneTitle {
		// 實際播放模式：依 BIOS tick 推進，按鍵中止，與原版同一條腳本。
		frames, err := sequence.Expand(newRand(*seed))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		g.player = title.NewPlayer(frames, title.ThunderSeconds*title.TickHz)
		g.frame = g.player.Current()
	}

	switch {
	case interactive:
		// 互動遊玩不排 jobs：畫面停在目前這一張，由 Update 讀輸入推進。
		g.scene = sceneFlow
		g.flow = flow.New(g.roster, g.state, g.remove, innState, playSession)
		attachCreation(g.flow)
		if *loadPath != "" {
			g.flow.Scene = flow.SceneTown
			if len(playSession.Party) == 0 {
				g.flow.Scene = flow.SceneMenu
			}
		}
	case *bookShots != "":
		g.scene = sceneBookOpen
		for i := 0; i <= len(g.bookAnim.Steps); i++ {
			job := shotJob{name: fmt.Sprintf("open-%03d", i),
				path: filepath.Join(*bookShots, fmt.Sprintf("open-%03d.png", i))}
			if i > 0 {
				job.apply = func() { g.book.Step() }
			}
			g.jobs = append(g.jobs, job)
		}
	case *titleTimeline != "":
		frames, err := sequence.Expand(newRand(*seed))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		thunder := title.ThunderSeconds * title.TickHz
		times := sequence.Timeline(frames, thunder)
		raw, err := json.MarshalIndent(map[string]any{
			"schema": "kol-title-timeline-v1", "seed": *seed,
			"thunder_seconds": title.ThunderSeconds, "tick_hz": title.TickHz,
			"loop_seconds": times[len(times)-1],
			"frames":       frames, "start_seconds": times[:len(times)-1],
		}, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.WriteFile(*titleTimeline, append(raw, '\n'), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("frames=%d loop_seconds=%.2f\n", len(frames), times[len(times)-1])
		return
	case *titleFrame >= 0:
		if *shot == "" {
			fmt.Fprintln(os.Stderr, "-title-frame 需要搭配 -shot")
			os.Exit(2)
		}
		frames, err := sequence.Expand(newRand(*seed))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if *titleFrame >= len(frames) {
			fmt.Fprintf(os.Stderr, "-title-frame %d 超出 %d 個畫面\n", *titleFrame, len(frames))
			os.Exit(2)
		}
		g.scene = sceneTitle
		g.frame = frames[*titleFrame]
		g.jobs = []shotJob{{name: "title", path: *shot}}
	case *titleShots != "":
		var chooser title.Rand
		switch *titleChoice {
		case "rng":
			chooser = newRand(*seed)
		case "first":
			chooser = fixedChoice(0)
		case "second":
			chooser = fixedChoice(1)
		default:
			fmt.Fprintf(os.Stderr, "未知的 -title-choice %q\n", *titleChoice)
			os.Exit(2)
		}
		frames, err := sequence.Expand(chooser)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		g.scene = sceneTitle
		log := make([]title.Frame, 0, len(frames))
		for i, f := range frames {
			frame := f
			log = append(log, frame)
			g.jobs = append(g.jobs, shotJob{
				name:  fmt.Sprintf("title-%03d", i),
				path:  filepath.Join(*titleShots, fmt.Sprintf("title-%03d.png", i)),
				apply: func() { g.frame = frame },
			})
		}
		g.frame = frames[0]
		raw, err := json.MarshalIndent(map[string]any{
			"schema": "kol-title-frames-v1", "choice": *titleChoice, "seed": *seed,
			"sequence": *titlePath, "frames": log}, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.MkdirAll(*titleShots, 0o755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.WriteFile(filepath.Join(*titleShots, "frames.json"),
			append(raw, '\n'), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *flowScript != "":
		if *shotDir == "" && *shot == "" {
			fmt.Fprintln(os.Stderr, "-flow-script 需要搭配 -shot-dir 或 -shot")
			os.Exit(2)
		}
		g.scene = sceneFlow
		g.flow = flow.New(g.roster, g.state, g.remove, innState, playSession)
		attachCreation(g.flow)
		if *shotDir == "" {
			for _, token := range strings.Fields(*flowScript) {
				if g.leftTown() || g.endingReached() {
					if err := g.playToken(token); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(2)
					}
					continue
				}
				g.flow.Key(token)
			}
			g.jobs = []shotJob{{name: "flow", path: *shot}}
			break
		}
		keep, err := parseSteps(*shotSteps)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		g.jobs = append(g.jobs, shotJob{
			name: "step-000", path: filepath.Join(*shotDir, "step-000.png"),
			skip: keep != nil && !keep[0]})
		for i, token := range strings.Fields(*flowScript) {
			tok := token
			g.jobs = append(g.jobs, shotJob{
				name: fmt.Sprintf("step-%03d", i+1),
				path: filepath.Join(*shotDir, fmt.Sprintf("step-%03d.png", i+1)),
				skip: keep != nil && !keep[i+1],
				apply: func() {
					// 走出城門之後 flow 不再管事，token 交給遊玩那一條
					// （野外、戰鬥與結局）。
					if g.leftTown() || g.endingReached() {
						if err := g.playToken(tok); err != nil {
							fmt.Fprintln(os.Stderr, err)
							os.Exit(2)
						}
						return
					}
					g.flow.Key(tok)
				},
			})
		}
	case *worldScript != "":
		if *shotDir == "" && *shot == "" {
			fmt.Fprintln(os.Stderr, "-world-script 需要搭配 -shot-dir 或 -shot")
			os.Exit(2)
		}
		g.scene = sceneWorld
		if *shotDir == "" {
			for _, token := range strings.Fields(*worldScript) {
				if err := applyWorldToken(g.worldSession, token); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(2)
				}
			}
			g.jobs = []shotJob{{name: "world", path: *shot}}
			break
		}
		g.jobs = append(g.jobs, shotJob{
			name: "step-000", path: filepath.Join(*shotDir, "step-000.png")})
		for i, token := range strings.Fields(*worldScript) {
			tok := token
			g.jobs = append(g.jobs, shotJob{
				name: fmt.Sprintf("step-%03d", i+1),
				path: filepath.Join(*shotDir, fmt.Sprintf("step-%03d.png", i+1)),
				apply: func() {
					if err := applyWorldToken(g.worldSession, tok); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(2)
					}
				},
			})
		}
	case scene(*sceneName) == sceneEnding && *shotDir != "":
		// 結局是七段直線，每段等一個鍵（RE-106）：一段截一張。
		for i := range ending.Panels {
			g.jobs = append(g.jobs, shotJob{
				name:  fmt.Sprintf("panel-%d", i),
				path:  filepath.Join(*shotDir, fmt.Sprintf("panel-%d.png", i)),
				apply: func() { g.ending.Advance() },
			})
		}
	case *playScript != "":
		if *shotDir == "" && *shot == "" {
			fmt.Fprintln(os.Stderr, "-play-script 需要搭配 -shot-dir 或 -shot")
			os.Exit(2)
		}
		g.scene = scenePlay
		if *shotDir == "" {
			for _, token := range strings.Fields(*playScript) {
				if err := g.playToken(token); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(2)
				}
			}
			g.jobs = []shotJob{{name: "play", path: *shot}}
			break
		}
		g.jobs = append(g.jobs, shotJob{
			name: "step-000", path: filepath.Join(*shotDir, "step-000.png")})
		for i, token := range strings.Fields(*playScript) {
			tok := token
			g.jobs = append(g.jobs, shotJob{
				name: fmt.Sprintf("step-%03d", i+1),
				path: filepath.Join(*shotDir, fmt.Sprintf("step-%03d.png", i+1)),
				apply: func() {
					if err := g.playToken(tok); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(2)
					}
				},
			})
		}
	case *deleteScript != "":
		if *shotDir == "" && *shot == "" {
			fmt.Fprintln(os.Stderr, "-delete-script 需要搭配 -shot-dir 或 -shot")
			os.Exit(2)
		}
		g.scene = sceneDelete
		if *shotDir == "" {
			// 只要最後一張：先把整串 token 套完，再截一張圖。
			for _, token := range strings.Fields(*deleteScript) {
				if err := applyDeleteToken(g.remove, token); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(2)
				}
			}
			g.jobs = []shotJob{{name: "delete", path: *shot}}
			break
		}
		g.jobs = append(g.jobs, shotJob{
			name: "step-000", path: filepath.Join(*shotDir, "step-000.png")})
		for i, token := range strings.Fields(*deleteScript) {
			tok := token
			g.jobs = append(g.jobs, shotJob{
				name: fmt.Sprintf("step-%03d", i+1),
				path: filepath.Join(*shotDir, fmt.Sprintf("step-%03d.png", i+1)),
				apply: func() {
					if err := applyDeleteToken(g.remove, tok); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(2)
					}
				},
			})
		}
	case *script != "":
		if *shotDir == "" && *shot == "" {
			fmt.Fprintln(os.Stderr, "-script 需要搭配 -shot-dir 或 -shot")
			os.Exit(2)
		}
		g.scene = sceneScript
		if *lang != "en" {
			g.scene = sceneScriptCJK
		}
		if *shotDir == "" {
			// 只要最後一張：先把整串 token 套完，再截一張圖。
			for _, token := range strings.Fields(*script) {
				if err := applyToken(g.state, token); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(2)
				}
			}
			g.jobs = []shotJob{{name: "script", path: *shot}}
			break
		}
		g.jobs = append(g.jobs, shotJob{
			name: "step-000", path: filepath.Join(*shotDir, "step-000.png")})
		for i, token := range strings.Fields(*script) {
			tok := token
			g.jobs = append(g.jobs, shotJob{
				name: fmt.Sprintf("step-%03d", i+1),
				path: filepath.Join(*shotDir, fmt.Sprintf("step-%03d.png", i+1)),
				apply: func() {
					if err := applyToken(g.state, tok); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(2)
					}
				},
			})
		}
	case *shotDir != "":
		for _, s := range []struct {
			name  string
			scene scene
		}{{"cover", sceneCover}, {"contents", sceneContents}, {"spread", sceneOpenSpread}} {
			sc := s.scene
			g.jobs = append(g.jobs, shotJob{
				name: s.name, path: filepath.Join(*shotDir, s.name+".png"),
				apply: func() { g.scene = sc },
			})
		}
		for i := 0; i < lib.Count("CHARPICS"); i++ {
			index := i
			g.jobs = append(g.jobs, shotJob{
				name:  fmt.Sprintf("newchar-%03d", i),
				path:  filepath.Join(*shotDir, fmt.Sprintf("newchar-%03d.png", i)),
				apply: func() { g.scene, g.portrait = sceneNewChar, index },
			})
		}
		g.scene = sceneCover
	case *shot != "":
		valid := false
		for _, s := range scenes {
			if scene(*sceneName) == s {
				valid = true
			}
		}
		if !valid {
			fmt.Fprintf(os.Stderr, "未知場景 %q\n", *sceneName)
			os.Exit(2)
		}
		g.jobs = []shotJob{{name: *sceneName, path: *shot}}
	}

	bounds := g.canvas.Image().Bounds()
	ebiten.SetWindowSize(bounds.Dx()**scaleFlag, bounds.Dy()**scaleFlag)
	ebiten.SetWindowTitle("武士傳說")
	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// startTextLog 讓字模繪製回報每個畫出的字串；同一組只記第一次，每行立即寫出，
// 批次流程以 os.Exit 結束也不會漏掉。
func startTextLog(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("無法建立文字紀錄：%w", err)
	}
	seen := map[[2]string]bool{}
	font.DrawObserver = func(name, text string) {
		key := [2]string{name, text}
		if seen[key] {
			return
		}
		seen[key] = true
		line, _ := json.Marshal(map[string]string{"font": name, "text": text})
		f.Write(append(line, '\n'))
	}
	return nil
}
