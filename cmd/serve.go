package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/encode"
	"github.com/tronbyt/pixlet/server"
	"github.com/tronbyt/pixlet/server/loader"
)

var (
	host          string
	port          int
	path          string
	watch         bool
	serveFormat   string
	configOutFile string
)

func init() {
	ServeCmd.Flags().StringVarP(&configOutFile, "saveconfig", "o", "", "Output file for config changes")
	ServeCmd.Flags().StringVarP(&host, "host", "i", "127.0.0.1", "Host interface for serving rendered images")
	ServeCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port for serving rendered images")
	ServeCmd.Flags().BoolVarP(&watch, "watch", "w", true, "Reload scripts on change. Does not recurse sub-directories.")
	ServeCmd.Flags().IntVarP(&maxDuration, "max-duration", "d", 15000, "Maximum allowed animation duration (ms)")
	ServeCmd.Flags().IntVarP(&timeout, "timeout", "", 30000, "Timeout for execution (ms)")
	ServeCmd.Flags().StringVarP(&serveFormat, "format", "", "webp", "Image format. One of webp|gif|avif")
	ServeCmd.Flags().StringVarP(&path, "path", "", "/", "Path to serve the app on")
	ServeCmd.Flags().IntVar(&width, "width", 64, "Set width")
	ServeCmd.Flags().IntVarP(&height, "height", "t", 32, "Set height")
	ServeCmd.Flags().BoolVarP(&output2x, "2x", "2", false, "Render at 2x resolution")
	ServeCmd.Flags().Int32VarP(
		&webpLevel,
		webpLevelFlag,
		"z",
		encode.WebPLevelDefault,
		"WebP compression level (0–9): 0 fast/large, 9 slow/small",
	)

	// Deprecated flags
	ServeCmd.Flags().IntVar(&maxDuration, "max_duration", 15000, "Maximum allowed animation duration (ms)")
	if err := ServeCmd.Flags().MarkDeprecated(
		"max_duration", "use --max-duration instead",
	); err != nil {
		panic(err)
	}
}

var ServeCmd = &cobra.Command{
	Use:   "serve [path]",
	Short: "Serve a Pixlet app in a web server",
	Args:  cobra.ExactArgs(1),
	RunE:  serve,
	Long: `Serve a Pixlet app in a web server.

The path argument should be the path to the Pixlet program to run. The
program can be a single file with the .star extension, or a directory
containing multiple Starlark files and resources.`,
}

func serve(cmd *cobra.Command, args []string) error {
	imageFormat := loader.ImageWebP
	switch serveFormat {
	case "webp":
		imageFormat = loader.ImageWebP
		if cmd.Flags().Lookup(webpLevelFlag).Changed {
			encode.SetWebPLevel(webpLevel)
		}
	case "gif":
		imageFormat = loader.ImageGIF
	case "avif":
		imageFormat = loader.ImageAVIF
	default:
		log.Printf("Invalid image format %q. Defaulting to WebP.", serveFormat)
	}

	s, err := server.NewServer(host, port, path, watch, args[0], width, height, maxDuration, timeout, imageFormat, configOutFile, output2x)
	if err != nil {
		return err
	}
	return s.Run()
}
