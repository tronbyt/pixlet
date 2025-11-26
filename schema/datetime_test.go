package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

func TestDateTime(t *testing.T) {
	const source = `
load("schema.star", "schema")
load("assert.star", "assert")

t = schema.DateTime(
	id = "event_name",
	name = "Event Name",
	desc = "The time of the event.",
	icon = "gear",
)

assert.eq(t.id, "event_name")
assert.eq(t.name, "Event Name")
assert.eq(t.desc, "The time of the event.")
assert.eq(t.icon, "gear")

def main():
	return []
`

	app, err := runtime.NewApplet("date_time.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
