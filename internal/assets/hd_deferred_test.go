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
	"testing"
)

func TestDeferredHDChecksPixelsAtUseAndReload(t *testing.T) {
	dir := t.TempDir()
	pic := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	pic.SetNRGBA(2, 1, color.NRGBA{31, 83, 171, 128})
	var buf bytes.Buffer
	if err := png.Encode(&buf, pic); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "paint.png")
	if err := os.WriteFile(path, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	info := HDImageInfo{SHA256: fmt.Sprintf("%x", sha256.Sum256(buf.Bytes())), Width: 3, Height: 2}
	lazy, err := openHDImage(dir, "paint.png", info)
	if err != nil {
		t.Fatal(err)
	}
	crop := lazy.(interface {
		SubImage(image.Rectangle) image.Image
	}).SubImage(image.Rect(1, 0, 3, 2))
	first, err := ReadHDImage(crop)
	if err != nil || first.Bounds() != crop.Bounds() {
		t.Fatalf("crop: %v %v", first, err)
	}
	if got := color.NRGBAModel.Convert(first.At(2, 1)); got != pic.NRGBAAt(2, 1) {
		t.Fatalf("pixel: %v", got)
	}
	if err := os.WriteFile(path, []byte("broken after indexing"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadHDImage(crop); err == nil {
		t.Fatal("cached stale pixels after replacement")
	}
	if err := os.WriteFile(path, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadHDImage(crop); err != nil {
		t.Fatal(err)
	}
	info.Width++
	if _, err := openHDImage(dir, "paint.png", info); err == nil {
		t.Fatal("wrong dimensions accepted")
	}
	if _, err := openHDImage(dir, "../paint.png", info); err == nil {
		t.Fatal("escaping path accepted")
	}
}
