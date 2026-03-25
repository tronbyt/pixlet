package starlarkhttp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.starlark.net/starlark"
)

func TestGetAppIdentifier(t *testing.T) {
	tests := map[string]struct {
		threadName string
		want       string
	}{
		"thread name with run suffix": {
			threadName: "weather/abc123",
			want:       "weather",
		},
		"thread name without separator": {
			threadName: "weather",
			want:       "weather",
		},
		"thread name with leading separator": {
			threadName: "/abc123",
			want:       "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			thread := &starlark.Thread{Name: tc.threadName}
			got := getAppIdentifier(thread)
			assert.Equal(t, tc.want, got)
		})
	}
}
