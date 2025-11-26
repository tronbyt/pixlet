package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

const typeaheadSource = `
load("assert.star", "assert")
load("schema.star", "schema")

def search(pattern):
    return [
        schema.Option(
            display = "Grand Central",
            value = "abc123",
        ),
        schema.Option(
            display = "Penn Station",
            value = "xyz123",
        ),
    ]

t = schema.Typeahead(
    id = "search",
    name = "Search",
    desc = "A list of items that match search.",
    icon = "gear",
    handler = search,
)

assert.eq(t.id, "search")
assert.eq(t.name, "Search")
assert.eq(t.desc, "A list of items that match search.")
assert.eq(t.icon, "gear")
assert.eq(t.handler("")[0].display, "Grand Central")

def main():
    return []

`

func TestTypeahead(t *testing.T) {
	app, err := runtime.NewApplet("typeahead.star", []byte(typeaheadSource), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
