package cmd

import (
	"fmt"

	"github.com/bazelbuild/buildtools/differ"
	"github.com/spf13/cobra"
)

type lintOptions struct {
	verbose      bool
	recursive    bool
	fix          bool
	outputFormat string
}

func newLintOptions() *lintOptions {
	return &lintOptions{}
}

func NewLintCmd() *cobra.Command {
	opts := newLintOptions()

	cmd := &cobra.Command{
		Use: "lint <pathspec>...",
		Example: `  pixlet lint app.star
  pixlet lint --recursive --fix ./`,
		Short: "Lints Tronbyt apps",
		Long: `The lint command provides a linter for Tronbyt apps. It's capable of linting a
file, a list of files, or directory with the recursive option. Additionally, it
provides an option to automatically fix resolvable linter issues.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return lintRun(args, opts)
		},
		ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
	}

	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", opts.verbose, "print verbose information to standard error")
	cmd.Flags().BoolVarP(&opts.recursive, "recursive", "r", opts.recursive, "find starlark files recursively")
	cmd.Flags().BoolVarP(&opts.fix, "fix", "f", opts.fix, "automatically fix resolvable lint issues")
	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", opts.outputFormat, "output format: text, json, or off")
	_ = cmd.RegisterFlagCompletionFunc("output", cobra.FixedCompletions([]string{"text", "json", "off"}, cobra.ShellCompDirectiveNoFileComp))

	return cmd
}

func lintRun(args []string, opts *lintOptions) error {
	// Mode refers to formatting mode for buildifier, with the options being
	// check, diff, or fix. For the pixlet lint command, we only want to check
	// formatting.
	mode := "check"

	// Lint refers to the lint mode for buildifier, with the options being off,
	// warn, or fix. For pixlet lint, we want to warn by default but offer a
	// flag to automatically fix resolvable issues.
	lint := "warn"

	// If the fix flag is enabled, the lint command should both format and lint.
	if opts.fix {
		mode = "fix"
		lint = "fix"
	}

	// Copied from the buildifier source, we need to supply a diff program for
	// the differ.
	differ, _ := differ.Find()
	diff = differ

	// Run buildifier and exit with the returned exit code.
	exitCode := runBuildifier(args, lint, mode, opts.outputFormat, opts.recursive, opts.verbose)
	if exitCode != 0 {
		return fmt.Errorf("linting failed with exit code: %d", exitCode)
	}

	// Buildifier will return a zero exit status when the fix flag is provided,
	// even if there are still lint issues that could not be fixed. So we need
	// to run it twice to get the full picture - once with fix enabled and once
	// more to determine what else needs to be fixed manually.
	if opts.fix {
		mode = "check"
		lint = "warn"

		exitCode := runBuildifier(args, lint, mode, opts.outputFormat, opts.recursive, opts.verbose)
		if exitCode != 0 {
			return fmt.Errorf("linting failed with exit code: %d", exitCode)
		}
	}

	return nil
}
