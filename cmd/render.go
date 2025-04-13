package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"tidbyt.dev/pixlet/runtime"
	"tidbyt.dev/pixlet/server/loader"
)

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
		"max_duration",
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
	path := args[0]

	// check if path exists, and whether it is a directory or a file
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}

	var outPath string
	if info.IsDir() {
		outPath = filepath.Join(path, filepath.Base(path))
	} else {
		if !strings.HasSuffix(path, ".star") {
			return fmt.Errorf("script file must have suffix .star: %s", path)
		}

		outPath = strings.TrimSuffix(path, ".star")
	}

	imageFormat = loader.ImageWebP
	switch imageOutputFormat {
	case "webp":
		imageFormat = loader.ImageWebP
		outPath += ".webp"
	case "gif":
		imageFormat = loader.ImageGIF
		outPath += ".gif"
	case "avif":
		imageFormat = loader.ImageAVIF
		outPath += ".avif"
	default:
		fmt.Printf("Invalid image format %q. Defaulting to WebP.", imageOutputFormat)
	}
	if output != "" {
		outPath = output
	}

	config := map[string]string{}

	if configJson != "" {
		// Open the JSON file.
		file, err := os.Open(configJson)
		if err != nil {
			return fmt.Errorf("file open error %v", err)
		}

		// Use the `json.Unmarshal()` function to unmarshal the JSON file into the map variable.
		fileData, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("file read error %v", err)
		}
		err = json.Unmarshal(fileData, &config)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON %v: %w", configJson, err)
		}
	}

	for _, param := range args[1:] {
		split := strings.Split(param, "=")
		if len(split) < 2 {
			return fmt.Errorf("parameters must be in form <key>=<value>, found %s", param)
		}
		config[split[0]] = strings.Join(split[1:], "=")
	}

	cache := runtime.NewInMemoryCache()
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)

	buf, _, err := loader.RenderApplet(path, config, width, height, magnify, maxDuration, timeout, imageFormat, silenceOutput)
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

	return nil
}
