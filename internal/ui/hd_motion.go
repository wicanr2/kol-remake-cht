package ui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/world"
	"image"
	"image/color"
	imagedraw "image/draw"
	"math"
	"strings"
)

type HDMotionImage struct {
	Name            string
	Source          HDPageSource
	Size            image.Point
	Original, Paint image.Image
}

func motionAbs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

type hdMotion struct {
	sourceSizes                map[string]image.Point
	tiles                      map[[2]int]*ebiten.Image
	helmets                    [5]*ebiten.Image
	cards                      map[string]int
	images                     map[string]*ebiten.Image
	cpu                        map[string]*image.NRGBA
	originals                  map[string]image.Image
	byRGB                      map[string]string
	masks                      map[int]*image.Alpha
	clouds                     map[int]*image.NRGBA
	weather                    *ebiten.Image
	weatherGroup, weatherPhase int
	weatherValid               bool
	body                       *ebiten.Image
	bodyKey                    battle.BodyPanelData
	bodyValid                  bool
}

func (h *HD) AttachMotion(arts []HDMotionImage, masks map[int]image.Image) error {
	if h.pages == nil || len(arts) != 65 || len(masks) != 4 {
		return fmt.Errorf("HD動畫需要完整65圖與四張資料遮罩")
	}
	seen := map[string]bool{}
	sources := append([]HDPageSource{}, h.pages.sources...)
	for _, a := range arts {
		if a.Name == "" || seen[a.Name] || a.Paint == nil || a.Original == nil || a.Original.Bounds().Size() != a.Size || len(a.Source.RGB) != 64 || a.Paint.Bounds().Dx() > 768 || a.Paint.Bounds().Dy() > 768 {
			return fmt.Errorf("HD動畫來源或尺寸不符")
		}
		seen[a.Name] = true
		exists := false
		for _, s := range sources {
			if s.Archive == a.Source.Archive && s.Index == a.Source.Index {
				if s.RGB != a.Source.RGB {
					return fmt.Errorf("HD動畫來源衝突")
				}
				exists = true
			}
		}
		if !exists {
			sources = append(sources, a.Source)
		}
	}
	for _, g := range []int{0, 1, 2, 5} {
		m := masks[g]
		if m == nil || m.Bounds().Size() != image.Pt(64, 62) {
			return fmt.Errorf("HD雲層來源遮罩不符")
		}
	}
	next := &hdMotion{images: map[string]*ebiten.Image{}, cpu: map[string]*image.NRGBA{}, originals: map[string]image.Image{}, byRGB: map[string]string{}, masks: map[int]*image.Alpha{}, clouds: map[int]*image.NRGBA{}}
	for _, a := range arts {
		b := a.Paint.Bounds()
		pic := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		imagedraw.Draw(pic, pic.Bounds(), a.Paint, b.Min, imagedraw.Src)
		next.cpu[a.Name] = pic
		next.images[a.Name] = ebiten.NewImageFromImage(a.Paint)
		next.originals[a.Name] = a.Original
		if next.sourceSizes == nil {
			next.sourceSizes = map[string]image.Point{}
		}
		next.sourceSizes[a.Source.RGB] = a.Size
		if !strings.HasPrefix(a.Name, "battle-") {
			next.byRGB[a.Source.RGB] = a.Name
		}
	}
	for _, g := range []int{0, 1, 2, 5} {
		src := masks[g]
		mask := image.NewAlpha(image.Rect(0, 0, 128, 31))
		for i := range mask.Pix {
			r, gg, b, _ := src.At(src.Bounds().Min.X+i%64, src.Bounds().Min.Y+i/64).RGBA()
			if r == 0 && gg == 0 && b == 0 {
				mask.Pix[i] = 255
			} else if r != 65535 || gg != 65535 || b != 65535 {
				return fmt.Errorf("HD雲層遮罩須為二值")
			}
		}
		next.masks[g] = mask
		texture, e := cloudMaterial(next.cpu[fmt.Sprintf("weather-cloud-%d", g)])
		if e != nil {
			return e
		}
		next.clouds[g] = texture
	}
	if err := next.prepareTitleMattes(); err != nil {
		return err
	}
	h.motion = next
	h.pages.sources = sources
	h.pages.validated = false
	return nil
}

