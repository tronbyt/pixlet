package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/runtime"
)

type schemaOptions struct {
	output string
}

func NewSchemaCmd() *cobra.Command {
	opts := &schemaOptions{}

	cmd := &cobra.Command{
		Use:   "schema [path]",
		Short: "Print the configuration schema for a Pixlet app",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return schemaRun(args, opts)
		},
		Long: `Determine the configuration schema for a Pixlet app.

The path argument should be the path to the Pixlet app to run. The
app can be a single file with the .star extension, or a directory
containing multiple Starlark files and resources. The output is in
JSON format.
	`,
		ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
	}

	cmd.Flags().StringVarP(&opts.output, "output", "o", opts.output, "Path for schema")
	_ = cmd.RegisterFlagCompletionFunc("output", cobra.FixedCompletions([]string{"json"}, cobra.ShellCompDirectiveFilterFileExt))

	return cmd
}

func schemaRun(args []string, opts *schemaOptions) error {
	path := args[0]

	applet, err := runtime.NewAppletFromPath(path)
	if err != nil {
		return fmt.Errorf("failed to load applet: %w", err)
	}
	defer applet.Close()

	if opts.output == "" || opts.output == "-" {
		buf, err := json.MarshalIndent(applet.Schema, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(buf))
	} else {
		b, err := json.Marshal(applet.Schema)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

		err = os.WriteFile(opts.output, b, 0644)
		if err != nil {
			return fmt.Errorf("failed to write schema to file: %w", err)
		}
	}

	return nil
}
