package loader

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/canvas"
)

func TestNewRenderConfigSurfacesMaxDurationToCanvas(t *testing.T) {
	yes, no := true, false

	for _, tc := range []struct {
		name    string
		options []Option
		// what the app sees via canvas.max_duration_ms()
		wantMeta time.Duration
		// what the encoder is handed; ShowFullAnimation is applied later, in
		// RenderAppletRoot, so this stays as configured either way
		wantConf time.Duration
	}{
		{
			name:     "no ceiling configured",
			options:  nil,
			wantMeta: 0,
			wantConf: 0,
		},
		{
			name:     "ceiling reaches the canvas metadata",
			options:  []Option{WithMaxDuration(15 * time.Second)},
			wantMeta: 15 * time.Second,
			wantConf: 15 * time.Second,
		},
		{
			// WithMeta assigns Meta wholesale, so reconciling inside
			// WithMaxDuration would lose the value at this option order.
			name:     "WithMeta after WithMaxDuration does not clobber it",
			options:  []Option{WithMaxDuration(9 * time.Second), WithMeta(canvas.Metadata{Width: 64, Height: 32})},
			wantMeta: 9 * time.Second,
			wantConf: 9 * time.Second,
		},
		{
			name:     "WithMaxDuration after WithMeta also survives",
			options:  []Option{WithMeta(canvas.Metadata{Width: 64, Height: 32}), WithMaxDuration(9 * time.Second)},
			wantMeta: 9 * time.Second,
			wantConf: 9 * time.Second,
		},
		{
			// ShowFullAnimation makes the encoder ignore the ceiling, so the
			// app must be told there is none -- while conf.MaxDuration itself
			// is left alone for RenderAppletRoot to deal with.
			name:     "forced show_full_animation reports no ceiling",
			options:  []Option{WithMaxDuration(15 * time.Second), WithShowFullAnimation(&yes)},
			wantMeta: 0,
			wantConf: 15 * time.Second,
		},
		{
			name:     "explicitly disabled show_full_animation keeps the ceiling",
			options:  []Option{WithMaxDuration(15 * time.Second), WithShowFullAnimation(&no)},
			wantMeta: 15 * time.Second,
			wantConf: 15 * time.Second,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conf := NewRenderConfig("app.star", nil, tc.options...)
			assert.Equal(t, tc.wantMeta, conf.Meta.MaxDuration, "canvas metadata")
			assert.Equal(t, tc.wantConf, conf.MaxDuration, "encoder ceiling")
		})
	}
}

// The metadata must also keep the dimensions WithMeta supplied.
func TestNewRenderConfigKeepsMetaDimensions(t *testing.T) {
	conf := NewRenderConfig("app.star", nil,
		WithMeta(canvas.Metadata{Width: 128, Height: 64, Is2x: true}),
		WithMaxDuration(12*time.Second),
	)
	assert.Equal(t, 128, conf.Meta.Width)
	assert.Equal(t, 64, conf.Meta.Height)
	assert.True(t, conf.Meta.Is2x)
	assert.Equal(t, 12*time.Second, conf.Meta.MaxDuration)
}
