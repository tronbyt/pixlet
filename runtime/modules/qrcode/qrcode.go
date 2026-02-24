package qrcode

import (
	"fmt"
	"image/color"
	"sync"

	goqrcode "github.com/skip2/go-qrcode"
	"github.com/tronbyt/pixlet/internal/colorutil"
	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/canvas"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName = "qrcode"
)

var (
	once   sync.Once
	module starlark.StringDict
)

func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		module = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"generate": starlark.NewBuiltin("generate", generateQRCode),
				},
			},
		}
	})

	return module, nil
}

func generateQRCode(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		starUrl        starlark.String
		starSize       starlark.String
		starColor      starlark.Value = starlark.None
		starBackground starlark.Value = starlark.None
	)

	if err := starlark.UnpackArgs(
		"generate",
		args, kwargs,
		"url", &starUrl,
		"size", &starSize,
		"color?", &starColor,
		"background?", &starBackground,
	); err != nil {
		return nil, fmt.Errorf("unpacking arguments for generate: %w", err)
	}

	scale := 1
	if meta, err := canvas.FromThread(thread); err == nil && meta.Is2x {
		scale = 2
	}

	// Determine QRCode sizing information.
	version := 0
	imgSize := 0
	switch starSize.GoString() {
	case "small":
		version = 1
		imgSize = 21 * scale
	case "medium":
		version = 2
		imgSize = 25 * scale
	case "large":
		version = 3
		imgSize = 29 * scale
	default:
		return nil, fmt.Errorf("size must be small, medium, or large")
	}

	url := starUrl.GoString()
	code, err := goqrcode.NewWithForcedVersion(url, version, goqrcode.Low)
	if err != nil {
		return nil, err
	}

	// Set default styles.
	code.DisableBorder = true
	code.ForegroundColor = color.White
	code.BackgroundColor = color.Transparent

	// Override color if one is provided.
	if starColor != starlark.None {
		foreground, err := colorutil.Parse(starColor)
		if err != nil {
			return nil, fmt.Errorf("parsing foreground color: %w", err)
		}
		code.ForegroundColor = foreground
	}

	// Override background if one is provided.
	if starBackground != starlark.None {
		background, err := colorutil.Parse(starBackground)
		if err != nil {
			return nil, fmt.Errorf("parsing background color: %w", err)
		}
		code.BackgroundColor = background
	}

	png, err := code.PNG(imgSize)
	if err != nil {
		return nil, err
	}

	return starlark.String(png), nil
}
