package ui

import (
	"image"
	imagedraw "image/draw"

	"github.com/hajimehoshi/ebiten/v2"
)

// Only lightweight handles survive eviction. Deallocate preserves subimages
// and their bindings while freeing the GPU texture and its restoration data.
const hdTextureBudget = 192 << 20

type hdCachedImage struct {
	source   image.Image
	load     func() (image.Image, error)
	image    *ebiten.Image
	bytes    int64
	used     uint64
	resident bool
}

type hdImageCache struct {
	entries map[*ebiten.Image]*hdCachedImage
	parents map[*ebiten.Image]*ebiten.Image
	frame   uint64
	bytes   int64
	err     error
}

func (h *HD) imageCache() *hdImageCache {
	if h.cache == nil {
		h.cache = &hdImageCache{entries: map[*ebiten.Image]*hdCachedImage{}, parents: map[*ebiten.Image]*ebiten.Image{}}
	}
	return h.cache
}

func (h *HD) newPaint(src image.Image) *ebiten.Image {
	b := src.Bounds()
	pic := ebiten.NewImage(b.Dx(), b.Dy())
	h.imageCache().entries[pic] = &hdCachedImage{source: src, image: pic, bytes: int64(b.Dx()) * int64(b.Dy()) * 4}
	return pic
}

func (h *HD) newPaintLoader(bounds image.Rectangle, load func() (image.Image, error)) *ebiten.Image {
	pic := ebiten.NewImage(bounds.Dx(), bounds.Dy())
	h.imageCache().entries[pic] = &hdCachedImage{load: load, image: pic, bytes: int64(bounds.Dx()) * int64(bounds.Dy()) * 4}
	return pic
}

func (h *HD) paintSubImage(src *ebiten.Image, rect image.Rectangle) *ebiten.Image {
	pic := src.SubImage(rect).(*ebiten.Image)
	c := h.imageCache()
	root := src
	if parent := c.parents[src]; parent != nil {
		root = parent
	}
	c.parents[pic] = root
	return pic
}

func (h *HD) paint(src *ebiten.Image) *ebiten.Image {
	c := h.imageCache()
	root := src
	if parent := c.parents[src]; parent != nil {
		root = parent
	}
	e := c.entries[root]
	if e == nil {
		return src
	}
	e.used = c.frame
	if e.resident || c.err != nil {
		return src
	}
	var raw image.Image
	var err error
	if e.load != nil {
		raw, err = e.load()
	} else {
		raw, err = readPaintImage(e.source)
	}
	if err != nil {
		c.err = err
		return src
	}
	rgba := image.NewRGBA(root.Bounds())
	imagedraw.Draw(rgba, rgba.Bounds(), raw, raw.Bounds().Min, imagedraw.Src)
	root.WritePixels(rgba.Pix)
	e.resident = true
	c.bytes += e.bytes
	return src
}

func readPaintImage(src image.Image) (image.Image, error) {
	if provider, ok := src.(interface{ ReadHDImage() (image.Image, error) }); ok {
		return provider.ReadHDImage()
	}
	return src, nil
}

// BeginHDFrame runs before rendering; entries used in the current frame are
// pinned until the next frame so queued draws never reference released pixels.
func (h *HD) BeginHDFrame() {
	c := h.imageCache()
	c.frame++
	c.err = nil
}

func (h *HD) EndHDFrame() error {
	c := h.imageCache()
	if p := h.equipment; p != nil && p.usedFrame != c.frame {
		h.releaseEquipmentMemory()
	}
	for c.bytes > hdTextureBudget {
		var oldest *hdCachedImage
		for _, e := range c.entries {
			if e.resident && e.used != c.frame && (oldest == nil || e.used < oldest.used) {
				oldest = e
			}
		}
		if oldest == nil {
			break
		}
		oldest.image.Deallocate()
		oldest.resident = false
		c.bytes -= oldest.bytes
	}
	return c.err
}

// ReleaseHDMemory drops every cached texture when switching to EGA. Handles
// and the validated index remain reusable; the next HD draw reloads its scene.
func (h *HD) ReleaseHDMemory() {
	c := h.imageCache()
	for _, e := range c.entries {
		if e.resident {
			e.image.Deallocate()
			e.resident = false
		}
	}
	c.bytes = 0
	for _, pic := range h.samples {
		pic.Dispose()
	}
	h.samples = nil
	h.sampleFrames = nil
	if m := h.motion; m != nil {
		m.cpu = map[string]*image.NRGBA{}
		m.clouds = map[int]*image.NRGBA{}
		if m.weather != nil {
			m.weather.Dispose()
			m.weather = nil
		}
		if m.body != nil {
			m.body.Dispose()
			m.body = nil
		}
		m.weatherValid, m.bodyValid = false, false
	}
	h.releaseEquipmentMemory()
}

func (h *HD) releaseEquipmentMemory() {
	if p := h.equipment; p != nil {
		for _, pic := range p.Warped {
			pic.Dispose()
		}
		for _, pic := range p.BodyShapes {
			pic.Dispose()
		}
		p.Warped = map[string]*ebiten.Image{}
		p.BodyShapes = map[string]*ebiten.Image{}
		p.Base.Deallocate()
		p.Figure.Deallocate()
	}
}
