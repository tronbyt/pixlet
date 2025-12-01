package flags

import (
	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/render"
	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/canvas"
)

type Meta struct {
	canvas.Metadata
}

func NewMeta() Meta {
	return Meta{
		Metadata: canvas.Metadata{
			Width:  render.DefaultFrameWidth,
			Height: render.DefaultFrameHeight,
		},
	}
}

func (f *Meta) Register(cmd *cobra.Command) {
	fs := cmd.Flags()
	fs.IntVarP(&f.Width, "width", "w", f.Width, "Set width")
	_ = cmd.RegisterFlagCompletionFunc("width", cobra.NoFileCompletions)

	fs.IntVarP(&f.Height, "height", "t", f.Height, "Set height")
	_ = cmd.RegisterFlagCompletionFunc("height", cobra.NoFileCompletions)

	fs.BoolVarP(&f.Is2x, "2x", "2", f.Is2x, "Render at 2x resolution")
}