// Black in an overlay PNG is carrier data. Extend only painted cloud color
// beneath the separately validated source stencil, keeping black out of the sky.
func cloudMaterial(src *image.NRGBA) (*image.NRGBA, error) {
	out := image.NewNRGBA(src.Bounds())
	copy(out.Pix, src.Pix)
	var sum [3]int64
	count := int64(0)
	for y := 0; y < src.Bounds().Dy(); y++ {
		for x := 0; x < src.Bounds().Dx(); x++ {
			c := src.NRGBAAt(x, y)
			if int(c.R)+int(c.G)+int(c.B) > 45 {
				sum[0] += int64(c.R)
				sum[1] += int64(c.G)
				sum[2] += int64(c.B)
				count++
			}
		}
	}
	if count == 0 {
		return nil, fmt.Errorf("HD雲材質沒有可見畫稿")
	}
	average := color.NRGBA{uint8(sum[0] / count), uint8(sum[1] / count), uint8(sum[2] / count), 255}
	for y := 0; y < src.Bounds().Dy(); y++ {
		var positions []int
		for x := 0; x < src.Bounds().Dx(); x++ {
			c := src.NRGBAAt(x, y)
			if int(c.R)+int(c.G)+int(c.B) > 45 {
				positions = append(positions, x)
			}
		}
		at := 0
		for x := 0; x < src.Bounds().Dx(); x++ {
			c := src.NRGBAAt(x, y)
			if int(c.R)+int(c.G)+int(c.B) > 45 {
				continue
			}
			fill := average
			if len(positions) > 0 {
				for at+1 < len(positions) && motionAbs(positions[at+1]-x) < motionAbs(positions[at]-x) {
					at++
				}
				fill = src.NRGBAAt(positions[at], y)
			}
			out.SetNRGBA(x, y, fill)
		}
	}
	return out, nil
}
func stencilAlpha(m *image.Alpha, x, y float64) uint8 {
	x0, y0 := int(math.Floor(x)), int(math.Floor(y))
	fx, fy := x-float64(x0), y-float64(y0)
	at := func(xx, yy int) float64 {
		xx = ((xx % 128) + 128) % 128
		if yy < 0 {
			yy = 0
		}
		if yy > 30 {
			yy = 30
		}
		return float64(m.AlphaAt(xx, yy).A)
	}
	v := (at(x0, y0)*(1-fx)+at(x0+1, y0)*fx)*(1-fy) + (at(x0, y0+1)*(1-fx)+at(x0+1, y0+1)*fx)*fy
	return uint8(math.Round(v))
}
func (m *hdMotion) weatherFrame(group, phase int) (*ebiten.Image, error) {
	if group < 0 || group > 5 {
		return nil, fmt.Errorf("無效HD天候組")
	}
	if group == 3 || group == 4 {
		return m.images[fmt.Sprintf("weather-%d-%d", group, ((phase%3)+3)%3)], nil
	}
	phase = ((phase % 128) + 128) % 128
	if m.weatherValid && m.weatherGroup == group && m.weatherPhase == phase {
		return m.weather, nil
	}
	base := m.cpu[fmt.Sprintf("weather-base-%d", group)]
	out := image.NewNRGBA(base.Bounds())
	copy(out.Pix, base.Pix)
	cloud := m.clouds[group]
	for y := 0; y < 186; y++ {
		for x := 0; x < 384; x++ {
			sx := ((x-phase*6)%768 + 768) % 768
			a := int(stencilAlpha(m.masks[group], float64(x)/6+1.0/12-.5-float64(phase), float64(y)/6+1.0/12-.5))
			b, c := out.NRGBAAt(x, y), cloud.NRGBAAt(sx, y)
			mix := func(u, v uint8) uint8 { return uint8((int(u)*(255-a) + int(v)*a + 127) / 255) }
			out.SetNRGBA(x, y, color.NRGBA{mix(b.R, c.R), mix(b.G, c.G), mix(b.B, c.B), 255})
		}
	}
	if m.weather == nil {
		m.weather = ebiten.NewImage(384, 372)
	}
	m.weather.WritePixels(out.Pix)
	m.weatherGroup, m.weatherPhase, m.weatherValid = group, phase, true
	return m.weather, nil
}
func drawHDWeather(dst *ebiten.Image, s *world.Screen) (bool, error) {
	n := activeHDUI
	if n == nil || n.dst != dst || n.h.motion == nil {
		return false, nil
	}
	pic, e := n.h.motion.weatherFrame(s.Group, s.Phase)
	if e != nil {
		return true, e
	}
	l := world.CJK
	n.h.drawHD(dst, pic, image.Rect(l.WeatherPicX*2, l.WeatherPicY*2, (l.WeatherPicX+64*l.Scale)*2, (l.WeatherPicY+62*l.Scale)*2))
	return true, nil
}

