package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

func TestLocation(t *testing.T) {
	const source = `
load("schema.star", "schema")
load("assert.star", "assert")

s = schema.Location(
	id = "location",
	name = "Location",
	desc = "Location for which to display time.",
	icon = "locationDot",
)

assert.eq(s.id, "location")
assert.eq(s.name, "Location")
assert.eq(s.desc, "Location for which to display time.")
assert.eq(s.icon, "locationDot")

def main():
	return []
`

	app, err := runtime.NewApplet("location.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
