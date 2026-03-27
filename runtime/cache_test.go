package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheGetAndSet(t *testing.T) {
	src := `
load("render.star", "render")
load("cache.star", "cache")

def main():
    cache.set("key_one", '1')
    cache.set("key_two", '2')

    one, two = cache.get("key_one"), cache.get("key_two")

    if one != '1' or two != '2':
        fail("didn't get what I set")

    three = cache.get("key_three")
    if three != None:
        fail("got something I hadn't set")

    cache.set("key_three", '3')
    three = cache.get("key_three")
    if three != '3':
        fail("didn't get the previously unset thing even though I just set it")

    return [render.Root(child=render.Box()) for i in range(int(one) + int(two) + int(three))]
`
	c := NewInMemoryCache()
	t.Cleanup(c.Close)
	InitCache(c)
	app, err := NewApplet(t.Context(), "test.star", []byte(src), WithTests(t))
	require.NoError(t, err)
	assert.NotNil(t, app)
	roots, err := app.Run(t.Context())
	require.NoError(t, err)
	assert.NotNil(t, roots)
	assert.Len(t, roots, 1+2+3)
}

func TestCacheSurvivesExecution(t *testing.T) {
	src := `
load("render.star", "render")
load("cache.star", "cache")

def main():
    i = int(cache.get("counter") or '1')
    frames = [render.Root(child=render.Box()) for _ in range(i)]
    cache.set("counter", str(i + 1))
    return frames
`
	c := NewInMemoryCache()
	t.Cleanup(c.Close)
	InitCache(c)
	app, err := NewApplet(t.Context(), "test.star", []byte(src), WithTests(t))
	require.NoError(t, err)
	assert.NotNil(t, app)

	// first time, i == 1
	roots, err := app.Run(t.Context())
	require.NoError(t, err)
	assert.NotNil(t, roots)
	assert.Len(t, roots, 1)

	// i == 2
	roots, err = app.Run(t.Context())
	require.NoError(t, err)
	assert.NotNil(t, roots)
	assert.Len(t, roots, 2)

	// but run the same code using different filename, and cached
	// data ends up in a different namespace
	app, err = NewApplet(t.Context(), "test2.star", []byte(src), WithTests(t))
	require.NoError(t, err)
	assert.NotNil(t, app)

	roots, _ = app.Run(t.Context())
	assert.Len(t, roots, 1)

	roots, _ = app.Run(t.Context())
	assert.Len(t, roots, 2)

	roots, _ = app.Run(t.Context())
	assert.Len(t, roots, 3)
}

func TestCacheNoInit(t *testing.T) {
	src := `
load("render.star", "render")
load("cache.star", "cache")

def main():
    cache.set("key_one", str(1))

    one, two = cache.get("key_one"), cache.get("key_two")

    if one != None or two != None:
        fail("without cache init we should only get None")

    return render.Root(child=render.Box())
`
	InitCache(nil)
	app, err := NewApplet(t.Context(), "test.star", []byte(src), WithTests(t))
	require.NoError(t, err)
	assert.NotNil(t, app)
	screens, err := app.Run(t.Context())
	require.NoError(t, err)
	assert.NotNil(t, screens)
}

func TestCacheBadValue(t *testing.T) {
	src := `
load("render.star", "render")
load("cache.star", "cache")

def main():
    cache.set("that's not a string value", 1)
    return render.Root(child=render.Box())
`
	InitCache(nil)
	app, err := NewApplet(t.Context(), "test.star", []byte(src), WithTests(t))
	require.NoError(t, err)
	assert.NotNil(t, app)
	screens, err := app.Run(t.Context())
	require.Error(t, err)
	assert.Nil(t, screens)
}
