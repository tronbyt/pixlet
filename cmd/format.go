package cmd

import (
	"fmt"

	"github.com/bazelbuild/buildtools/differ"
	"github.com/spf13/cobra"
)

type formatOptions struct {
	verbose   bool
	recursive bool
	dryRun    bool
}

func newFormatOptions() *formatOptions {
	return &formatOptions{}
}

func NewFormatCmd() *cobra.Command {
	opts := newFormatOptions()

	cmd := &cobra.Command{
		Use:   "format <pathspec>...",
		Short: "Formats Tronbyt apps",
		Example: `  pixlet format app.star
  pixlet format app.star --dry-run
  pixlet format --recursive ./`,
		Long: `The format command provides a code formatter for Tronbyt apps. By default, it
will format your starlark source code in line. If you wish you see the output
before applying, add the --dry-run flag.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return formatRun(args, opts)
		},
		ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
	}

	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", opts.verbose, "print verbose information to standard error")
	cmd.Flags().BoolVarP(&opts.recursive, "recursive", "r", opts.recursive, "find starlark files recursively")
	cmd.Flags().BoolVarP(&opts.dryRun, "dry-run", "d", opts.dryRun, "display a diff of formatting changes without modification")

	return cmd
}

func formatRun(args []string, opts *formatOptions) error {
	// Lint refers to the lint mode for buildifier, with the options being off,
	// warn, or fix. For pixlet format, we don't want to lint at all.
	lint := "off"

	// Mode refers to formatting mode for buildifier, with the options being
	// check, diff, or fix. For the pixlet format command, we want to fix the
	// resolvable issue by default and provide a dry run flag to be able to
	// diff the changes before fixing them.
	mode := "fix"
	if opts.dryRun {
		mode = "diff"
	}

	// Copied from the buildifier source, we need to supply a diff program for
	// the differ.
	differ, _ := differ.Find()
	diff = differ

	// Run buildifier and exit with the returned exit code.
	exitCode := runBuildifier(args, lint, mode, "", opts.recursive, opts.verbose)
	if exitCode != 0 {
		return fmt.Errorf("formatting returned non-zero exit status: %d", exitCode)
	}

	return nil
}
