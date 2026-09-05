package town

// 這一支只給「產生腳本」用：它算的是「要按哪幾個方向鍵才走得到那一格」，
// 讓對拍的按鍵序列由地圖本身算出來，而不是人工試出來再抄進規格檔。
// 玩家路徑本身仍然只按 `Walk`，這裡沒有捷徑也沒有座標注入。

// routeStep 是四個方向鍵。斜走鍵走得到的地方直走也走得到，只是多幾步，
// 所以不放進搜尋——序列短一點不值得讓腳本多四個符號。
var routeStep = []struct {
	Token  string
	DX, DY int
}{
	{"L", -1, 0}, {"R", 1, 0}, {"U", 0, -1}, {"D", 0, 1},
}

// routeNode 是走動的完整狀態。**光記隊伍座標不夠**：同一格可以由「鏡頭在這裡、
// 標記偏移 0」或「鏡頭在前一格、偏移 1」兩種狀態組成，而下一步能不能走要看
// 是哪一種（`step` 的死區規則）。
type routeNode struct{ cx, cy, mc, mr int }

func (o *Outdoor) node() routeNode {
	return routeNode{o.Camera.X, o.Camera.Y, o.Marker.Col, o.Marker.Row}
}

func (o *Outdoor) restore(n routeNode) {
	o.Camera.X, o.Camera.Y = n.cx, n.cy
	o.Marker.Col, o.Marker.Row = n.mc, n.mr
}

// RouteTo 是從現在的位置走到 (x, y) 的最短方向鍵序列；走不到就是 nil。
// 走的是 `Walk` 本人，所以地形、鏡頭死區與邊界夾限全部照玩家路徑那一套。
func (o *Outdoor) RouteTo(x, y int) []string {
	start := o.node()
	defer o.restore(start)

	type entry struct {
		node routeNode
		path []string
	}
	seen := map[routeNode]bool{start: true}
	queue := []entry{{start, nil}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, s := range routeStep {
			o.restore(cur.node)
			if !o.Walk(s.DX, s.DY) {
				continue
			}
			next := o.node()
			if seen[next] {
				continue
			}
			seen[next] = true
			path := append(append([]string(nil), cur.path...), s.Token)
			if px, py := o.PartyPosition(); px == x && py == y {
				return path
			}
			// 踩到門那一格畫面就換成門口圖示列，接下來的方向鍵移動的是白框
			// 而不是隊伍——那條路走不下去，當成死路。
			if o.AtDoorCell {
				continue
			}
			queue = append(queue, entry{next, path})
		}
	}
	return nil
}
