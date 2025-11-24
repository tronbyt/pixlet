package loader

import (
	"time"

	"github.com/tronbyt/pixlet/encode"
	"github.com/tronbyt/pixlet/render"
	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/canvas"
)

type RenderConfig struct {
	Path          string
	Config        map[string]string
	Meta          canvas.Metadata
	MaxDuration   time.Duration
	Timeout       time.Duration
	ImageFormat   ImageFormat
	SilenceOutput bool
	Location      *time.Location
	Filters       encode.RenderFilters
}

func NewRenderConfig(path string, config map[string]string, options ...Option) *RenderConfig {
	conf := &RenderConfig{
		Path:   path,
		Config: config,
		Meta: canvas.Metadata{
			Width:  render.DefaultFrameWidth,
			Height: render.DefaultFrameHeight,
		},
		ImageFormat: ImageWebP,
		Location:    time.Local,
		Filters: encode.RenderFilters{
			Magnify: 1,
		},
	}
	for _, option := range options {
		option(conf)
	}
	return conf
}

type Option func(config *RenderConfig)

func WithLocation(location *time.Location) Option {
	return func(config *RenderConfig) {
		if location != nil {
			config.Location = location
		}
	}
}

func WithSilenceOutput(silenceOutput bool) Option {
	return func(config *RenderConfig) {
		config.SilenceOutput = silenceOutput
	}
}

func WithFilters(filters encode.RenderFilters) Option {
	return func(config *RenderConfig) {
		if filters.Magnify == 0 {
			filters.Magnify = 1
		}
		config.Filters = filters
	}
}

func WithImageFormat(imageFormat ImageFormat) Option {
	return func(config *RenderConfig) {
		config.ImageFormat = imageFormat
	}
}

func WithMaxDuration(maxDuration time.Duration) Option {
	return func(config *RenderConfig) {
		config.MaxDuration = maxDuration
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(config *RenderConfig) {
		config.Timeout = timeout
	}
}

func WithMeta(meta canvas.Metadata) Option {
	return func(config *RenderConfig) {
		if meta.Width == 0 {
			meta.Width = render.DefaultFrameWidth
		}
		if meta.Height == 0 {
			meta.Height = render.DefaultFrameHeight
		}
		config.Meta = meta
	}
}
