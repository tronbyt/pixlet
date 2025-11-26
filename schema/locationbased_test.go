package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

func TestLocationBased(t *testing.T) {
	const source = `
load("encoding/json.star", "json")
load("schema.star", "schema")
load("assert.star", "assert")

DEFAULT_LOCATION = """
{
	"lat": "40.6781784",
	"lng": "-73.9441579",
	"description": "Brooklyn, NY, USA",
	"locality": "Brooklyn",
	"place_id": "ChIJCSF8lBZEwokRhngABHRcdoI",
	"timezone": "America/New_York"
}
"""

def get_stations(location):
    loc = json.decode(location)
    lat, lng = float(loc["lat"]), float(loc["lng"])

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

t = schema.LocationBased(
    id = "station",
    name = "Train Station",
    desc = "A list of train stations based on a location.",
    icon = "train",
    handler = get_stations,
)

assert.eq(t.id, "station")
assert.eq(t.name, "Train Station")
assert.eq(t.desc, "A list of train stations based on a location.")
assert.eq(t.icon, "train")
assert.eq(t.handler(DEFAULT_LOCATION)[0].display, "Grand Central")

def main():
    return []

`

	app, err := runtime.NewApplet("location_based.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
