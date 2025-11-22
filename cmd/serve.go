package cmd

import (
	"log/slog"
	"time"

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
	_ = ServeCmd.RegisterFlagCompletionFunc("saveconfig", cobra.FixedCompletions([]string{"json"}, cobra.ShellCompDirectiveFilterFileExt))
	ServeCmd.Flags().StringVarP(&host, "host", "i", "127.0.0.1", "Host interface for serving rendered images")
	_ = ServeCmd.RegisterFlagCompletionFunc("host", cobra.NoFileCompletions)
	ServeCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port for serving rendered images")
	_ = ServeCmd.RegisterFlagCompletionFunc("port", cobra.NoFileCompletions)
	ServeCmd.Flags().BoolVarP(&watch, "watch", "w", true, "Reload scripts on change. Does not recurse sub-directories.")
	ServeCmd.Flags().DurationVarP(&maxDuration, "max-duration", "d", 15*time.Second, "Maximum allowed animation duration")
	_ = ServeCmd.RegisterFlagCompletionFunc("max-duration", cobra.NoFileCompletions)
	ServeCmd.Flags().DurationVarP(&timeout, "timeout", "", 30*time.Second, "Timeout for execution")
	_ = ServeCmd.RegisterFlagCompletionFunc("timeout", cobra.NoFileCompletions)
	ServeCmd.Flags().StringVarP(&serveFormat, "format", "", "webp", "Image format. One of webp|gif|avif")
	_ = ServeCmd.RegisterFlagCompletionFunc("format", cobra.FixedCompletions(formats, cobra.ShellCompDirectiveNoFileComp))
	ServeCmd.Flags().StringVarP(&path, "path", "", "/", "Path to serve the app on")
	_ = ServeCmd.RegisterFlagCompletionFunc("path", cobra.NoFileCompletions)
	ServeCmd.Flags().IntVar(&width, "width", 64, "Set width")
	_ = ServeCmd.RegisterFlagCompletionFunc("width", cobra.NoFileCompletions)
	ServeCmd.Flags().IntVarP(&height, "height", "t", 32, "Set height")
	_ = ServeCmd.RegisterFlagCompletionFunc("height", cobra.NoFileCompletions)
	ServeCmd.Flags().BoolVarP(&output2x, "2x", "2", false, "Render at 2x resolution")
	ServeCmd.Flags().Int32VarP(
		&webpLevel,
		webpLevelFlag,
		"z",
		encode.WebPLevelDefault,
		"WebP compression level (0–9): 0 fast/large, 9 slow/small",
	)
	_ = ServeCmd.RegisterFlagCompletionFunc(webpLevelFlag, completeWebPLevel)
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
	ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
}

func serve(cmd *cobra.Command, args []string) error {
	imageFormat := loader.ImageWebP
	switch serveFormat {
	case "webp":
		imageFormat = loader.ImageWebP
		if flag := cmd.Flags().Lookup(webpLevelFlag); flag != nil && flag.Changed {
			encode.SetWebPLevel(webpLevel)
		}
	case "gif":
		imageFormat = loader.ImageGIF
	case "avif":
		imageFormat = loader.ImageAVIF
	default:
		slog.Warn("Invalid image format; defaulting to WebP.", "format", serveFormat)
	}

	s, err := server.NewServer(host, port, path, watch, args[0], width, height, maxDuration, timeout, imageFormat, configOutFile, output2x)
	if err != nil {
		return err
	}
	return s.Run(cmd.Context())
}
