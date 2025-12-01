package cmd

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/community"
	"github.com/tronbyt/pixlet/cmd/flags"
	"github.com/tronbyt/pixlet/manifest"
)

type checkOptions struct {
	recursive     bool
	maxRenderTime time.Duration
}

func NewCheckCmd() *cobra.Command {
	opts := &checkOptions{
		maxRenderTime: 1 * time.Second,
	}

	cmd := &cobra.Command{
		Use:     "check <path>...",
		Example: `pixlet check examples/clock`,
		Short:   "Check if an app is ready to publish",
		Long: `Check if an app is ready to publish.

The path argument should be the path to the Pixlet app to check. The
app can be a single file with the .star extension, or a directory
containing multiple Starlark files and resources.

The check command runs a series of checks to ensure your app is ready
to publish in the community repo. Every failed check will have a solution
provided. If your app fails a check, try the provided solution and reach out on
Discord if you get stuck.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkRun(cmd, args, opts)
		},
		ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
	}

	cmd.Flags().BoolVarP(&opts.recursive, "recursive", "r", opts.recursive, "find apps recursively")
	cmd.Flags().DurationVarP(&opts.maxRenderTime, "max-render-time", "", opts.maxRenderTime, "override the default max render time")
	_ = cmd.RegisterFlagCompletionFunc("max-render-time", cobra.NoFileCompletions)

	return cmd
}

func checkRun(cmd *cobra.Command, args []string, opts *checkOptions) error {
	if opts.recursive {
		// TODO: implement recursive traversal for check command.
	}

	// check every path.
	foundIssue := false
	for _, path := range args {
		// check if path exists, and whether it is a directory or a file
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", path, err)
		}

		baseDir := path
		if !info.IsDir() {
			if !strings.HasSuffix(path, ".star") {
				return fmt.Errorf("script file must have suffix .star: %s", path)
			}
			baseDir = filepath.Dir(path)
		}

		fsys := os.DirFS(baseDir)

		// Check if an app can load.
		err = community.LoadApp(cmd, []string{path})
		if err != nil {
			foundIssue = true
			failure(path, fmt.Errorf("app failed to load: %w", err), "try `pixlet community load-app` and resolve any runtime issues")
			continue
		}

		// Ensure icons are valid.
		err = community.ValidateIcons(cmd, []string{path})
		if err != nil {
			foundIssue = true
			failure(path, fmt.Errorf("app has invalid icons: %w", err), "try `pixlet community list-icons` for the full list of valid icons")
			continue
		}

		// Check app manifest exists
		if !doesManifestExist(baseDir) {
			foundIssue = true
			failure(path, fmt.Errorf("couldn't find app manifest"), fmt.Sprintf("try `pixlet community create-manifest %s`", filepath.Join(baseDir, manifest.ManifestFileName)))
			continue
		}

		// Validate manifest.
		manifestFile := filepath.Join(baseDir, manifest.ManifestFileName)
		err = community.ValidateManifest(cmd, []string{manifestFile})
		if err != nil {
			foundIssue = true
			failure(path, fmt.Errorf("manifest didn't validate: %w", err), "try correcting the validation issue by updating your manifest")
			continue
		}

		// Create temporary file for app rendering.
		f, err := os.CreateTemp("", "pixlet-check-"+filepath.Base(baseDir)+"-*")
		if err != nil {
			return fmt.Errorf("could not create temp file for rendering, check your system: %w", err)
		}
		defer os.Remove(f.Name())

		// Check if app renders.
		renderOpts := newRenderOptions()
		renderOpts.silenceOutput = true
		renderOpts.output = f.Name()
		renderOpts.log = slog.New(
			tint.NewHandler(os.Stderr, &tint.Options{
				Level:      slog.LevelWarn,
				TimeFormat: time.TimeOnly,
				NoColor:    !isatty.IsTerminal(os.Stderr.Fd()),
			}),
		)

		err = renderRun(cmd, []string{path}, renderOpts)
		if err != nil {
			foundIssue = true
			failure(path, fmt.Errorf("app failed to render: %w", err), "try `pixlet render` and resolve any runtime issues")
			continue
		}

		// Check performance.
		p, err := ProfileApp(path, map[string]string{}, flags.NewMeta().Metadata)
		if err != nil {
			return fmt.Errorf("could not profile app: %w", err)
		}
		if p.DurationNanos > opts.maxRenderTime.Nanoseconds() {
			foundIssue = true
			failure(
				path,
				fmt.Errorf("app takes too long to render %s", time.Duration(p.DurationNanos)),
				fmt.Sprintf("try optimizing your app using `pixlet profile %s` to get it under %s", path, time.Duration(opts.maxRenderTime)),
			)
			continue
		}

		// run format and lint on *.star files in the fs
		fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() || !strings.HasSuffix(p, ".star") {
				return nil
			}

			realPath := filepath.Join(baseDir, p)

			formatOpts := newFormatOptions()
			formatOpts.dryRun = true
			if err := formatRun([]string{realPath}, formatOpts); err != nil {
				foundIssue = true
				failure(p, fmt.Errorf("app is not formatted correctly: %w", err), fmt.Sprintf("try `pixlet format %s`", realPath))
			}

			lintOpts := newLintOptions()
			lintOpts.outputFormat = "off"
			err = lintRun([]string{realPath}, lintOpts)
			if err != nil {
				foundIssue = true
				failure(p, fmt.Errorf("app has lint warnings: %w", err), fmt.Sprintf("try `pixlet lint --fix %s`", realPath))
			}

			return nil
		})

		// If we're here, the app and manifest are good to go!
		success(path)
	}

	if foundIssue {
		return fmt.Errorf("one or more apps failed checks")
	}

	return nil
}

func doesManifestExist(dir string) bool {
	file := filepath.Join(dir, manifest.ManifestFileName)
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false
	}

	if err != nil {
		return false
	}

	return true
}

func success(app string) {
	c := color.New(color.FgGreen)
	c.Printf("✔️ %s\n", app)
}

func failure(app string, err error, sol string) {
	c := color.New(color.FgRed)
	c.Printf("✖ %s\n", app)

	// Ensure multiline errors are properly indented.
	multilineError := strings.Split(err.Error(), "\n")
	for index, line := range multilineError {
		if index == 0 {
			continue
		}

		// The builtin starlark Backtrace function prints the last line at an
		// awkward indentation level. This check helps keep the failure indented
		// at one more level to ensure it's even more clear what is broken.
		if (strings.Contains(line, "Error in") || strings.Contains(line, "Error:")) && index == len(multilineError)-1 {
			multilineError[index] = fmt.Sprintf("      %s", line)
		} else {
			multilineError[index] = fmt.Sprintf("  %s", line)
		}
	}
	problem := strings.Join(multilineError, "\n")

	fmt.Printf("  ▪️ Problem: %v\n", problem)
	fmt.Printf("  ▪️ Solution: %v\n", sol)
}
