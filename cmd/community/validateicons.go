package community

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/flags"

	"github.com/tronbyt/pixlet/icons"
	"github.com/tronbyt/pixlet/runtime"
)

func NewValidateIconsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate-icons [path]",
		Short:   "Validates the schema icons used are available in our mobile app.",
		Example: `pixlet community validate-icons examples/schema_hello_world`,
		Long: `This command determines if the icons selected in your app schema are supported
by our mobile app.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) != 0 {
				path = args[0]
			}

			if err := ValidateIcons(cmd.Context(), path); err != nil {
				return err
			}

			if path == "." {
				if abs, err := filepath.Abs(path); err == nil {
					path = filepath.Base(abs)
				}
			}

			slog.Info("App icons are valid", "path", path)
			return nil
		},
		ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
	}
	return cmd
}

func ValidateIcons(ctx context.Context, path string) error {
	cache, err := flags.NewCache().Load(ctx)
	if err != nil {
		return err
	}
	defer cache.Close()

	applet, err := runtime.NewAppletFromPath(
		ctx, path,
		runtime.WithPrintDisabled(),
		runtime.WithCanvasMeta(flags.NewMeta().Metadata),
	)
	if err != nil {
		return fmt.Errorf("failed to load applet: %w", err)
	}
	defer func() { _ = applet.Close() }()

	if applet.Schema != nil {
		for _, field := range applet.Schema.Fields {
			if field.Icon == "" {
				continue
			}

			if _, ok := icons.IconsMap[field.Icon]; !ok {
				return fmt.Errorf("app '%s' contains unknown icon: '%s'", applet.ID, field.Icon)
			}
		}
	}

	return nil
}
