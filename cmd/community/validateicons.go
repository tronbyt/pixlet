package community

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/flags"

	"github.com/tronbyt/pixlet/icons"
	"github.com/tronbyt/pixlet/runtime"
)

func NewValidateIconsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate-icons <path>",
		Short:   "Validates the schema icons used are available in our mobile app.",
		Example: `pixlet community validate-icons examples/schema_hello_world`,
		Long: `This command determines if the icons selected in your app schema are supported
by our mobile app.`,
		Args:              cobra.ExactArgs(1),
		RunE:              ValidateIcons,
		ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
	}
	return cmd
}

func ValidateIcons(_ *cobra.Command, args []string) error {
	path := args[0]

	cache := runtime.NewInMemoryCache()
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)

	applet, err := runtime.NewAppletFromPath(
		path,
		runtime.WithPrintDisabled(),
		runtime.WithCanvasMeta(flags.NewMeta().Metadata),
	)
	if err != nil {
		return fmt.Errorf("failed to load applet: %w", err)
	}
	defer applet.Close()

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
