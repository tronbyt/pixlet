package manifest_test

import (
	"bytes"
	_ "embed"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tronbyt/pixlet/manifest"
	"github.com/tronbyt/pixlet/manifest/testdata"
)

//go:embed testdata/source.star
var source []byte

var output string = `---
id: foo-tracker
name: Foo Tracker
summary: Track realtime foo
desc: The foo tracker provides realtime feeds for foo.
author: Tidbyt
`

func TestManifest(t *testing.T) {
	m := manifest.Manifest{
		ID:      "foo-tracker",
		Name:    "Foo Tracker",
		Summary: "Track realtime foo",
		Desc:    "The foo tracker provides realtime feeds for foo.",
		Author:  "Tidbyt",
		Source:  source,
	}

	expected, err := testdata.FS.ReadFile("source.star")
	require.NoError(t, err)
	assert.Equal(t, expected, m.Source)
}

func TestLoadManifest(t *testing.T) {
	f, err := testdata.FS.Open("manifest.yaml")
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	m, err := manifest.LoadManifest(f)
	require.NoError(t, err)

	assert.Equal(t, "fuzzy-clock", m.ID)
	assert.Equal(t, "Fuzzy Clock", m.Name)
	assert.Equal(t, "Max Timkovich", m.Author)
	assert.Equal(t, "Human readable time", m.Summary)
	assert.Equal(t, "Display the time in a groovy, human-readable way.", m.Desc)
}

func TestWriteManifest(t *testing.T) {
	m := manifest.Manifest{
		ID:      "foo-tracker",
		Name:    "Foo Tracker",
		Summary: "Track realtime foo",
		Desc:    "The foo tracker provides realtime feeds for foo.",
		Author:  "Tidbyt",
		Source:  source,
	}

	buff := bytes.Buffer{}
	err := m.WriteManifest(&buff)
	require.NoError(t, err)

	b, err := io.ReadAll(&buff)
	require.NoError(t, err)

	assert.Equal(t, output, string(b))
}

func TestGeneratePackageName(t *testing.T) {
	type test struct {
		input string
		want  string
	}

	tests := []test{
		{input: "Cool App", want: "coolapp"},
		{input: "CoolApp", want: "coolapp"},
		{input: "cool-app", want: "coolapp"},
		{input: "cool_app", want: "coolapp"},
	}

	for _, tc := range tests {
		got := manifest.GenerateDirName(tc.input)
		assert.Equal(t, tc.want, got)
	}
}

func TestGenerateFileName(t *testing.T) {
	type test struct {
		input string
		want  string
	}

	tests := []test{
		{input: "Cool App", want: "cool_app.star"},
		{input: "CoolApp", want: "coolapp.star"},
		{input: "cool-app", want: "cool_app.star"},
		{input: "cool_app", want: "cool_app.star"},
	}

	for _, tc := range tests {
		got := manifest.GenerateFileName(tc.input)
		assert.Equal(t, tc.want, got)
	}
}

func TestGenerateID(t *testing.T) {
	type test struct {
		input string
		want  string
	}

	tests := []test{
		{input: "Cool App", want: "cool-app"},
		{input: "CoolApp", want: "coolapp"},
		{input: "cool-app", want: "cool-app"},
		{input: "cool_app", want: "cool-app"},
	}

	for _, tc := range tests {
		got := manifest.GenerateID(tc.input)
		assert.Equal(t, tc.want, got)
	}
}
