package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/runtime"
)

var schemaOutput string

func NewSchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema [path]",
		Short: "Print the configuration schema for a Pixlet app",
		Args:  cobra.MinimumNArgs(1),
		RunE:  schemaRun,
		Long: `Determine the configuration schema for a Pixlet app.

The path argument should be the path to the Pixlet app to run. The
app can be a single file with the .star extension, or a directory
containing multiple Starlark files and resources. The output is in
JSON format.
	`,
		ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
	}

	cmd.Flags().StringVarP(&schemaOutput, "output", "o", "", "Path for schema")
	_ = cmd.RegisterFlagCompletionFunc("output", cobra.FixedCompletions([]string{"json"}, cobra.ShellCompDirectiveFilterFileExt))

	return cmd
}

func schemaRun(_ *cobra.Command, args []string) error {
	path := args[0]

	applet, err := runtime.NewAppletFromPath(path)
	if err != nil {
		return fmt.Errorf("failed to load applet: %w", err)
	}
	defer applet.Close()

	if schemaOutput == "" || schemaOutput == "-" {
		buf, err := json.MarshalIndent(applet.Schema, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(buf))
	} else {
		err = os.WriteFile(schemaOutput, applet.SchemaJSON, 0644)
		if err != nil {
			return fmt.Errorf("failed to write schema to file: %w", err)
		}
	}

	return nil
}
