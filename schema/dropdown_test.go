package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

func TestDropdown(t *testing.T) {
	const source = `
load("schema.star", "schema")
load("assert.star", "assert")

options = [
	schema.Option(
		display = "Green",
		value = "#00FF00",
	),
	schema.Option(
		display = "Red",
		value = "#FF0000",
	),
]
	
s = schema.Dropdown(
	id = "colors",
	name = "Text Color",
	desc = "The color of text to be displayed.", 
	icon = "brush",
	default = options[0].value,
	options = options,
)

assert.eq(s.id, "colors")
assert.eq(s.name, "Text Color")
assert.eq(s.desc, "The color of text to be displayed.")
assert.eq(s.icon, "brush")
assert.eq(s.default, "#00FF00")

assert.eq(s.options[0].display, "Green")
assert.eq(s.options[0].value, "#00FF00")

assert.eq(s.options[1].display, "Red")
assert.eq(s.options[1].value, "#FF0000")

def main():
	return []
`

	app, err := runtime.NewApplet("dropdown.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
