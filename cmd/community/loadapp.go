package community

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/flags"
	"github.com/tronbyt/pixlet/runtime"
)

func NewLoadAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "load-app [path]",
		Short:   "Validates an app can be successfully loaded in our runtime.",
		Example: `pixlet community load-app examples/clock`,
		Long:    `This command ensures an app can be loaded into our runtime successfully.`,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) != 0 {
				path = args[0]
			}

			if err := LoadApp(cmd.Context(), path); err != nil {
				return err
			}

			if path == "." {
				if abs, err := filepath.Abs(path); err == nil {
					path = filepath.Base(abs)
				}
			}

			slog.Info("App loaded successfully", "path", path)
			return nil
		},
		ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
	}
	return cmd
}

func LoadApp(ctx context.Context, path string) error {
	cache, err := flags.NewCache().Load(ctx)
	if err != nil {
		return err
	}
	defer cache.Close()

	app, err := runtime.NewAppletFromPath(
		ctx, path,
		runtime.WithPrintDisabled(),
		runtime.WithCanvasMeta(flags.NewMeta().Metadata),
	)
	if err != nil {
		return fmt.Errorf("failed to load applet: %w", err)
	}
	defer func() { _ = app.Close() }()

	return nil
}
