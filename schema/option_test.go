package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

func TestOption(t *testing.T) {
	const source = `
load("schema.star", "schema")
load("assert.star", "assert")

s = schema.Option(
	display = "Green",
	value = "#00FF00",
)

assert.eq(s.display, "Green")
assert.eq(s.value, "#00FF00")

def main():
	return []
`

	app, err := runtime.NewApplet("option.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
