package battle

// setScratch保存原版word_39A54的覆寫順序（SPEC-011／RE-191至194）。
// 世界成本的表外讀取另走規格審查；這裡不依勝負推定成本。
func (s *State) setScratch(value int) {
	s.ReturnScratch = uint16(value)
}

func (s *State) trackExecStart(i int) {
	a := s.Actors[i]
	s.setScratch(a.Quickness * a.FatigueRatio() / 200)
	s.setScratch(TimeCost(a.Command, a.Style))
}

func (s *State) trackMain(i int) {
	s.setScratch(0)
	a := s.Actors[i]
	switch s.LastCommand {
	case CmdWeapon, CmdUnarmed, CmdCast, CmdMove, CmdMountedMove:
		s.setScratch(1)
	case CmdMissile:
		if a.Flag1D != 0 {
			s.setScratch(1)
		}
	}
}

func (s *State) trackRoundEdge() {
	s.setScratch(0)
	carrying := s.FleeOffer().Carrying
	for i := 0; i < carrying; i++ {
		s.setScratch(i + 1)
	}
}

func (s *State) trackLootClear() {
	s.setScratch(0)
	for i := 1; i <= 16; i++ {
		s.setScratch(i)
	}
}
