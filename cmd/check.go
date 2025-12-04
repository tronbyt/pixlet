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
	skipBroken    bool
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
	cmd.Flags().BoolVarP(&opts.skipBroken, "skip-broken", "s", opts.skipBroken, "skip apps marked as broken in their manifest")
	cmd.Flags().DurationVarP(&opts.maxRenderTime, "max-render-time", "", opts.maxRenderTime, "override the default max render time")
	_ = cmd.RegisterFlagCompletionFunc("max-render-time", cobra.NoFileCompletions)

	return cmd
}

func checkRun(cmd *cobra.Command, args []string, opts *checkOptions) error {
	foundIssue := false

	// checkApp is a helper to run checks on a single app.
	checkApp := func(path string) bool {
		// check if path exists, and whether it is a directory or a file
		info, err := os.Stat(path)
		if err != nil {
			// This path might be checked inside a WalkDir where we expect it to exist,
			// or passed as an argument.
			// If it's passed as arg, we want to error out.
			// But checkApp helper needs to be robust.
			// Let's rely on failure() to report issues.
			failure(path, fmt.Errorf("failed to stat %s: %w", path, err), "ensure the path exists")
			return true
		}

		baseDir := path
		if !info.IsDir() {
			if !strings.HasSuffix(path, ".star") {
				failure(path, fmt.Errorf("script file must have suffix .star: %s", path), "ensure the script file ends with .star")
				return true
			}
			baseDir = filepath.Dir(path)
		}

		fsys := os.DirFS(baseDir)

		// Check if app manifest exists and load it.
		manifestFile := filepath.Join(baseDir, manifest.ManifestFileName)
		manifestBytes, err := os.ReadFile(manifestFile)
		if err != nil {
			if os.IsNotExist(err) {
				failure(path, fmt.Errorf("couldn't find app manifest"), fmt.Sprintf("try `pixlet community create-manifest %s`", manifestFile))
			} else {
				failure(path, fmt.Errorf("couldn't read app manifest: %w", err), "ensure the manifest file is readable")
			}
			return true
		}

		m, err := manifest.LoadManifest(strings.NewReader(string(manifestBytes)))
		if err != nil {
			failure(path, fmt.Errorf("couldn't parse app manifest: %w", err), "ensure the manifest file is valid YAML")
			return true
		}

		if err := m.Validate(); err != nil {
			failure(path, fmt.Errorf("manifest didn't validate: %w", err), "try correcting the validation issue by updating your manifest")
			return true
		}

		if opts.skipBroken && m.Broken {
			fmt.Printf("Skipping %s: marked as broken in manifest\n", path)
			return false
		}

		// Check if an app can load.
		err = community.LoadApp(cmd, []string{path})
		if err != nil {
			failure(path, fmt.Errorf("app failed to load: %w", err), "try `pixlet community load-app` and resolve any runtime issues")
			return true
		}

		// Ensure icons are valid.
		err = community.ValidateIcons(cmd, []string{path})
		if err != nil {
			failure(path, fmt.Errorf("app has invalid icons: %w", err), "try `pixlet community list-icons` for the full list of valid icons")
			return true
		}

		// Check if app renders.
		renderOpts := newRenderOptions()
		renderOpts.silenceOutput = true
		renderOpts.output = os.DevNull
		renderOpts.log = slog.New(
			tint.NewHandler(os.Stderr, &tint.Options{
				Level:      slog.LevelWarn,
				TimeFormat: time.TimeOnly,
				NoColor:    !isatty.IsTerminal(os.Stderr.Fd()),
			}),
		)

		err = renderRun(cmd, []string{path}, renderOpts)
		if err != nil {
			failure(path, fmt.Errorf("app failed to render: %w", err), "try `pixlet render` and resolve any runtime issues")
			return true
		}

		// Check performance.
		p, err := ProfileApp(path, map[string]string{}, flags.NewMeta().Metadata)
		if err != nil {
			failure(path, fmt.Errorf("app profiling failed: %w", err), "try `pixlet profile` to debug performance issues")
			return true
		}
		if p.DurationNanos > opts.maxRenderTime.Nanoseconds() {
			failure(
				path,
				fmt.Errorf("app takes too long to render %s", time.Duration(p.DurationNanos)),
				fmt.Sprintf("try optimizing your app using `pixlet profile %s` to get it under %s", path, time.Duration(opts.maxRenderTime)),
			)
			return true
		}

		issueFound := false
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
				failure(p, fmt.Errorf("app is not formatted correctly: %w", err), fmt.Sprintf("try `pixlet format %s`", realPath))
				issueFound = true
			}

			lintOpts := newLintOptions()
			lintOpts.outputFormat = "off"
			err = lintRun([]string{realPath}, lintOpts)
			if err != nil {
				failure(p, fmt.Errorf("app has lint warnings: %w", err), fmt.Sprintf("try `pixlet lint --fix %s`", realPath))
				issueFound = true
			}

			return nil
		})

		if issueFound {
			return true
		}

		// If we're here, the app and manifest are good to go!
		success(path)
		return false
	}

	for _, path := range args {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", path, err)
		}

		if opts.recursive && info.IsDir() {
			err := filepath.WalkDir(path, func(walkPath string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.IsDir() {
					// Skip subdirectories that start with a dot.
					if walkPath != path && strings.HasPrefix(d.Name(), ".") {
						return filepath.SkipDir
					}
					return nil
				}

				if d.Name() == manifest.ManifestFileName {
					if checkApp(filepath.Dir(walkPath)) {
						foundIssue = true
					}
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to walk %s: %w", path, err)
			}
		} else {
			if checkApp(path) {
				foundIssue = true
			}
		}
	}

	if foundIssue {
		return fmt.Errorf("one or more apps failed checks")
	}

	return nil
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

