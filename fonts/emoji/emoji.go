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
	sheetOnce sync.Once
	sheetImg  *image.RGBA
)

func Sheet() *image.RGBA {
	sheetOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(spritesPNG))
		if err != nil {
			return
		}
		if rgba, ok := img.(*image.RGBA); ok {
			sheetImg = rgba
		} else {
			rb := image.NewRGBA(img.Bounds())
			draw.Draw(rb, rb.Bounds(), img, image.Point{}, draw.Src)
			sheetImg = rb
		}
	})
	return sheetImg
}
