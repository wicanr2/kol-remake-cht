package town

// 從世界地圖進城（RE-069）。`sub_29CA3` 讀世界地圖的城鎮記錄，把鏡頭、四個夾限、
// 三個字母的代號與磁片編號寫進全域，標記偏移歸零，最後算出城鎮編號與入城點：
// 記錄 i 進的是 TOWN{i+1}；i ≥ 6 全部進 TOWN7，用入城點 i−6 區分。

// TownEntry 是世界地圖城鎮記錄裡跟進城有關的欄位。
type TownEntry struct {
	Code                   string
	Town                   int
	Camera                 Point
	MinX, MaxX, MinY, MaxY int
	EntryPoint             int
}

// EnterFromWorld 把城鎮狀態重設成剛進城的樣子。呼叫者負責提供該城鎮的
// 格子陣列與建築資料——這兩份都是逐城鎮的檔案。
func (s *Session) EnterFromWorld(e TownEntry, info *TownBuildings, grid *Grid) {
	o := s.Outdoor
	o.Town = e.Town
	s.TownCode = e.Code
	o.Camera.X, o.Camera.Y = e.Camera.X, e.Camera.Y
	o.Bounds.MinX, o.Bounds.MaxX = e.MinX, e.MaxX
	o.Bounds.MinY, o.Bounds.MaxY = e.MinY, e.MaxY
	o.Marker.Row, o.Marker.Col = 0, 0
	o.EntryPoint = e.EntryPoint
	o.AtEdge, o.AtDoorCell = false, false
	o.Icons, o.Selected, o.DoorCursor = nil, 0, 0
	if grid != nil {
		o.SetGrid(grid)
	}
	if info != nil {
		o.SetBuildings(info)
		s.Buildings = info
	}
	s.Mode = ModeMap
	s.Cursor = 0
	s.Inside = nil
	s.pending = nil
	s.leaving = false
	s.told = [TopicSlots]bool{}
	s.greeted = false
	s.LeftTown = false
	s.WorldEntered = false
}
