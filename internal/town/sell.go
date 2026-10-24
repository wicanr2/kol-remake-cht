package town

// SPEC-041；sub_1716A、sub_11A93:0x135EF..0x13759。
const (
	SellOfferRecord = 0x34
	EmptyPackRecord = 0x5E
	ShopPoorRecord  = 0x5F
	SoldRecord      = 0x60
)

func (s *Session) StartSell() {
	if s.Mode != ModeInside || s.Inside == nil || s.door == nil || s.door.IconFlags[0]&4 == 0 {
		return
	}
	i := s.speakerIndex()
	if i < 0 {
		return
	}
	sh := &shopState{stage: shopSellList, member: i}
	for slot, g := range s.Party[i].Backpack {
		if g.Name != "" {
			sh.goods = append(sh.goods, g)
			sh.sellSlots = append(sh.sellSlots, slot)
		}
	}
	if len(sh.goods) == 0 {
		s.say(nonEmpty(s.uiLines(EmptyPackRecord)), nil)
		return
	}
	s.shop = sh
	s.Inside.Dialog, s.Inside.DialogZh, s.Inside.Detail = nil, nil, nil
	s.Inside.List = s.shopView()
	s.setInsideIcons([]int{IconYes, IconUTurn})
}

func (s *Session) sellAnswer(id int) {
	sh := s.shop
	if sh.stage == shopSellList {
		if id == IconUTurn {
			s.endShop()
			return
		}
		sh.stage = shopSellConfirm
		s.Inside.List = nil
		s.Inside.Dialog = nonEmpty(s.uiLines(SellOfferRecord))
		s.Inside.DialogZh = nil
		s.Inside.Detail = s.goodDetail(sh.goods[sh.cursor])
		s.setInsideIcons([]int{IconYes, IconNo})
		return
	}
	if id != IconYes {
		s.endShop()
		return
	}
	g := sh.goods[sh.cursor]
	// 固定原版八城的收購上限高位均為零，抽取器另作不變量檢查。
	if g.Price > s.door.Price {
		s.Inside.Detail = nil
		s.shopSay(nonEmpty(s.uiLines(ShopPoorRecord)), s.endShop)
		return
	}
	m := &s.Party[sh.member]
	m.Gold = (m.Gold + g.Price) & 0xFFFFFF
	m.clearPack(sh.sellSlots[sh.cursor])
	m.SyncCarried()
	s.syncParty()
	s.Inside.Detail = nil
	s.shopSay(nonEmpty(s.uiLines(SoldRecord)), s.endShop)
}
