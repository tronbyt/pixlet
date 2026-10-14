package community

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPromptCategory(t *testing.T) {
	var out bytes.Buffer
	got, err := promptCategory(strings.NewReader("22\nwether\n\n21\n"), &out)
	require.NoError(t, err)
	require.Equal(t, "weather", got)

	// The category list is printed once and never reprinted or erased.
	require.Equal(t, 1, strings.Count(out.String(), "Categories:"))
	require.Equal(t, 3, strings.Count(out.String(), "enter a number between 1 and 21"))
	require.Equal(t, 4, strings.Count(out.String(), "Category (enter a number or name): "))
	require.NotContains(t, out.String(), "\x1b")
}

func TestPromptCategoryByName(t *testing.T) {
	var out bytes.Buffer
	got, err := promptCategory(strings.NewReader("smart-home\n"), &out)
	require.NoError(t, err)
	require.Equal(t, "smart-home", got)
}

func TestPromptCategoryEOF(t *testing.T) {
	var out bytes.Buffer
	_, err := promptCategory(strings.NewReader("22\n"), &out)
	require.Error(t, err)
}
