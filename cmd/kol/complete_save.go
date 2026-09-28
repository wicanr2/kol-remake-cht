package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/save"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

func savedCharacter(m town.Member) (*save.Character, error) {
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var c save.Character
	if err = json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}
	// town 的 byte 切片由 JSON 編成 base64；保存契約明確使用 16-byte hex。
	for i := range c.Carried {
		if i < len(m.Carried) {
			c.Carried[i] = hex.EncodeToString(m.Carried[i])
		} else {
			c.Carried[i] = "00000000000000000000000000000000"
		}
	}
	return &c, nil
}

func restoredCharacter(c *save.Character) (town.Member, error) {
	copy := *c
	copy.Carried = [17]string{}
	raw, err := json.Marshal(copy)
	if err != nil {
		return town.Member{}, err
	}
	var m town.Member
	if err = json.Unmarshal(raw, &m); err != nil {
		return m, err
	}
	m.Carried = decodeCarried(c.Carried[:])
	return m, nil
}

func completeParty(s *save.Save) []town.Member {
	var out []town.Member
	for i, slot := range s.Characters.Party {
		m, err := restoredCharacter(s.Characters.Characters[slot])
		if err != nil {
			panic(err)
		}
		m.Speaker = i == s.Characters.Speaker
		out = append(out, m)
	}
	return out
}

func restoreRoster(s *save.Save, r *roster.Roster, session *town.Session) error {
	r.Layout.SlotCount = 16
	r.Slots = make([]roster.Slot, 16)
	session.Roster = nil
	for i, c := range s.Characters.Characters {
		r.Slots[i].Slot = i
		if c == nil {
			continue
		}
		m, err := restoredCharacter(c)
		if err != nil {
			return err
		}
		r.Slots[i] = roster.Slot{Slot: i, ID: c.ID, Name: c.Name, Occupied: true, Place: c.Place, Town: c.TownCode, Portrait: c.Portrait, Gold: c.Gold, Female: c.Female, Wounds: c.Wounds, Building: c.BuildingID}
		cand := town.CandidateFromMember(m)
		cand.Slot = i
		session.Roster = append(session.Roster, cand)
	}
	return nil
}

func buildCompleteSave(s *town.Session, w *world.Screen, r *roster.Roster, faithful bool) (*save.Save, error) {
	if r == nil || len(r.Slots) != 16 || w == nil || w.Clock == nil {
		return nil, fmt.Errorf("完整保存缺少名冊或世界時間")
	}
	out := buildSave(s, w, faithful)
	out.Schema = save.CompleteSchema
	out.Party = nil
	c := w.Clock
	out.Clock = &save.Clock{Ticks: c.Ticks, Part: c.Part, Day: c.Day, Month: c.Month, Year: c.Year, Weather: c.Weather}
	state := &save.Characters{Speaker: -1}
	members := map[string]town.Member{}
	for _, candidate := range s.Roster {
		members[candidate.Name] = candidate.Member()
	}
	for _, m := range s.Party {
		members[m.Name] = m
	}
	indices := map[string]int{}
	for i, slot := range r.Slots {
		if !slot.Occupied {
			continue
		}
		m, ok := members[slot.Name]
		if !ok {
			return nil, fmt.Errorf("名冊角色 %s 缺少完整資料", slot.Name)
		}
		if m.ID == "" {
			m.ID = slot.ID
		}
		if m.ID == "" {
			sum := sha256.Sum256([]byte(fmt.Sprintf("original:%d:%s", i, m.Name)))
			m.ID = fmt.Sprintf("%x", sum[:16])
		}
		character, err := savedCharacter(m)
		if err != nil {
			return nil, err
		}
		state.Characters[i] = character
		indices[m.Name] = i
	}
	for i, m := range s.Party {
		slot, ok := indices[m.Name]
		if !ok {
			return nil, fmt.Errorf("隊員 %s 不在名冊", m.Name)
		}
		state.Party = append(state.Party, slot)
		if m.Speaker || i == 0 {
			state.Speaker = i
		}
	}
	out.Characters = state
	return out, out.Validate()
}

func restoreWorld(s *save.Save, w *world.Screen, session *town.Session) {
	if w == nil {
		return
	}
	w.Travel.X = s.World.X
	w.Travel.Y = s.World.Y
	w.Travel.SubX = s.World.SubX
	w.Travel.SubY = s.World.SubY
	if c := s.Clock; c != nil {
		w.Clock = &world.Clock{Ticks: c.Ticks, Part: c.Part, Day: c.Day, Month: c.Month, Year: c.Year, Weather: c.Weather}
	}
	w.Date = w.Clock.DateLines()
	w.DateZh = nil
	w.TimeIndex = w.Clock.Part
	w.Group = w.Clock.Weather
	w.SyncTerrain()
	session.LeftTown = s.Location.Kind == save.LocationWorld
}
