package ui

import (
	"errors"
	"image"
	"reflect"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// 戶外與野外的隊員格一律是藍框，但不能順手改掉角色狀態裡的發言者。
func TestIdlePartyKeepsSpeakerInState(t *testing.T) {
	party := []town.Member{{Name: "A", Speaker: true}, {Name: "B"}}
	out := idleParty(party)
	for i, m := range out {
		if m.Speaker {
			t.Fatalf("第 %d 位仍畫成發言者", i)
		}
	}
	if !party[0].Speaker {
		t.Fatal("idleParty 改到了原本的隊伍")
	}
}

type outdoorAssetKey struct {
	archive string
	index   int
}

type outdoorProbeSprites struct {
	image *ebiten.Image
	seen  []outdoorAssetKey
	fail  *outdoorAssetKey
	err   error
}

func (s *outdoorProbeSprites) Get(archive string, index int) (*ebiten.Image, error) {
	key := outdoorAssetKey{archive, index}
	s.seen = append(s.seen, key)
	if s.fail != nil && key == *s.fail {
		return nil, s.err
	}
	return s.image, nil
}

// 分別走正常地圖、建築門口及城門；邊界即使同時有門口旗標也應選林地圖。
// 圖像缺失必須回傳原錯誤，繪製不得改寫戶外狀態。
func TestOutdoorCJKMapStatesAndMissingAssets(t *testing.T) {
	grid, err := town.LoadGrid(image.NewGray(image.Rect(0, 0, town.GridSide, town.GridSide)))
	if err != nil {
		t.Fatal(err)
	}
	canvas := ebiten.NewImage(960, 600)
	defer canvas.Deallocate()
	pic := ebiten.NewImage(512, 512)
	defer pic.Deallocate()
	for _, tc := range []struct {
		name       string
		edge, door bool
		want       outdoorAssetKey
	}{
		{"地圖", false, false, outdoorAssetKey{"TOWN7", town.TileAtlas.Index}},
		{"建築門口", false, true, outdoorAssetKey{"TWN7PICS", 19}},
		{"城門", true, false, outdoorAssetKey{town.EdgePicture.Archive, town.EdgePicture.Index}},
		{"邊界優先", true, true, outdoorAssetKey{town.EdgePicture.Archive, town.EdgePicture.Index}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := &town.Outdoor{Town: 7, AtEdge: tc.edge, AtDoorCell: tc.door}
			out.Camera.X, out.Camera.Y = 25, 7
			out.Bounds.MaxX, out.Bounds.MaxY = 63, 63
			out.SetGrid(grid)
			out.SetBuildings(&town.TownBuildings{Doors: []town.Door{{X: 25, Y: 7, Picture: 19}}})
			before := *out
			for _, missing := range []bool{false, true} {
				s := &outdoorProbeSprites{image: pic, err: errors.New("測試素材缺失")}
				if missing {
					s.fail = &tc.want
				}
				err := OutdoorCJK(canvas, s, nil, nil, out)
				if missing && err != s.err || !missing && err != nil {
					t.Fatalf("素材缺失=%t：回傳 %v", missing, err)
				}
				found := false
				for _, key := range s.seen {
					found = found || key == tc.want
				}
				if !found {
					t.Fatalf("未取得預期素材 %+v，實際 %+v", tc.want, s.seen)
				}
				if !reflect.DeepEqual(before, *out) {
					t.Fatal("繪製改寫了戶外狀態")
				}
			}
		})
	}
}
