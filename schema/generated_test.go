package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

func TestGenerated(t *testing.T) {
	const source = `
load("schema.star", "schema")
load("assert.star", "assert")

def validate(value):
	assert.eq(value, value)

s = schema.Generated(
	id = "foo",
        source = "bar",
        handler = validate,
)

assert.eq(s.id, "foo")
assert.eq(s.source, "bar")
assert.eq(s.handler, validate)

def main():
	return []
`

	app, err := runtime.NewApplet("generated.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
