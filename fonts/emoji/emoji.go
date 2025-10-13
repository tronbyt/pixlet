package emoji

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"sync"
)

//go:embed sprites.png
var spritesPNG []byte

var (
	sheetImg *image.NRGBA
	mu       sync.Mutex
)

func Sheet() (*image.NRGBA, error) {
	mu.Lock()
	defer mu.Unlock()

	if sheetImg == nil {
		img, err := png.Decode(bytes.NewReader(spritesPNG))
		if err != nil {
			return nil, err
		}
		if nrgba, ok := img.(*image.NRGBA); ok {
			sheetImg = nrgba
		} else {
			rb := image.NewNRGBA(img.Bounds())
			draw.Draw(rb, rb.Bounds(), img, image.Point{}, draw.Src)
			sheetImg = rb
		}
	}

	return sheetImg, nil
}

func Get(s string) (*image.NRGBA, error) {
	// Check if the emoji exists in our index
	point, exists := Index[s]
	if !exists {
		return nil, fmt.Errorf("emoji %q not found in emoji index", s)
	}

	// Get the emoji sprite sheet
	sheet, err := Sheet()
	if err != nil {
		return nil, fmt.Errorf("failed to load emoji sheet: %w", err)
	}

	// Extract the emoji from the sprite sheet
	srcRect := image.Rect(
		point.X*CellW, point.Y*CellH,
		(point.X+1)*CellW, (point.Y+1)*CellH,
	)

	// Create source image for this emoji
	srcImg := image.NewNRGBA(image.Rect(0, 0, CellW, CellH))
	draw.Draw(srcImg, srcImg.Bounds(), sheet, srcRect.Min, draw.Src)
	return srcImg, nil
}
