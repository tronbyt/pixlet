package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

func TestToggle(t *testing.T) {
	const source = `
load("schema.star", "schema")
load("assert.star", "assert")

t = schema.Toggle(
	id = "display_weather",
	name = "Display Weather",
	desc = "A toggle to determine if the weather should be displayed.",
	icon = "cloud",
	default = True,
)

assert.eq(t.id, "display_weather")
assert.eq(t.name, "Display Weather")
assert.eq(t.desc, "A toggle to determine if the weather should be displayed.")
assert.eq(t.icon, "cloud")
assert.eq(t.default, True)

def main():
	return []
`

	app, err := runtime.NewApplet("toggle.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
