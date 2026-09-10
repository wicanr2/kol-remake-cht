package main

import (
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

func townBackpack(items []*roster.Item) []town.Good {
	var out []town.Good
	for _, item := range items {
		if good := townGood(item); good != nil {
			out = append(out, *good)
		}
	}
	return out
}

// battleItem 保留完整已解碼物品，避免跨戰鬥遺失價格、狀況或任務比對記錄。
func battleItem(g *town.Good) *battle.Item {
	if g == nil {
		return nil
	}
	i := &battle.Item{Name: g.Name, Price: g.Price, Kind: g.Kind, Damage: g.Damage,
		Hands: g.Hands, Class: g.Class, Position: g.Position, Encumbrance: g.Encumbrance,
		Condition: g.Condition, Percent: g.Percent, SizeClass: g.SizeClass, Record: g.Record}
	i.Dice.Count, i.Dice.Sides, i.Dice.Bonus = g.Dice.Count, g.Dice.Sides, g.Dice.Bonus
	i.Dice.Low, i.Dice.High = g.Dice.Low, g.Dice.High
	return i
}

func battleGood(i *battle.Item) *town.Good {
	if i == nil {
		return nil
	}
	return &town.Good{Name: i.Name, Price: i.Price, Kind: i.Kind, Damage: i.Damage,
		Hands: i.Hands, Class: i.Class, Position: i.Position, Encumbrance: i.Encumbrance,
		Condition: i.Condition, Percent: i.Percent, SizeClass: i.SizeClass, Record: i.Record,
		Dice: town.Dice{Count: i.Dice.Count, Sides: i.Dice.Sides, Bonus: i.Dice.Bonus, Low: i.Dice.Low, High: i.Dice.High}}
}

// writeBattleParty 依 SPEC-027 共同出口折算傷勢，保留本場已發生的物品修改。
// 未知掉落機率及額外敗北處置另列缺口，不在此猜測。
func (g *game) writeBattleParty(s *battle.State) {
	if g.play == nil || (s.Verdict != battle.VerdictWon && s.Verdict != battle.VerdictFled && s.Verdict != battle.VerdictLost) {
		return
	}
	for actor, a := range s.Actors {
		if actor >= len(s.Party) || a.Member < 0 || a.Member >= len(s.Party) {
			continue
		}
		i := a.Member
		// 競技場仍保留全隊索引，但觀眾的 Side 是空；不回寫未參戰者。
		if a.Side != battle.SideParty {
			continue
		}
		if i >= len(g.play.Party) {
			continue
		}
		m, result := &g.play.Party[i], s.Party[a.Member]
		m.Wounds = battle.PostBattleWounds(result.Wounds, a.Damage)
		for pos, item := range a.Armour {
			m.Worn[pos] = battleGood(item)
		}
		m.InHand, m.Worn[9] = battleGood(result.HandItem), battleGood(result.BeltItem)
		m.Backpack = nil
		for _, item := range result.PackItems {
			if good := battleGood(item); good != nil {
				m.Backpack = append(m.Backpack, *good)
			}
		}
		m.SyncCarried()
	}
}
