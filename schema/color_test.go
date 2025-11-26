package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

func TestColorSuccess(t *testing.T) {
	const source = `
load("schema.star", "schema")
load("assert.star", "assert")

# no palette, 3 char default
s1 = schema.Color(
    id = "colors",
    name = "Colors",
    desc = "The color to display",
    icon = "brush",
    default = "#fff",
)

assert.eq(s1.id, "colors")
assert.eq(s1.name, "Colors")
assert.eq(s1.desc, "The color to display")
assert.eq(s1.icon, "brush")
assert.eq(s1.default, "#fff")

# with palette
s2 = schema.Color(
    id = "colors",
    name = "Colors",
    desc = "The color to display",
    icon = "brush",
    default = "123456",
    palette = ["#f0f", "#aabbcd", "103", "323334"],
)

assert.eq(s2.id, "colors")
assert.eq(s2.name, "Colors")
assert.eq(s2.desc, "The color to display")
assert.eq(s2.icon, "brush")
assert.eq(s2.default, "#123456")
print(s2.palette)
assert.eq(len(s2.palette), 4)
assert.eq(s2.palette[0], "#f0f")
assert.eq(s2.palette[1], "#aabbcd")
assert.eq(s2.palette[2], "#103")
assert.eq(s2.palette[3], "#323334")

def main():
    return []
`
	app, err := runtime.NewApplet("colors.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}

func TestColorMalformedColors(t *testing.T) {
	const source = `
load("schema.star", "schema")
load("encoding/json.star", "json")
load("assert.star", "assert")

def main(config):
    s = schema.Color(
        id = "colors",
        name = "Colors",
        desc = "The color to display",
        icon = "brush",
        default = config["default"],
        palette = json.decode(config["palette"]),
    )

    assert.eq(s.id, "colors")
    assert.eq(s.name, "Colors")
    assert.eq(s.desc, "The color to display")
    assert.eq(s.icon, "brush")

    return []
`
	app, err := runtime.NewApplet("colors.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	// Well formed input -> success
	screens, err := app.RunWithConfig(t.Context(), map[string]string{"default": "#ffaa77", "palette": "[]"})
	assert.NoError(t, err)
	assert.NotNil(t, screens)

	// Bad default
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "#nothex", "palette": "[]"})
	assert.Error(t, err)
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "", "palette": "[]"})
	assert.Error(t, err)
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "0", "palette": "[]"})
	assert.Error(t, err)
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "01", "palette": "[]"})
	assert.Error(t, err)
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "#01", "palette": "[]"})
	assert.Error(t, err)
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "0123", "palette": "[]"})
	assert.Error(t, err)
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "#0123", "palette": "[]"})
	assert.Error(t, err)
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "01234", "palette": "[]"})
	assert.Error(t, err)
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "#01234", "palette": "[]"})
	assert.Error(t, err)
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "0123456", "palette": "[]"})
	assert.Error(t, err)
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "#0123456", "palette": "[]"})
	assert.Error(t, err)

	// Bad palette
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "#ffaa77", "palette": `["nothex"]`})
	assert.Error(t, err)
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "#ffaa77", "palette": `["fff", "ffaabb", "0"]`})
	assert.Error(t, err)
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "#ffaa77", "palette": `["fff", "ffaabb", "#0f"]`})
	assert.Error(t, err)
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "#ffaa77", "palette": `["fff", "ffaabb", "0123"]`})
	assert.Error(t, err)
	_, err = app.RunWithConfig(t.Context(), map[string]string{"default": "#ffaa77", "palette": `["fff", "ffaabb", "0123456"]`})
	assert.Error(t, err)
}
