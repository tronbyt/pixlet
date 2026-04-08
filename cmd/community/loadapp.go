package community

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/flags"
	"github.com/tronbyt/pixlet/runtime"
)

func NewLoadAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "load-app [PATH]",
		Short:   "Validates an app can be successfully loaded in our runtime.",
		Example: `pixlet community load-app examples/clock`,
		Long:    `This command ensures an app can be loaded into our runtime successfully.`,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) != 0 {
				path = args[0]
			}

			app, err := LoadApp(cmd.Context(), path)
			if err != nil {
				return err
			}
			_ = app.Close()

			slog.Info("App loaded successfully", "path", app.MainFile)
			return nil
		},
		ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
	}
	return cmd
}

func LoadApp(ctx context.Context, path string) (*runtime.Applet, error) {
	cache, err := flags.NewCache().Load(ctx)
	if err != nil {
		return nil, err
	}
	defer cache.Close()

	app, err := runtime.NewAppletFromPath(
		ctx, path,
		runtime.WithPrintDisabled(),
		runtime.WithCanvasMeta(flags.NewMeta().Metadata),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load applet: %w", err)
	}

	return app, nil
}
