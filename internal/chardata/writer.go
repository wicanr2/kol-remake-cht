package chardata

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// RecordBytes 是一個角色槽位的長度。
const RecordBytes = 768

// SlotCount 是 CHARDATA 的槽位數。
const SlotCount = 16

// NameBytes 是名字欄在槽位內佔用的長度。
const NameBytes = 16

// StartingContract 是一個職業的起始金冠與武器技能，出自已證實的 IDA 摘要。
type StartingContract struct {
	RecordIndex        int    `json:"record_index"`
	MenuName           string `json:"menu_name"`
	GoldCrowns         int    `json:"gold_crowns"`
	PrimaryWeapon      string `json:"primary_weapon"`
	PrimarySkillPair   []int  `json:"primary_skill_pair_internal"`
	SecondaryWeapon    string `json:"secondary_weapon"`
	SecondarySkillPair []int  `json:"secondary_skill_pair_internal"`
}

type writerSummary struct {
	Schema    string             `json:"schema"`
	Contracts []StartingContract `json:"profession_starting_contracts"`
}

// Contracts 是 39 筆職業起始資料的查表。
type Contracts struct {
	byIndex map[int]StartingContract
}

// LoadContracts 讀取已證實的職業起始契約。這份 JSON 是 IDA 安全摘要，只有語意與數值，
// 沒有原版 bytes。
func LoadContracts(path string) (*Contracts, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取職業起始契約失敗：%w", err)
	}
	var s writerSummary
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("解析職業起始契約失敗：%w", err)
	}
	if !strings.Contains(s.Schema, "character-writer-summary") {
		return nil, fmt.Errorf("非預期的摘要 schema：%s", s.Schema)
	}
	if len(s.Contracts) != 39 {
		return nil, fmt.Errorf("職業起始契約有 %d 筆，不是 39", len(s.Contracts))
	}
	out := &Contracts{byIndex: map[int]StartingContract{}}
	for _, c := range s.Contracts {
		out.byIndex[c.RecordIndex] = c
	}
	return out, nil
}

// Get 取得指定職業記錄索引的起始契約。
func (c *Contracts) Get(index int) (StartingContract, bool) {
	v, ok := c.byIndex[index]
	return v, ok
}

// WeaponTemplates 是起始武器樣板表與每個職業的武器 selector。
//
// 樣板來自 `BOOKDATA` 的第 3 個項目（原版載到 dgroup+0x79ee），記錄索引 ＝ selector − 1；
// selector 藏在職業 metadata 第一個 byte 的兩個 nibble 裡。規則見 RE-050。
type WeaponTemplates struct {
	Schema      string   `json:"schema"`
	RecordBytes int      `json:"record_bytes"`
	RecordCount int      `json:"record_count"`
	Records     []string `json:"records"`
	Professions []struct {
		RecordIndex     int  `json:"record_index"`
		PrimaryRecord   int  `json:"primary_record"`
		SecondaryRecord *int `json:"secondary_record"`
	} `json:"professions"`
}

// LoadWeaponTemplates 讀取武器樣板表。
func LoadWeaponTemplates(path string) (*WeaponTemplates, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取武器樣板失敗：%w", err)
	}
	var t WeaponTemplates
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, fmt.Errorf("解析武器樣板失敗：%w", err)
	}
	if t.Schema != "kol-weapon-templates-v1" {
		return nil, fmt.Errorf("非預期的武器樣板 schema：%s", t.Schema)
	}
	if len(t.Records) != t.RecordCount || len(t.Professions) != 39 {
		return nil, fmt.Errorf("武器樣板內容不完整")
	}
	return &t, nil
}

// Weapons 回傳指定職業的主／副武器樣板索引；沒有副武器時 secondary 為 -1。
func (t *WeaponTemplates) Weapons(professionRecord int) (primary, secondary int, ok bool) {
	for _, p := range t.Professions {
		if p.RecordIndex != professionRecord {
			continue
		}
		secondary = -1
		if p.SecondaryRecord != nil {
			secondary = *p.SecondaryRecord
		}
		return p.PrimaryRecord, secondary, true
	}
	return 0, 0, false
}

// Record 回傳一筆武器樣板的 bytes。
func (t *WeaponTemplates) Record(index int) ([]byte, error) {
	if index < 0 || index >= len(t.Records) {
		return nil, fmt.Errorf("武器樣板索引 %d 超出範圍", index)
	}
	return hex.DecodeString(t.Records[index])
}

// NewCharacter 是要寫進槽位的角色。屬性是內部值，不是畫面值。
type NewCharacter struct {
	Name            string
	Race            int
	Secondary       int
	ProfessionIndex int
	Primary         [7]int
}

// 已證實的槽位欄位位移，出自 docs/re/ida-fingerprints/KOLBOOK-character-writer-summary.json。
const (
	offName          = 0x00
	offSecondary     = 0x1a
	offProfession    = 0x1c
	offPrimaryStats  = 0x1d
	offRace          = 0x24
	offRaceSelector  = 0x28
	offSkillSum      = 0x29
	offGold          = 0x2b
	offPrimarySkill1 = 0x54
	offPrimarySkill2 = 0x56
	offSecondSkill1  = 0x61
	offSecondSkill2  = 0x63
	offConstantThree = 0x7f

	// 起始武器樣板：主武器整筆 16 bytes，另外主／副武器各再抄一次前 8 bytes 與第 13 byte。
	offPrimaryTemplate  = 0x30
	offPrimaryPrefix    = 0x4a
	offPrimaryByte13    = 0x52
	offSecondaryPrefix  = 0x57
	offSecondaryByte13  = 0x5f
	weaponTemplateBytes = 16
	weaponPrefixBytes   = 8
	weaponByte13Index   = 13
)

