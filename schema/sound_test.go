package schema_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tronbyt/pixlet/runtime"
)

func TestSound(t *testing.T) {
	const source = `
load("schema.star", "schema")
load("sound.mp3", "file")
load("assert.star", "assert")

s = schema.Sound(
	id = "sound1",
	title = "Sneezing Elephant",
	file = file,
)

assert.eq(s.id, "sound1")
assert.eq(s.title, "Sneezing Elephant")
assert.eq(s.file, file)
assert.eq(s.file.readall(), "sound data")

def main():
	return []
`

	vfs := fstest.MapFS{
		"sound.mp3":  &fstest.MapFile{Data: []byte("sound data")},
		"sound.star": &fstest.MapFile{Data: []byte(source)},
	}
	app, err := runtime.NewAppletFromFS(t.Context(), vfs, "sound", runtime.WithTests(t))
	require.NoError(t, err)

	screens, err := app.Run(t.Context())
	require.NoError(t, err)
	assert.NotNil(t, screens)
}
