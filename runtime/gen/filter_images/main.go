package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/tronbyt/pixlet/assets"
	"github.com/tronbyt/pixlet/encode"
	"github.com/tronbyt/pixlet/runtime"
	"golang.org/x/sync/errgroup"
)

func main() {
	// We use triple quotes for the SVG string in Starlark to handle inner quotes easily.
	// Since SVG might contain triple quotes (unlikely but possible), we should be safe.
	// The provided SVG uses double quotes.

	examples := map[string]string{
		"filter_Blur_0": `
filter.Blur(
    child = %s,
    radius = 2.0,
)
`,
		"filter_Brightness_0": `
filter.Brightness(
    child = %s,
    change = -0.5,
)
`,
		"filter_Contrast_0": `
filter.Contrast(
    child = %s,
    factor = 2.0,
)
`,
		"filter_EdgeDetection_0": `
filter.EdgeDetection(
    child = %s,
    radius = 2.0,
)
`,
		"filter_Emboss_0": `
filter.Emboss(
    child = %s,
)
`,
		"filter_FlipHorizontal_0": `
filter.FlipHorizontal(
    child = %s,
)
`,
		"filter_FlipVertical_0": `
filter.FlipVertical(
    child = %s,
)
`,
		"filter_Gamma_0": `
filter.Gamma(
    child = %s,
    gamma = 0.5,
)
`,
		"filter_Grayscale_0": `
filter.Grayscale(
    child = %s,
)
`,
		"filter_Hue_0": `
filter.Hue(
    child = %s,
    change = 180.0,
)
`,
		"filter_Invert_0": `
filter.Invert(
    child = %s,
)
`,
		"filter_Rotate_0": `
filter.Rotate(
    child = %s,
    angle = 10.0,
)
`,
		"filter_Saturation_0": `
filter.Saturation(
    child = %s,
    factor = 1,
)
`,
		"filter_Sepia_0": `
filter.Sepia(
    child = %s,
)
`,
		"filter_Sharpen_0": `
filter.Sharpen(
    child = %s,
)
`,
		"filter_Shear_0": `
filter.Shear(
    child = %s,
    x_angle = 10.0,
    y_angle = 0.0,
)
`,
		"filter_Threshold_0": `
filter.Threshold(
    child = %s,
    level = 128.0,
)
`,
	}

	encode.SetWebPLevel(9)
	group, ctx := errgroup.WithContext(context.Background())

	for name, snippet := range examples {
		group.Go(func() error {
			snippet = fmt.Sprintf(snippet, `render.Column(
        cross_align = "center",
        children = [
            render.Image(src=LOGO, width=64, height=64),
            render.Text("Tronbyt"),
        ])
        `)

			src := fmt.Sprintf(`
load("render.star", "render")
load("filter.star", "filter")

LOGO = '''%s'''

def main():
    w = %s
    return render.Root(child=w)
`, assets.Logo, strings.TrimSpace(snippet))

			app, err := runtime.NewApplet(name, []byte(src))
			if err != nil {
				return err
			}

			roots, err := app.Run(ctx)
			if err != nil {
				return err
			}

			webp, err := encode.ScreensFromRoots(roots, 64, 74).EncodeWebP(15000, encode.Magnify(3))
			if err != nil {
				return err
			}

			path := filepath.Join("docs", "img", name+".webp")

			err = os.WriteFile(path, webp, 0644)
			if err != nil {
				return err
			}

			slog.Info("Generated image", "path", path)
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		panic(err)
	}
}
