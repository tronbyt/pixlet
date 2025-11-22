package community

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tronbyt/pixlet/icons"
	"github.com/tronbyt/pixlet/runtime"
	"github.com/tronbyt/pixlet/schema"
)

var ValidateIconsCmd = &cobra.Command{
	Use:     "validate-icons <path>",
	Short:   "Validates the schema icons used are available in our mobile app.",
	Example: `pixlet community validate-icons examples/schema_hello_world`,
	Long: `This command determines if the icons selected in your app schema are supported
by our mobile app.`,
	Args:              cobra.ExactArgs(1),
	RunE:              ValidateIcons,
	ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
}

func ValidateIcons(cmd *cobra.Command, args []string) error {
	path := args[0]

	cache := runtime.NewInMemoryCache()
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)

	applet, err := runtime.NewAppletFromPath(path, runtime.WithPrintDisabled())
	if err != nil {
		return fmt.Errorf("failed to load applet: %w", err)
	}
	defer applet.Close()

	s := schema.Schema{}
	js := applet.SchemaJSON
	if len(js) == 0 {
		return nil
	}

	err = json.Unmarshal(js, &s)
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	for _, field := range s.Fields {
		if field.Icon == "" {
			continue
		}

		if _, ok := icons.IconsMap[field.Icon]; !ok {
			return fmt.Errorf("app '%s' contains unknown icon: '%s'", applet.ID, field.Icon)
		}
	}

	return nil
}
