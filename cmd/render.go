package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/encode"
	"github.com/tronbyt/pixlet/runtime"
	"github.com/tronbyt/pixlet/server/loader"
)

const webpLevelFlag = "webp-level"

var (
	configJson        string
	output            string
	magnify           int
	imageOutputFormat string
	maxDuration       int
	silenceOutput     bool
	width             int
	height            int
	timeout           int
	colorFilter       string
	output2x          bool
	webpLevel         int32

	// Deprecated: flag behavior has been moved to community.ListColorFiltersCmd
	listColorFilters bool
)

func init() {
	RenderCmd.Flags().StringVarP(&configJson, "config", "c", "", "Config file in json format")
	RenderCmd.Flags().StringVarP(&output, "output", "o", "", "Path for rendered image")
	RenderCmd.Flags().StringVarP(&imageOutputFormat, "format", "", "webp", "Output format. One of webp|gif|avif")
	RenderCmd.Flags().BoolVarP(&silenceOutput, "silent", "", false, "Silence print statements when rendering app")
	RenderCmd.Flags().IntVarP(
		&magnify,
		"magnify",
		"m",
		1,
		"Increase image dimension by a factor (useful for debugging)",
	)
	RenderCmd.Flags().StringVar(
		&colorFilter,
		"color-filter",
		"",
		`Apply a color filter. (See "pixlet community list-color-filters")`,
	)
	RenderCmd.Flags().IntVarP(
		&width,
		"width",
		"w",
		64,
		"Set width",
	)
	RenderCmd.Flags().IntVarP(
		&height,
		"height",
		"t",
		32,
		"Set height",
	)
	RenderCmd.Flags().IntVarP(
		&maxDuration,
		"max-duration",
		"d",
		15000,
		"Maximum allowed animation duration (ms)",
	)
	RenderCmd.Flags().IntVarP(
		&timeout,
		"timeout",
		"",
		30000,
		"Timeout for execution (ms)",
	)
	RenderCmd.Flags().BoolVarP(
		&output2x,
		"2x",
		"2",
		false,
		"Render at 2x resolution",
	)
	RenderCmd.Flags().Int32VarP(
		&webpLevel,
		webpLevelFlag,
		"z",
		encode.WebPLevelDefault,
		"WebP compression level (0–9): 0 fast/large, 9 slow/small",
	)

	// Deprecated flags
	RenderCmd.Flags().IntVar(
		&maxDuration,
		"max_duration",
		15000,
		"Maximum allowed animation duration (ms)",
	)
	if err := RenderCmd.Flags().MarkDeprecated(
		"max_duration", "use --max-duration instead",
	); err != nil {
		panic(err)
	}

	RenderCmd.Flags().StringVar(
		&colorFilter,
		"color_filter",
		"",
		"Apply a color filter (warm, cool, etc)",
	)
	if err := RenderCmd.Flags().MarkDeprecated(
		"color_filter", "use --color-filter instead",
	); err != nil {
		panic(err)
	}

	RenderCmd.Flags().BoolVar(
		&listColorFilters,
		"list-color-filters",
		false,
		"List available color filters",
	)
	if err := RenderCmd.Flags().MarkDeprecated(
		"list-color-filters", `use "pixlet community list-color-filters" instead`,
	); err != nil {
		panic(err)
	}
}

var RenderCmd = &cobra.Command{
	Use:   "render [path] [<key>=value>]...",
	Short: "Run a Pixlet app with provided config parameters",
	Args:  cobra.MinimumNArgs(1),
	RunE:  render,
	Long: `Render a Pixlet app with provided config parameters.

The path argument should be the path to the Pixlet app to run. The
app can be a single file with the .star extension, or a directory
containing multiple Starlark files and resources.
	`,
}

func render(cmd *cobra.Command, args []string) error {
	if listColorFilters {
		fmt.Println("Supported color filters:")
		for _, f := range encode.ColorFilterStrings() {
			fmt.Println(" -", f)
		}
		return nil
	}
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

	filters := &encode.RenderFilters{
		Magnify:  magnify,
		Output2x: output2x,
	}
	if colorFilter != "" {
		if filters.ColorFilter, err = encode.ColorFilterString(colorFilter); err != nil {
			return err
		}
	}

	buf, _, err := loader.RenderApplet(path, config, width, height, magnify, maxDuration, timeout, imageFormat, silenceOutput, filters)
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
