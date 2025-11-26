package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

func TestStarlarkSchema(t *testing.T) {
	const source = `
load("schema.star", "schema")
load("assert.star", "assert")

s = schema.Schema(
	version = "1",
	fields = [
		schema.Toggle(
			id = "display_weather",
			name = "Display Weather",
			desc = "A toggle to determine if the weather should be displayed.",
			icon = "cloud",
		),
	],
)

assert.eq(s.version, "1")
assert.eq(s.fields[0].name, "Display Weather")

def main():
	return []
`

	app, err := runtime.NewApplet("starlark.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
func TestSchemaModuleLoads(t *testing.T) {
	const source = `
load("schema.star", "schema")

def main():
	return []
`

	app, err := runtime.NewApplet("source.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
