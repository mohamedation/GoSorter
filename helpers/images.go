// Package helpers - images
package helpers

import (
	"image/png"
	"os"

	"github.com/mohamedation/GoSorter/model"
)

// HasTransparency - maybe also set a size or resolution limit?
func HasTransparency(filePath string, cfg model.Config, logger Logger) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if logger != nil {
			logger.Log(cfg, Error, "Failed to open image file: "+err.Error())
		}
		return false, err
	}
	defer func() {
		if err := file.Close(); err != nil && logger != nil {
			logger.Log(cfg, Error, "error closing image file: "+err.Error())
		}
	}()

	img, err := png.Decode(file)
	if err != nil {
		if logger != nil {
			logger.Log(cfg, Error, "Failed to decode PNG: "+err.Error())
		}
		return false, err
	}

	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	totalPixels := width * height
	if totalPixels == 0 {
		return false, nil
	}

	transparentPixels := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a < 0xffff {
				transparentPixels++
			}
		}
	}
	// transparency is significant?
	return float64(transparentPixels)/float64(totalPixels) > 0.5, nil
}
