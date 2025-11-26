package hmac_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

const hmacSource = `
load("hmac.star", "hmac")
load("assert.star", "assert")

# Assert.

assert.eq(hmac.md5("secret", "helloworld"), "8bd4df4530c3c2cafabf6986740e44bd")
assert.eq(hmac.sha1("secret", "helloworld"), "e92eb69939a8b8c9843a75296714af611c73fb53")
assert.eq(hmac.sha256("secret", "helloworld"), "7a7c2bf41973489be3b318ad2f16c75fc875c340deecb12a3f79b28bb7135c97")

def main():
	return []
`

func TestHmac(t *testing.T) {
	app, err := runtime.NewApplet("hmac_test.star", []byte(hmacSource), runtime.WithTests(t))
	assert.NoError(t, err)
	assert.NotNil(t, app)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
