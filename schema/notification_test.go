package schema_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/tronbyt/pixlet/runtime"
)

func TestNotification(t *testing.T) {
	const source = `
load("assert.star", "assert")
load("schema.star", "schema")
load("sound.mp3", "file")

sounds = [
	schema.Sound(
		id = "ding",
		title = "Ding!",
		file = file,
	),

]

s = schema.Notification(
	id = "notification1",
	name = "New message",
	desc = "A new message has arrived",
	icon = "message",
	sounds = sounds,
	builder = lambda: None,
)

assert.eq(s.id, "notification1")
assert.eq(s.name, "New message")
assert.eq(s.desc, "A new message has arrived")
assert.eq(s.icon, "message")

assert.eq(s.sounds[0].id, "ding")
assert.eq(s.sounds[0].title, "Ding!")
assert.eq(s.sounds[0].file, file)

def main():
	return []
`

	vfs := fstest.MapFS{
		"sound.mp3":         &fstest.MapFile{Data: []byte("sound data")},
		"notification.star": &fstest.MapFile{Data: []byte(source)},
	}
	app, err := runtime.NewAppletFromFS("sound", vfs, runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
