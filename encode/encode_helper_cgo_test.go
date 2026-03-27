//go:build !nativewebp

package encode

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tronbyt/go-libwebp/webp"
)

func webpDelays(t *testing.T, webpData []byte) []int {
	decoder, err := webp.NewAnimationDecoder(webpData)
	require.NoError(t, err)
	img, err := decoder.Decode()
	require.NoError(t, err)
	delays := make([]int, 0, len(img.Timestamp))
	last := 0
	for _, ts := range img.Timestamp {
		d := ts - last
		last = ts
		delays = append(delays, d)
	}
	return delays
}
