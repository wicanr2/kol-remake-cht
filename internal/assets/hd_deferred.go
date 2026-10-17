package assets

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

// HDLoadOption keeps the exhaustive offline validators available while the game
// opens only an immutable index. PNG bytes and pixels are checked when needed.
type HDLoadOption uint8

const DeferHDPixels HDLoadOption = 1

func deferHDPixels(options []HDLoadOption) bool {
	return len(options) != 0 && options[0] == DeferHDPixels
}

type deferredHDImage struct {
	path string
	info HDImageInfo
	rect image.Rectangle
}

func openHDImage(dir, name string, info HDImageInfo) (image.Image, error) {
	if filepath.Base(name) != name || !validHDHash(info.SHA256) || info.Width <= 0 || info.Height <= 0 || info.Width > 8192 || info.Height > 8192 {
		return nil, fmt.Errorf("HD圖片索引無效：%s", name)
	}
	path := filepath.Join(dir, name)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	cfg, err := png.DecodeConfig(f)
	f.Close()
	if err != nil || cfg.Width != info.Width || cfg.Height != info.Height {
		return nil, fmt.Errorf("HD圖片尺寸不符：%s", name)
	}
	return &deferredHDImage{path: path, info: info, rect: image.Rect(0, 0, info.Width, info.Height)}, nil
}

func (p *deferredHDImage) Bounds() image.Rectangle { return p.rect }
func (p *deferredHDImage) ColorModel() color.Model { return color.NRGBAModel }

// Runtime rendering uses ReadHDImage, so I/O failures propagate as errors.
func (p *deferredHDImage) At(x, y int) color.Color {
	img, err := p.read()
	if err != nil {
		panic(err)
	}
	return img.At(x, y)
}
func (p *deferredHDImage) SubImage(r image.Rectangle) image.Image {
	copy := *p
	copy.rect = r.Intersect(p.rect)
	return &copy
}
func (p *deferredHDImage) read() (image.Image, error) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		return nil, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != p.info.SHA256 {
		return nil, fmt.Errorf("HD圖片雜湊不符：%s", filepath.Base(p.path))
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if img.Bounds() != image.Rect(0, 0, p.info.Width, p.info.Height) {
		return nil, fmt.Errorf("HD圖片實際尺寸不符：%s", filepath.Base(p.path))
	}
	return img.(interface {
		SubImage(image.Rectangle) image.Image
	}).SubImage(p.rect), nil
}

// ReadHDImage satisfies the renderer's image-provider interface without making
// the UI depend on the filesystem asset package.
func (p *deferredHDImage) ReadHDImage() (image.Image, error) { return p.read() }

// ReadHDImage does not retain decoded pixels. The renderer owns its bounded GPU
// cache; eviction reloads and verifies the original PNG rather than a stale copy.
func ReadHDImage(img image.Image) (image.Image, error) {
	if p, ok := img.(*deferredHDImage); ok {
		return p.ReadHDImage()
	}
	return img, nil
}
