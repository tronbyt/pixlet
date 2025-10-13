package emoji

import (
	"bytes"
	_ "embed"
	"image"
	"image/draw"
	"image/png"
	"sync"
)

//go:embed sprites.png
var spritesPNG []byte

var (
	sheetImg *image.RGBA
	mu       sync.Mutex
)

func Sheet() (*image.RGBA, error) {
	mu.Lock()
	defer mu.Unlock()

	if sheetImg == nil {
		img, err := png.Decode(bytes.NewReader(spritesPNG))
		if err != nil {
			return nil, err
		}
		if rgba, ok := img.(*image.RGBA); ok {
			sheetImg = rgba
		} else {
			rb := image.NewRGBA(img.Bounds())
			draw.Draw(rb, rb.Bounds(), img, image.Point{}, draw.Src)
			sheetImg = rb
		}
	}

	return sheetImg, nil
}
