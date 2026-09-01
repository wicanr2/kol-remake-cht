package chardata

import (
	"os"
	"testing"
)

const (
	contractPath = "../../docs/re/ida-fingerprints/KOLBOOK-character-writer-summary.json"
	weaponPath   = "../../assets/original-derived-data/weapon-templates.json"
)

// 武器樣板是原版衍生的 typed data，只留在本機；沒有就跳過，不讓測試在乾淨 clone 上假失敗。
func loadWeapons(t *testing.T) *WeaponTemplates {
	t.Helper()
	if _, err := os.Stat(weaponPath); err != nil {
		t.Skipf("缺少 %s（由 tools/extract_weapon_templates.py 從固定原版輸入重生）", weaponPath)
	}
	w, err := LoadWeaponTemplates(weaponPath)
	if err != nil {
		t.Fatalf("載入武器樣板失敗：%v", err)
	}
	return w
}

func TestWriteRecordCoversAllProfessions(t *testing.T) {
	contracts, err := LoadContracts(contractPath)
	if err != nil {
		t.Fatalf("載入職業起始契約失敗：%v", err)
	}
	weapons := loadWeapons(t)
	template := make([]byte, RecordBytes)
	for index := 0; index < 39; index++ {
		contract, ok := contracts.Get(index)
		if !ok {
			t.Fatalf("職業記錄 %d 沒有起始契約", index)
		}
		c := NewCharacter{Name: "t", ProfessionIndex: index}
		for i := range c.Primary {
			c.Primary[i] = 100
		}
		record, err := WriteRecord(template, c, contract, weapons)
		if err != nil {
			t.Fatalf("職業記錄 %d 寫入失敗：%v", index, err)
		}
		primary, secondary, ok := weapons.Weapons(index)
		if !ok {
			t.Fatalf("職業記錄 %d 沒有武器樣板", index)
		}
		pr, err := weapons.Record(primary)
		if err != nil {
			t.Fatalf("職業記錄 %d 的主武器樣板讀取失敗：%v", index, err)
		}
		for i := 0; i < weaponTemplateBytes; i++ {
			if record[offPrimaryTemplate+i] != pr[i] {
				t.Fatalf("職業記錄 %d 的主武器樣板第 %d byte 不符", index, i)
			}
		}
		if record[offPrimaryByte13] != pr[weaponByte13Index] {
			t.Fatalf("職業記錄 %d 的主武器第 13 byte 不符", index)
		}
		if secondary < 0 {
			// 沒有副武器的職業不動 0x57／0x5f，因此應維持樣板原值（此處為 0）。
			if record[offSecondaryPrefix] != 0 || record[offSecondaryByte13] != 0 {
				t.Fatalf("職業記錄 %d 沒有副武器卻寫了副武器欄位", index)
			}
			continue
		}
		sr, err := weapons.Record(secondary)
		if err != nil {
			t.Fatalf("職業記錄 %d 的副武器樣板讀取失敗：%v", index, err)
		}
		for i := 0; i < weaponPrefixBytes; i++ {
			if record[offSecondaryPrefix+i] != sr[i] {
				t.Fatalf("職業記錄 %d 的副武器前綴第 %d byte 不符", index, i)
			}
		}
	}
}

func TestWriteRecordFieldSemantics(t *testing.T) {
	contracts, err := LoadContracts(contractPath)
	if err != nil {
		t.Fatalf("載入職業起始契約失敗：%v", err)
	}
	contract, _ := contracts.Get(0)
	c := NewCharacter{Name: "zed", Race: 2, Secondary: 1, ProfessionIndex: 0,
		Primary: [7]int{136, 91, 104, 115, 101, 94, 99}}
	record, err := WriteRecord(make([]byte, RecordBytes), c, contract, nil)
	if err != nil {
		t.Fatalf("寫入失敗：%v", err)
	}
	if string(record[0:3]) != "zed" || record[3] != 0 {
		t.Fatalf("名字欄不是 NUL 結尾的 zed")
	}
	if record[offRace] != 2 || record[offSecondary] != 1 {
		t.Fatalf("種族或 secondary 欄不符")
	}
	if record[offRaceSelector] != (2+1)*16 {
		t.Fatalf("種族 selector 應為 (race+1)*16")
	}
	for i, want := range c.Primary {
		if int(record[offPrimaryStats+i]) != want {
			t.Fatalf("第 %d 個屬性寫成 %d，預期 %d", i, record[offPrimaryStats+i], want)
		}
	}
	// 金冠是 big-endian：1500 ＝ 0x05DC。
	if record[offGold] != 0x05 || record[offGold+1] != 0xDC {
		t.Fatalf("金冠不是 big-endian 的 0x05DC：%02x %02x",
			record[offGold], record[offGold+1])
	}
	if record[offConstantThree] != 3 {
		t.Fatalf("0x7f 應為常數 3")
	}
}
