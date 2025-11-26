package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

func TestPhotoSelect(t *testing.T) {
	const source = `
load("schema.star", "schema")
load("assert.star", "assert")

t = schema.PhotoSelect(
	id = "photo",
	name = "Add Photo",
	desc = "A photo.",
	icon = "gear",
)

assert.eq(t.id, "photo")
assert.eq(t.name, "Add Photo")
assert.eq(t.desc, "A photo.")
assert.eq(t.icon, "gear")

def main():
	return []
`

	app, err := runtime.NewApplet("photo_select.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
