package community

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/runtime"
)

var LoadAppCmd = &cobra.Command{
	Use:     "load-app <path>",
	Short:   "Validates an app can be successfully loaded in our runtime.",
	Example: `pixlet community load-app examples/clock`,
	Long:    `This command ensures an app can be loaded into our runtime successfully.`,
	Args:    cobra.ExactArgs(1),
	RunE:    LoadApp,
}

func LoadApp(cmd *cobra.Command, args []string) error {
	path := args[0]

	cache := runtime.NewInMemoryCache()
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)

	app, err := runtime.NewAppletFromPath(path, runtime.WithPrintDisabled())
	if err != nil {
		return fmt.Errorf("failed to load applet: %w", err)
	}
	defer app.Close()

	return nil
}
