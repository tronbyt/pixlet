package qrcode_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

const qrCodeSource = `
load("qrcode.star", "qrcode")
load("assert.star", "assert")

url = "https://tidbyt.com?utm_source=pixlet_example"
code = qrcode.generate(
    url = url,
    size = "large",
    color = "#fff",
    background = "#000",
)

def main():
	return []
`

func TestQRCode(t *testing.T) {
	app, err := runtime.NewApplet("test.star", []byte(qrCodeSource), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
