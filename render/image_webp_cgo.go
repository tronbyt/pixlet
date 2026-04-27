//go:build !nativewebp

package render

import (
	"fmt"
	"image"

	"github.com/tronbyt/go-libwebp/webp"
)

func (p *Image) InitFromWebP(data []byte) error {
	decoder, err := webp.NewAnimationDecoder(data)
	if err != nil {
		return fmt.Errorf("creating animation decoder: %w", err)
	}

	img, err := decoder.Decode()
	if err != nil {
		return fmt.Errorf("decoding image data: %w", err)
	}

	p.Delay = img.Timestamp[0]
	p.imgs = make([]image.Image, 0, len(img.Image))

	for _, im := range img.Image {
		p.imgs = append(p.imgs, im)
	}

	return nil
}
