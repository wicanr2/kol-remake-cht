package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/kol-remake-cht/internal/charcreate"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"os"
)

type creationStart struct {
	Schema        string `json:"schema"`
	Place         string `json:"place"`
	Town          string `json:"town"`
	BuildingID    int    `json:"building_id"`
	HeightBase    []int  `json:"height_base"`
	WeightBase    []int  `json:"weight_base"`
	WeightPerFoot []int  `json:"weight_per_foot"`
	Profiles      []struct {
		RecordIndex int              `json:"record_index"`
		Gold        int              `json:"gold"`
		Rank        int              `json:"rank"`
		Weapons     []town.WeaponRow `json:"weapons"`
		InHand      *town.Good       `json:"in_hand"`
	} `json:"profiles"`
}

func loadCreationStart(path string) (*creationStart, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d creationStart
	if err = json.Unmarshal(raw, &d); err != nil {
		return nil, err
	}
	if d.Schema != "kol-creation-start-v1" || len(d.Profiles) != 39 || len(d.HeightBase) != 7 || len(d.WeightBase) != 7 || len(d.WeightPerFoot) != 7 || d.Place == "" || d.Town == "" {
		return nil, fmt.Errorf("建角起始資料不完整")
	}
	for i, p := range d.Profiles {
		if p.RecordIndex != i || p.InHand == nil {
			return nil, fmt.Errorf("建角職業 %d 不完整", i)
		}
	}
	return &d, nil
}

func commitCreated(c *charcreate.State, d *creationStart, r *roster.Roster, s *town.Session) error {
	if c.Stage != charcreate.StageDone {
		return fmt.Errorf("尚未接受角色肖像")
	}
	slot := -1
	for i, v := range r.Slots {
		if v.Occupied && v.Name == c.Name {
			return fmt.Errorf("名冊已有同名角色")
		}
		if !v.Occupied && slot < 0 {
			slot = i
		}
	}
	if slot < 0 {
		return fmt.Errorf("名冊已滿，最多十六名角色")
	}
	index := c.ProfessionRecordIndex()
	if index < 0 || index >= len(d.Profiles) {
		return fmt.Errorf("職業索引錯誤")
	}
	p := d.Profiles[index]
	height := 2*c.Secondary + c.Race
	if height < 0 || height >= len(d.HeightBase) {
		return fmt.Errorf("體型索引錯誤")
	}
	inches := c.Primary[2] - d.HeightBase[height]
	feet, rem := inches/12, inches%12
	if rem < 0 {
		feet--
		rem += 12
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return err
	}
	m := town.Member{ID: hex.EncodeToString(id[:]), Name: c.Name, Portrait: c.Portrait, Female: c.Secondary != 0,
		Gold: p.Gold, Place: d.Place, TownCode: d.Town, BuildingID: d.BuildingID,
		Sheet: &town.Sheet{Stats: c.Primary, Race: c.Race, Profession: index, Rank: p.Rank, Ward: (c.Race + 1) * 16,
			HeightFeet: feet, HeightInches: rem, Weight: (feet-d.WeightBase[height])*d.WeightPerFoot[height] + rem*4,
			Weapons: append([]town.WeaponRow(nil), p.Weapons...)},
		Quests: make([]int, 6), Progress: make([]int, 5), InHand: p.InHand}
	m = town.CloneMember(m)
	m.SyncCarried()
	previous := r.Slots[slot]
	oldRoster := s.Roster
	r.Slots[slot] = roster.Slot{Slot: slot, ID: m.ID, Name: m.Name, Portrait: m.Portrait, Occupied: true, Place: m.Place, Town: m.TownCode, Gold: m.Gold, Building: m.BuildingID, Female: m.Female}
	cand := town.CandidateFromMember(m)
	cand.Slot = slot
	s.Roster = append(append([]town.Candidate(nil), s.Roster...), cand)
	if s.SaveGame != nil {
		if err := s.SaveGame(); err != nil {
			r.Slots[slot] = previous
			s.Roster = oldRoster
			return err
		}
	}
	return nil
}
