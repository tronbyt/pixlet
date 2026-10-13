package manifest_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tronbyt/pixlet/manifest"
)

func TestValidateName(t *testing.T) {
	type test struct {
		input     string
		shouldErr bool
	}

	tests := []test{
		{input: "Cool App", shouldErr: false},
		{input: "Cool app", shouldErr: true},
		{input: "cool app", shouldErr: true},
		{input: "coolApp", shouldErr: true},
		{input: "An exceptionally lengthy and excessively wordy application title which surpasses the maximum allowed limit", shouldErr: true},
		{input: "", shouldErr: true},
		{input: "Clark's App", shouldErr: false},
	}

	for _, tc := range tests {
		err := manifest.ValidateName(tc.input)

		if tc.shouldErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestValidateCategory(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
	}{
		{input: "weather", shouldErr: false},
		{input: "smart-home", shouldErr: false},
		{input: "Weather", shouldErr: true},
		{input: "wether", shouldErr: true},
		{input: "", shouldErr: true},
	}

	for _, tc := range tests {
		err := manifest.ValidateCategory(tc.input)

		if tc.shouldErr {
			require.Error(t, err, tc.input)
		} else {
			require.NoError(t, err, tc.input)
		}
	}
}

func TestCategoriesSorted(t *testing.T) {
	require.True(t, sort.StringsAreSorted(manifest.Categories))
}

func TestResolveCategory(t *testing.T) {
	tests := []struct {
		input     string
		want      string
		shouldErr bool
	}{
		{input: "1", want: "art"},
		{input: "21", want: "weather"},
		{input: " 19 ", want: "travel"},
		{input: "smart-home", want: "smart-home"},
		{input: "0", shouldErr: true},
		{input: "22", shouldErr: true},
		{input: "wether", shouldErr: true},
		{input: "", shouldErr: true},
	}

	for _, tc := range tests {
		got, err := manifest.ResolveCategory(tc.input)

		if tc.shouldErr {
			require.Error(t, err, tc.input)
		} else {
			require.NoError(t, err, tc.input)
			require.Equal(t, tc.want, got, tc.input)
		}
	}
}

func TestCategoryGrid(t *testing.T) {
	grid := manifest.CategoryGrid(3)
	require.Equal(t, 7, strings.Count(grid, "\n"))
	for _, c := range manifest.Categories {
		require.Contains(t, grid, c)
	}
}

func TestValidateSummary(t *testing.T) {
	type test struct {
		input     string
		shouldErr bool
	}

	tests := []test{
		{input: "A cool app", shouldErr: false},
		{input: "A really very extremely incredibly super duper extra highly mega ultra cool app", shouldErr: true},
		{input: "A cool app.", shouldErr: true},
		{input: "A cool app!", shouldErr: true},
		{input: "A cool app?", shouldErr: true},
		{input: "a cool app", shouldErr: true},
		{input: "NYC Subway departures", shouldErr: false},
		{input: "", shouldErr: true},
	}

	for _, tc := range tests {
		err := manifest.ValidateSummary(tc.input)

		if tc.shouldErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestValidateDesc(t *testing.T) {
	type test struct {
		input     string
		shouldErr bool
	}

	tests := []test{
		{input: "A really cool app that does really cool app things.", shouldErr: false},
		{input: "a really cool app that does really cool app things.", shouldErr: true},
		{input: "A really cool app that does really cool app things", shouldErr: true},
		{input: "", shouldErr: true},
	}

	for _, tc := range tests {
		err := manifest.ValidateDesc(tc.input)

		if tc.shouldErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestValidateID(t *testing.T) {
	type test struct {
		input     string
		shouldErr bool
	}

	tests := []test{
		{input: "foo-bar", shouldErr: false},
		{input: "foobar", shouldErr: false},
		{input: "FooBar", shouldErr: true},
		{input: "foo$", shouldErr: true},
		{input: "", shouldErr: true},
	}

	for _, tc := range tests {
		err := manifest.ValidateID(tc.input)

		if tc.shouldErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
