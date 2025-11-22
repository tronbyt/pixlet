package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/encode"
	"github.com/tronbyt/pixlet/runtime"
	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/canvas"
	"github.com/tronbyt/pixlet/server/loader"
)

const webpLevelFlag = "webp-level"

var (
	configJson        string
	output            string
	magnify           int
	imageOutputFormat string
	maxDuration       time.Duration
	silenceOutput     bool
	width             int
	height            int
	timeout           time.Duration
	colorFilter       string
	output2x          bool
	webpLevel         int32
)

func NewRenderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render [path] [<key>=value>]...",
		Short: "Run a Pixlet app with provided config parameters",
		Args:  cobra.MinimumNArgs(1),
		RunE:  renderRun,
		Long: `Render a Pixlet app with provided config parameters.

The path argument should be the path to the Pixlet app to run. The
app can be a single file with the .star extension, or a directory
containing multiple Starlark files and resources.
	`,
		ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
	}

	cmd.Flags().StringVarP(&configJson, "config", "c", "", "Config file in json format")
	_ = cmd.RegisterFlagCompletionFunc("config", cobra.FixedCompletions([]string{"json"}, cobra.ShellCompDirectiveNoFileComp))

	cmd.Flags().StringVarP(&output, "output", "o", "", "Path for rendered image")
	_ = cmd.RegisterFlagCompletionFunc("output", cobra.FixedCompletions(formats, cobra.ShellCompDirectiveFilterFileExt))

	cmd.Flags().StringVarP(&imageOutputFormat, "format", "", "webp", "Output format. One of webp|gif|avif")
	_ = cmd.RegisterFlagCompletionFunc("format", cobra.FixedCompletions(formats, cobra.ShellCompDirectiveNoFileComp))

	cmd.Flags().BoolVarP(&silenceOutput, "silent", "", false, "Silence print statements when rendering app")
	cmd.Flags().IntVarP(
		&magnify,
		"magnify",
		"m",
		1,
		"Increase image dimension by a factor (useful for debugging)",
	)
	_ = cmd.RegisterFlagCompletionFunc("magnify", cobra.NoFileCompletions)

	cmd.Flags().StringVar(
		&colorFilter,
		"color-filter",
		"",
		`Apply a color filter. (See "pixlet community list-color-filters")`,
	)
	_ = cmd.RegisterFlagCompletionFunc("color-filter", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		var s []string
		for _, v := range encode.ColorFilterValues() {
			desc, _ := v.Description()
			s = append(s, v.String()+"\t"+desc)
		}
		return s, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.Flags().IntVarP(
		&width,
		"width",
		"w",
		64,
		"Set width",
	)
	_ = cmd.RegisterFlagCompletionFunc("width", cobra.NoFileCompletions)

	cmd.Flags().IntVarP(
		&height,
		"height",
		"t",
		32,
		"Set height",
	)
	_ = cmd.RegisterFlagCompletionFunc("height", cobra.NoFileCompletions)

	cmd.Flags().DurationVarP(
		&maxDuration,
		"max-duration",
		"d",
		15*time.Second,
		"Maximum allowed animation duration",
	)
	_ = cmd.RegisterFlagCompletionFunc("max-duration", cobra.NoFileCompletions)

	cmd.Flags().DurationVarP(
		&timeout,
		"timeout",
		"",
		30*time.Second,
		"Timeout for execution",
	)
	_ = cmd.RegisterFlagCompletionFunc("timeout", cobra.NoFileCompletions)

	cmd.Flags().BoolVarP(
		&output2x,
		"2x",
		"2",
		false,
		"Render at 2x resolution",
	)
	cmd.Flags().Int32VarP(
		&webpLevel,
		webpLevelFlag,
		"z",
		encode.WebPLevelDefault,
		"WebP compression level (0–9): 0 fast/large, 9 slow/small",
	)
	_ = cmd.RegisterFlagCompletionFunc(webpLevelFlag, completeWebPLevel)

	return cmd
}

func renderRun(cmd *cobra.Command, args []string) error {
	path := args[0]

	// check if path exists, and whether it is a directory or a file
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}

	var outPath string
	if info.IsDir() {
		abs, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}

		outPath = filepath.Join(path, filepath.Base(abs))
	} else {
		if !strings.HasSuffix(path, ".star") {
			return fmt.Errorf("script file must have suffix .star: %s", path)
		}

		outPath = strings.TrimSuffix(path, ".star")
	}

	if output2x {
		outPath += "@2x"
	}

	imageFormat = loader.ImageWebP
	switch imageOutputFormat {
	case "webp":
		imageFormat = loader.ImageWebP
		outPath += ".webp"
		if flag := cmd.Flags().Lookup(webpLevelFlag); flag != nil && flag.Changed {
			encode.SetWebPLevel(webpLevel)
		}
	case "gif":
		imageFormat = loader.ImageGIF
		outPath += ".gif"
	case "avif":
		imageFormat = loader.ImageAVIF
		outPath += ".avif"
	default:
		slog.Warn("Invalid image format; defaulting to WebP.", "format", imageOutputFormat)
	}
	if output != "" {
		outPath = output
	}

	config := map[string]string{}

	if configJson != "" {
		// Open the JSON file.
		f, err := os.Open(configJson)
		if err != nil {
			return fmt.Errorf("file open error %v", err)
		}

		err = json.NewDecoder(f).Decode(&config)
		if err != nil {
			_ = f.Close()
			return fmt.Errorf("failed to unmarshal JSON %v: %w", configJson, err)
		}

		_ = f.Close()
	}

	for _, param := range args[1:] {
		key, val, ok := strings.Cut(param, "=")
		if !ok {
			return fmt.Errorf("parameters must be in form <key>=<value>, found %s", param)
		}
		config[key] = val
	}

	cache := runtime.NewInMemoryCache()
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)

	filters := &encode.RenderFilters{Magnify: magnify}
	if colorFilter != "" {
		if filters.ColorFilter, err = encode.ColorFilterString(colorFilter); err != nil {
			return err
		}
	}

	meta := canvas.Metadata{
		Width:  width,
		Height: height,
		Is2x:   output2x,
	}

	buf, _, err := loader.RenderApplet(path, config, meta, maxDuration, timeout, imageFormat, silenceOutput, nil, filters)
	if err != nil {
		return fmt.Errorf("error rendering: %w", err)
	}

	if outPath == "-" {
		_, err = os.Stdout.Write(buf)
	} else {
		err = os.WriteFile(outPath, buf, 0644)
	}

	if err != nil {
		return fmt.Errorf("writing %s: %s", outPath, err)
	}

	slog.Info("Rendered image", "path", outPath)
	return nil
}
