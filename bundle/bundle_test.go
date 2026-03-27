package bundle_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tronbyt/pixlet/bundle"
	"github.com/tronbyt/pixlet/bundle/testdata"
)

func TestBundleWriteAndLoad(t *testing.T) {
	// Ensure we can load the bundle from an app.
	sub, err := fs.Sub(testdata.FS, "testapp")
	require.NoError(t, err)
	ab, err := bundle.FromFS(sub)
	require.NoError(t, err)
	assert.Equal(t, "test-app", ab.Manifest.ID)
	assert.NotNil(t, ab.Source)

	// Create a temp directory.
	dir := t.TempDir()
	require.NoError(t, err)

	// Write bundle to the temp directory.
	err = ab.WriteBundleToPath(t.Context(), dir)
	require.NoError(t, err)

	// Ensure we can load up the bundle just created.
	path := filepath.Join(dir, bundle.AppBundleName)
	f, err := os.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	newBun, err := bundle.LoadBundle(f)
	require.NoError(t, err)
	assert.Equal(t, "test-app", newBun.Manifest.ID)
	assert.NotNil(t, ab.Source)

	// Ensure the loaded bundle contains the files we expect.
	filesExpected := []string{
		"manifest.yaml",
		"test_app.star",
		"test.txt",
		"a_subdirectory/hi.jpg",
	}
	for _, file := range filesExpected {
		_, err := newBun.Source.Open(file)
		require.NoError(t, err)
	}

	// Ensure the loaded bundle does not contain any extra files.
	_, err = newBun.Source.Open("unused.txt")
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestBundleWriteAndLoadWithoutRuntime(t *testing.T) {
	sub, err := fs.Sub(testdata.FS, "testapp")
	require.NoError(t, err)
	ab, err := bundle.FromFS(sub)
	require.NoError(t, err)
	assert.Equal(t, "test-app", ab.Manifest.ID)
	assert.NotNil(t, ab.Source)

	// Create a temp directory.
	dir := t.TempDir()
	require.NoError(t, err)

	// Write bundle to the temp directory, without tree-shaking.
	err = ab.WriteBundleToPath(t.Context(), dir, bundle.WithoutRuntime())
	require.NoError(t, err)

	// Ensure we can load up the bundle just created.
	path := filepath.Join(dir, bundle.AppBundleName)
	f, err := os.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	newBun, err := bundle.LoadBundle(f)
	require.NoError(t, err)
	assert.Equal(t, "test-app", newBun.Manifest.ID)
	assert.NotNil(t, ab.Source)

	// Ensure the loaded bundle contains the files we expect.
	filesExpected := []string{
		"manifest.yaml",
		"test_app.star",
		"test.txt",
		"a_subdirectory/hi.jpg",
		"unused.txt",
	}
	for _, file := range filesExpected {
		_, err := newBun.Source.Open(file)
		require.NoError(t, err)
	}
}

func TestLoadBundle(t *testing.T) {
	f, err := testdata.FS.Open("bundle.tar.gz")
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	ab, err := bundle.LoadBundle(f)
	require.NoError(t, err)
	assert.Equal(t, "test-app", ab.Manifest.ID)
	assert.NotNil(t, ab.Source)
}
func TestLoadBundleExcessData(t *testing.T) {
	f, err := testdata.FS.Open("excess-files.tar.gz")
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	ab, err := bundle.LoadBundle(f)
	require.NoError(t, err)
	assert.Equal(t, "test-app", ab.Manifest.ID)
	assert.NotNil(t, ab.Source)
}
