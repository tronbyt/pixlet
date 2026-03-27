//go:build nativewebp

package encode

import (
	"bytes"
	"testing"

	"github.com/gen2brain/webp"
	"github.com/stretchr/testify/require"
)

func webpDelays(t *testing.T, webpData []byte) []int {
	img, err := webp.DecodeAll(bytes.NewReader(webpData))
	require.NoError(t, err)
	return img.Delay
}
