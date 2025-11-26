package cmd

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/encode"
	"github.com/tronbyt/pixlet/render"
	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/canvas"
	"github.com/tronbyt/pixlet/server"
	"github.com/tronbyt/pixlet/server/loader"
	"golang.org/x/text/language"
)

type serveOptions struct {
	host          string
	port          int
	path          string
	watch         bool
	format        string
	configOutFile string
	maxDuration   time.Duration
	timeout       time.Duration
	width         int
	height        int
	output2x      bool
	webpLevel     int32
	locale        string
}

func NewServeCmd() *cobra.Command {
	opts := &serveOptions{
		host:        "127.0.0.1",
		port:        8080,
		path:        "/",
		watch:       true,
		format:      "webp",
		maxDuration: 15 * time.Second,
		timeout:     30 * time.Second,
		width:       render.DefaultFrameWidth,
		height:      render.DefaultFrameHeight,
		webpLevel:   encode.WebPLevelDefault,
	}

	cmd := &cobra.Command{
		Use:   "serve [path]",
		Short: "Serve a Pixlet app in a web server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return serveRun(cmd, args, opts)
		},
		Long: `Serve a Pixlet app in a web server.

The path argument should be the path to the Pixlet program to run. The
program can be a single file with the .star extension, or a directory
containing multiple Starlark files and resources.`,
		ValidArgsFunction: cobra.FixedCompletions([]string{"star"}, cobra.ShellCompDirectiveFilterFileExt),
	}

	cmd.Flags().StringVarP(&opts.configOutFile, "saveconfig", "o", opts.configOutFile, "Output file for config changes")
	_ = cmd.RegisterFlagCompletionFunc("saveconfig", cobra.FixedCompletions([]string{"json"}, cobra.ShellCompDirectiveFilterFileExt))
	cmd.Flags().StringVarP(&opts.host, "host", "i", opts.host, "Host interface for serving rendered images")
	_ = cmd.RegisterFlagCompletionFunc("host", cobra.NoFileCompletions)
	cmd.Flags().IntVarP(&opts.port, "port", "p", opts.port, "Port for serving rendered images")
	_ = cmd.RegisterFlagCompletionFunc("port", cobra.NoFileCompletions)
	cmd.Flags().BoolVarP(&opts.watch, "watch", "w", opts.watch, "Reload scripts on change. Does not recurse sub-directories.")
	cmd.Flags().DurationVarP(&opts.maxDuration, "max-duration", "d", opts.maxDuration, "Maximum allowed animation duration")
	_ = cmd.RegisterFlagCompletionFunc("max-duration", cobra.NoFileCompletions)
	cmd.Flags().DurationVarP(&opts.timeout, "timeout", "", opts.timeout, "Timeout for execution")
	_ = cmd.RegisterFlagCompletionFunc("timeout", cobra.NoFileCompletions)
	cmd.Flags().StringVarP(&opts.format, "format", "", opts.format, "Image format. One of webp|gif|avif")
	_ = cmd.RegisterFlagCompletionFunc("format", cobra.FixedCompletions(formats, cobra.ShellCompDirectiveNoFileComp))
	cmd.Flags().StringVarP(&opts.path, "path", "", opts.path, "Path to serve the app on")
	_ = cmd.RegisterFlagCompletionFunc("path", cobra.NoFileCompletions)
	cmd.Flags().IntVar(&opts.width, "width", opts.width, "Set width")
	_ = cmd.RegisterFlagCompletionFunc("width", cobra.NoFileCompletions)
	cmd.Flags().IntVarP(&opts.height, "height", "t", opts.height, "Set height")
	_ = cmd.RegisterFlagCompletionFunc("height", cobra.NoFileCompletions)
	cmd.Flags().BoolVarP(&opts.output2x, "2x", "2", opts.output2x, "Render at 2x resolution (initial value for the UI toggle)")
	cmd.Flags().Int32VarP(
		&opts.webpLevel,
		webpLevelFlag,
		"z",
		encode.WebPLevelDefault,
		"WebP compression level (0–9): 0 fast/large, 9 slow/small",
	)
	_ = cmd.RegisterFlagCompletionFunc(webpLevelFlag, completeWebPLevel)
	cmd.Flags().StringVar(&opts.locale, "locale", opts.locale, "Locale to use for rendering")
	_ = cmd.RegisterFlagCompletionFunc("locale", cobra.NoFileCompletions)

	return cmd
}

func serveRun(cmd *cobra.Command, args []string, opts *serveOptions) error {
	imageFormat := loader.ImageWebP
	switch opts.format {
	case "gif":
		imageFormat = loader.ImageGIF
	case "avif":
		imageFormat = loader.ImageAVIF
	default:
		if opts.format != "webp" {
			slog.Warn("Invalid image format; defaulting to WebP.", "format", opts.format)
		}
		if flag := cmd.Flags().Lookup(webpLevelFlag); flag != nil && flag.Changed {
			encode.SetWebPLevel(opts.webpLevel)
		}
	}

	lang := language.English
	if opts.locale != "" {
		var err error
		lang, err = language.Parse(opts.locale)
		if err != nil {
			return fmt.Errorf("invalid locale: %v", err)
		}
	}

	s, err := server.NewServer(
		opts.host,
		opts.port,
		opts.path,
		opts.watch,
		args[0],
		opts.configOutFile,
		loader.WithMeta(canvas.Metadata{
			Width:  opts.width,
			Height: opts.height,
			Is2x:   opts.output2x,
		}),
		loader.WithMaxDuration(opts.maxDuration),
		loader.WithTimeout(opts.timeout),
		loader.WithImageFormat(imageFormat),
		loader.WithLanguage(lang),
	)
	if err != nil {
		return err
	}
	return s.Run(cmd.Context())
}
