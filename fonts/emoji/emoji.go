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
