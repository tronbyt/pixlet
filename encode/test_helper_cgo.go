//go:build !nativewebp

package encode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/go-libwebp/webp"
)

func webpDelays(t *testing.T, webpData []byte) []int {
	decoder, err := webp.NewAnimationDecoder(webpData)
	assert.NoError(t, err)
	img, err := decoder.Decode()
	assert.NoError(t, err)
	delays := []int{}
	last := 0
	for _, ts := range img.Timestamp {
		d := ts - last
		last = ts
		delays = append(delays, d)
	}
	return delays
}
