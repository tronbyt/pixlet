package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

func TestHandler(t *testing.T) {
	const source = `
load("schema.star", "schema")
load("assert.star", "assert")

def foobar(param):
    return "derp"

h = schema.Handler(
    handler = foobar,
    type = schema.HandlerType.String,
)

assert.eq(h.handler, foobar)
assert.eq(h.type, schema.HandlerType.String)

def main():
	return []
`

	app, err := runtime.NewApplet("handler.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}

func TestHandlerBadParams(t *testing.T) {
	// Handler is a string
	app, err := runtime.NewApplet("text.star", []byte(`
load("schema.star", "schema")

def foobar(param):
    return "derp"

h = schema.Handler(
    handler = "foobar",
    type = schema.HandlerType.String,
)

def main():
	return []
`), runtime.WithTests(t))
	assert.Error(t, err)
	assert.Nil(t, app)

	// Type is not valid
	app, err = runtime.NewApplet("text.star", []byte(`
load("schema.star", "schema")

def foobar(param):
    return "derp"

h = schema.Handler(
    handler = foobar,
    type = 42,
)

def main():
	return []
`), runtime.WithTests(t))
	assert.Error(t, err)
	assert.Nil(t, app)
}