// WriteRecord 把角色寫進槽位樣板，回傳新的 768-byte 記錄。
//
// **只寫已證實來源的欄位**。武器樣板（0x30、0x4a、0x52、0x57、0x5f）來自 `BOOKDATA`
// 第 3 個項目（`sub_11820` 的 `[bx+79EEh]`）；其餘位移原版建立角色時不動，保留樣板
// 原值就是原版的結果。覆蓋率由 tools/verify_chardata_write.py 對原版實際寫入的結果
// 逐 byte 量出來：BARBARIAN 與 RANGER 各 58 個改動 byte 全部相同，沒有漏寫、寫錯或多寫
// （docs/verification/chardata-write/）。
func WriteRecord(template []byte, c NewCharacter, contract StartingContract,
	weapons *WeaponTemplates) ([]byte, error) {
	if len(template) != RecordBytes {
		return nil, fmt.Errorf("樣板長度 %d，不是 %d", len(template), RecordBytes)
	}
	name := []byte(c.Name)
	if len(name) >= NameBytes {
		return nil, fmt.Errorf("名字 %q 超過 %d 個 byte", c.Name, NameBytes-1)
	}
	out := make([]byte, RecordBytes)
	copy(out, template)

	for i := 0; i < NameBytes; i++ {
		if i < len(name) {
			out[offName+i] = name[i]
		} else {
			out[offName+i] = 0
		}
	}
	out[offSecondary] = byte(c.Secondary)
	out[offProfession] = byte(c.ProfessionIndex)
	for i, v := range c.Primary {
		out[offPrimaryStats+i] = byte(v)
	}
	out[offRace] = byte(c.Race)
	out[offRaceSelector] = byte((c.Race + 1) * 16)

	skills := append(append([]int{}, contract.PrimarySkillPair...),
		contract.SecondarySkillPair...)
	sum := 0
	for _, s := range skills {
		sum += s
	}
	out[offSkillSum] = byte(sum / 200)

	// 金冠是 big-endian word：職業表裡是 little-endian 的 0x05DC＝1500，
	// 寫進記錄時是 05 DC。原版整體用 big-endian 存多 byte 值。
	out[offGold] = byte(contract.GoldCrowns >> 8)
	out[offGold+1] = byte(contract.GoldCrowns & 0xFF)

	if len(contract.PrimarySkillPair) == 2 {
		out[offPrimarySkill1] = byte(contract.PrimarySkillPair[0])
		out[offPrimarySkill2] = byte(contract.PrimarySkillPair[1])
	}
	if len(contract.SecondarySkillPair) == 2 {
		out[offSecondSkill1] = byte(contract.SecondarySkillPair[0])
		out[offSecondSkill2] = byte(contract.SecondarySkillPair[1])
	}
	out[offConstantThree] = 3

	if weapons != nil {
		primary, secondary, ok := weapons.Weapons(c.ProfessionIndex)
		if !ok {
			return nil, fmt.Errorf("沒有職業記錄 %d 的武器樣板", c.ProfessionIndex)
		}
		pr, err := weapons.Record(primary)
		if err != nil {
			return nil, err
		}
		copy(out[offPrimaryTemplate:offPrimaryTemplate+weaponTemplateBytes], pr)
		copy(out[offPrimaryPrefix:offPrimaryPrefix+weaponPrefixBytes],
			pr[:weaponPrefixBytes])
		out[offPrimaryByte13] = pr[weaponByte13Index]
		// 沒有副武器的職業，原版不動 0x57 與 0x5f。
		if secondary >= 0 {
			sr, err := weapons.Record(secondary)
			if err != nil {
				return nil, err
			}
			copy(out[offSecondaryPrefix:offSecondaryPrefix+weaponPrefixBytes],
				sr[:weaponPrefixBytes])
			out[offSecondaryByte13] = sr[weaponByte13Index]
		}
	}
	return out, nil
}

// WrittenOffsets 是 WriteRecord 會寫到的槽位位移集合，供覆蓋率統計使用。
func WrittenOffsets() []int {
	out := []int{}
	for i := 0; i < NameBytes; i++ {
		out = append(out, offName+i)
	}
	out = append(out, offSecondary, offProfession)
	for i := 0; i < 7; i++ {
		out = append(out, offPrimaryStats+i)
	}
	out = append(out, offRace, offRaceSelector, offSkillSum, offGold, offGold+1,
		offPrimarySkill1, offPrimarySkill2, offSecondSkill1, offSecondSkill2,
		offConstantThree)
	for i := 0; i < weaponTemplateBytes; i++ {
		out = append(out, offPrimaryTemplate+i)
	}
	for i := 0; i < weaponPrefixBytes; i++ {
		out = append(out, offPrimaryPrefix+i, offSecondaryPrefix+i)
	}
	out = append(out, offPrimaryByte13, offSecondaryByte13)
	return out
}
