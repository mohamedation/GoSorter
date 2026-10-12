package helpers

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/mohamedation/GoSorter/model"
)

func writeTestPNG(t *testing.T, path string, transparent bool) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			if transparent {
				img.Set(x, y, color.RGBA{0, 0, 0, 0})
			} else {
				img.Set(x, y, color.RGBA{255, 0, 0, 255})
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create png: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
}

func TestHasTransparency_Transparent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "t.png")
	writeTestPNG(t, path, true)

	ok, err := HasTransparency(path, model.Config{Silent: true}, &CLILogger{})
	if err != nil {
		t.Fatalf("HasTransparency: %v", err)
	}
	if !ok {
		t.Fatal("expected transparent PNG")
	}
}

func TestHasTransparency_Opaque(t *testing.T) {
	path := filepath.Join(t.TempDir(), "o.png")
	writeTestPNG(t, path, false)

	ok, err := HasTransparency(path, model.Config{Silent: true}, nil)
	if err != nil {
		t.Fatalf("HasTransparency: %v", err)
	}
	if ok {
		t.Fatal("expected opaque PNG")
	}
}

func TestHasTransparency_Missing(t *testing.T) {
	_, err := HasTransparency(filepath.Join(t.TempDir(), "missing.png"), model.Config{Silent: true}, &CLILogger{})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestHasTransparency_InvalidPNG(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.png")
	if err := os.WriteFile(path, []byte("not a png"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := HasTransparency(path, model.Config{Silent: true}, &CLILogger{})
	if err == nil {
		t.Fatal("expected decode error")
	}
}
