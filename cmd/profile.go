package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	pprof_driver "github.com/google/pprof/driver"
	pprof_profile "github.com/google/pprof/profile"
	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/canvas"
	"go.starlark.net/starlark"

	"github.com/tronbyt/pixlet/runtime"
)

var (
	pprof_cmd       string
	profileWidth    int
	profileHeight   int
	profileOutput2x bool
)

func init() {
	ProfileCmd.Flags().StringVarP(
		&pprof_cmd, "pprof", "", "top 10", "Command to call pprof with",
	)
	_ = ProfileCmd.RegisterFlagCompletionFunc("pprof", cobra.NoFileCompletions)
	ProfileCmd.Flags().IntVarP(
		&profileWidth,
		"width",
		"w",
		64,
		"Set width",
	)
	_ = ProfileCmd.RegisterFlagCompletionFunc("width", cobra.NoFileCompletions)
	ProfileCmd.Flags().IntVarP(
		&profileHeight,
		"height",
		"t",
		32,
		"Set height",
	)
	_ = ProfileCmd.RegisterFlagCompletionFunc("height", cobra.NoFileCompletions)
	ProfileCmd.Flags().BoolVarP(
		&profileOutput2x,
		"2x",
		"2",
		false,
		"Render at 2x resolution",
	)
}

var ProfileCmd = &cobra.Command{
	Use:               "profile <path> [<key>=value>]...",
	Short:             "Run a Pixlet app and print its execution-time profile",
	Args:              cobra.MinimumNArgs(1),
	RunE:              profile,
	ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
}

// We save the profile into an in-memory buffer, which is simpler than the tool expects.
// Simple adapter to pipe it through.
type FetchFunc func(src string, duration, timeout time.Duration) (*pprof_profile.Profile, string, error)

func (f FetchFunc) Fetch(src string, duration, timeout time.Duration) (*pprof_profile.Profile, string, error) {
	return f(src, duration, timeout)
}
func MakeFetchFunc(prof *pprof_profile.Profile) FetchFunc {
	return func(src string, duration, timeout time.Duration) (*pprof_profile.Profile, string, error) {
		return prof, "", nil
	}
}

// Calls the pprof program to print the top users of CPU, then exit
type printUI struct{}

var pprof_printed = false

func (u printUI) ReadLine(prompt string) (string, error) {
	if pprof_printed {
		os.Exit(0)
	}
	pprof_printed = true
	return pprof_cmd, nil
}
func (u printUI) Print(args ...interface{})                    {}
func (u printUI) PrintErr(args ...interface{})                 {}
func (u printUI) IsTerminal() bool                             { return false }
func (u printUI) WantBrowser() bool                            { return false }
func (u printUI) SetAutoComplete(complete func(string) string) {}

func profile(cmd *cobra.Command, args []string) error {
	path := args[0]

	config := map[string]string{}
	for _, param := range args[1:] {
		split := strings.Split(param, "=")
		if len(split) != 2 {
			return fmt.Errorf("parameters must be on form <key>=<value>, found %s", param)
		}
		config[split[0]] = split[1]
	}

	profile, err := ProfileApp(path, config, profileWidth, profileHeight, profileOutput2x)
	if err != nil {
		return err
	}

	options := &pprof_driver.Options{
		Fetch: MakeFetchFunc(profile),
		UI:    printUI{},
	}
	if err = pprof_driver.PProf(options); err != nil {
		return fmt.Errorf("could not start pprof driver: %w", err)
	}

	return nil
}

func ProfileApp(path string, config map[string]string, width int, height int, is2x bool) (*pprof_profile.Profile, error) {
	cache := runtime.NewInMemoryCache()
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)

	applet, err := runtime.NewAppletFromPath(
		path,
		runtime.WithPrintDisabled(),
		runtime.WithCanvasMeta(canvas.Metadata{
			Width:  width,
			Height: height,
			Is2x:   is2x,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load applet: %w", err)
	}
	defer applet.Close()

	buf := new(bytes.Buffer)
	if err = starlark.StartProfile(buf); err != nil {
		return nil, fmt.Errorf("error starting profiler: %w", err)
	}

	_, err = applet.RunWithConfig(context.Background(), config)
	if err != nil {
		_ = starlark.StopProfile()
		return nil, fmt.Errorf("error running script: %w", err)
	}

	if err = starlark.StopProfile(); err != nil {
		return nil, fmt.Errorf("error stopping profiler: %w", err)
	}

	profile, err := pprof_profile.ParseData(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("could not parse pprof profile: %w", err)
	}

	return profile, nil
}
