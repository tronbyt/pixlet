package community

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/flags"
	"github.com/tronbyt/pixlet/runtime"
)

func NewLoadAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "load-app [path]",
		Short:             "Validates an app can be successfully loaded in our runtime.",
		Example:           `pixlet community load-app examples/clock`,
		Long:              `This command ensures an app can be loaded into our runtime successfully.`,
		Args:              cobra.MaximumNArgs(1),
		RunE:              LoadApp,
		ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
	}
	return cmd
}

func LoadApp(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) != 0 {
		path = args[0]
	}

	cache := runtime.NewInMemoryCache()
	defer cache.Close()
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)

	app, err := runtime.NewAppletFromPath(
		cmd.Context(),
		path,
		runtime.WithPrintDisabled(),
		runtime.WithCanvasMeta(flags.NewMeta().Metadata),
	)
	if err != nil {
		return fmt.Errorf("failed to load applet: %w", err)
	}
	defer func() { _ = app.Close() }()

	return nil
}
