package town

import "testing"

func TestSellUsesBackpackSlotAndKeepsOtherEquipment(t *testing.T) {
	s := shopSession(t, 100, [2]int{4, 2}, armourerGoods())
	s.door.Price = 500
	s.UIText = shopText()
	s.Party[0].InHand = &Good{Name: "KEEP", Price: 200}
	s.Party[0].Backpack = []Good{{}, {Name: "Halberd", Price: 170}, {}, {Name: "Fur Cap", Price: 32}}
	s.StartSell()
	if got := s.Inside.List.Rows; len(got) != 2 || got[0] != "Halberd" {
		t.Fatalf("%v", got)
	}
	s.Key("D")
	s.Key("RET")
	if s.Party[0].Gold != 100 {
		t.Fatal("確認前入帳")
	}
	s.Key("RET")
	m := s.Party[0]
	if m.Gold != 132 || len(m.Backpack) != 2 || m.Backpack[0].Name != "" || m.Backpack[1].Name != "Halberd" || m.InHand.Name != "KEEP" {
		t.Fatalf("%+v", m)
	}
}

func TestSellCancelEmptyAndInsufficientShopFunds(t *testing.T) {
	for _, mode := range []string{"empty", "cancel-list", "cancel-confirm", "poor"} {
		t.Run(mode, func(t *testing.T) {
			s := shopSession(t, 100, [2]int{4, 0}, nil)
			s.door.Price = 1
			s.UIText = shopText()
			if mode != "empty" {
				s.Party[0].Backpack = []Good{{Name: "Halberd", Price: 170}}
			}
			s.StartSell()
			if mode == "empty" {
				if s.Shopping() {
					t.Fatal("空背包開清單")
				}
				return
			}
			if mode == "cancel-list" {
				s.Key(">")
				s.Key("RET")
			} else {
				s.Key("RET")
				if mode == "cancel-confirm" {
					s.Key(">")
				}
				s.Key("RET")
			}
			if s.Party[0].Gold != 100 || s.Party[0].Backpack[0].Name != "Halberd" {
				t.Fatal("未成交卻改物品／金冠")
			}
		})
	}
}