func drawHDMotionBody(dst *ebiten.Image, data battle.BodyPanelData, rects [6]battle.BodyRect, scale int) bool {
	n := activeHDUI
	if n == nil || n.dst != dst || n.h.motion == nil {
		return false
	}
	m := n.h.motion
	id := fmt.Sprintf("body-%d", 3+data.Kind)
	base := m.cpu[id]
	source := m.originals[id]
	if base == nil || source == nil {
		n.err = fmt.Errorf("HD傷勢人形缺件")
		return true
	}
	if !m.bodyValid || m.bodyKey != data {
		out := image.NewNRGBA(base.Bounds())
		copy(out.Pix, base.Pix)
		for y := 0; y < out.Bounds().Dy(); y++ {
			for x := 0; x < out.Bounds().Dx(); x++ {
				sx, sy := x/6, y/6
				r, g, b, _ := source.At(source.Bounds().Min.X+sx, source.Bounds().Min.Y+sy).RGBA()
				v := uint8(r >> 8)
				if r != g || r != b || (v != 85 && v != 170 && v != 255) {
					continue
				}
				inside := false
				for _, rect := range rects {
					if sx+bodyX >= rect.Left && sx+bodyX < rect.Right && sy+bodyY >= rect.Top && sy+bodyY < rect.Bottom {
						inside = true
						break
					}
				}
				if inside {
					c := out.NRGBAAt(x, y)
					if int(c.R)+int(c.G)+int(c.B) > 15 {
						c.R = uint8((int(c.R)*3 + 255) / 4)
						c.G = uint8(int(c.G) * 85 / 255)
						c.B = uint8(int(c.B) * 85 / 255)
						out.SetNRGBA(x, y, c)
					}
				}
			}
		}
		if m.body == nil {
			m.body = ebiten.NewImage(336, 420)
		}
		m.body.WritePixels(out.Pix)
		m.bodyKey, m.bodyValid = data, true
	}
	n.h.drawHD(dst, m.body, image.Rect(bodyX*scale*2, bodyY*scale*2, (bodyX+bodyWidth)*scale*2, (bodyY+bodyHeight)*scale*2))
	return true
}
func (h *HD) drawMotionTownMarker(dst *ebiten.Image, out *town.Outdoor) bool {
	if h.motion == nil || out.MarkerOverlay() == nil {
		return false
	}
	index := out.Marker.Base + out.Marker.Frame
	if index < 0 || index >= 5 {
		return false
	}
	atlas := h.motion.images["walking-helmets"]
	row, col := out.MarkerCell()
	l := town.CJKOutdoor
	x, y := l.ViewX*2+col*l.Tile*2, l.ViewY*2+row*l.Tile*2
	pic := h.motion.helmets[index]
	if pic == nil {
		pic = atlas.SubImage(image.Rect(index*128, 0, (index+1)*128, 128)).(*ebiten.Image)
		h.motion.helmets[index] = pic
	}
	h.drawHD(dst, pic, image.Rect(x, y, x+96, y+96))
	return true
}

// DrawNativePage is the same typed page draw at the original CJK coordinates.
func DrawNativePage(dst, src *ebiten.Image, x, y, scale int) bool {
	return drawHDUI(dst, src, x, y, scale)
}

func drawHDMotionFatigue(dst *ebiten.Image, boundary, scale int) bool {
	n := activeHDUI
	if n == nil || n.dst != dst || n.h.motion == nil {
		return false
	}
	k := float32(scale * 2)
	vector.DrawFilledRect(dst, 132*k, 84*k, 7*k, float32(boundary-84+1)*k, color.RGBA{255, 133, 125, 255}, false)
	vector.DrawFilledRect(dst, 132*k, float32(boundary)*k, 7*k, float32(157-boundary)*k, color.RGBA{98, 118, 178, 255}, false)
	return true
}
