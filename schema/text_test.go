package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

func TestText(t *testing.T) {
	const source = `
load("assert.star", "assert")
load("schema.star", "schema")

s = schema.Text(
	id = "screen_name",
	name = "Screen Name",
	desc = "A text entry for your screen name.",
	icon = "user",
	default = "foo",
	secret = True,
)

assert.eq(s.id, "screen_name")
assert.eq(s.name, "Screen Name")
assert.eq(s.desc, "A text entry for your screen name.")
assert.eq(s.icon, "user")
assert.eq(s.default, "foo")
assert.eq(s.secret, True)

def main():
	return []
`

	app, err := runtime.NewApplet("text.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
