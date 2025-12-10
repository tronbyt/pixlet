//go:build nativewebp

package render

import (
	"bytes"
	"fmt"

	"github.com/gen2brain/webp"
)

func (p *Image) InitFromWebP(data []byte) error {
	webpImage, err := webp.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("decoding image data: %w", err)
	}

	p.Delay = 0
	if len(webpImage.Image) > 0 {
		// The delays in webpImage.Delay are in milliseconds
		p.Delay = webpImage.Delay[0]
	}

	// append all frames at once
	p.imgs = append(p.imgs, webpImage.Image...)

	return nil
}
