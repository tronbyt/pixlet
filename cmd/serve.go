package cmd

import (
	"log/slog"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/flags"
	"github.com/tronbyt/pixlet/cmd/groups"
	"github.com/tronbyt/pixlet/encode"
	"github.com/tronbyt/pixlet/server"
	"github.com/tronbyt/pixlet/server/loader"
)

type serveOptions struct {
	host          string
	port          int
	path          string
	watch         bool
	format        string
	meta          *flags.Meta
	cache         *flags.Cache
	configOutFile string
	maxDuration   time.Duration
	timeout       time.Duration
	webpLevel     int32
	noBrowser     bool
}

func NewServeCmd() *cobra.Command {
	opts := &serveOptions{
		host:        "127.0.0.1",
		port:        8080,
		path:        "/",
		watch:       true,
		format:      loader.ImageWebP.String(),
		maxDuration: 15 * time.Second,
		timeout:     30 * time.Second,
		webpLevel:   encode.WebPLevelDefault,
		meta:        flags.NewMeta(),
		cache:       flags.NewCache(),
	}

	cmd := &cobra.Command{
		Use:     "serve [PATH]",
		GroupID: groups.Applet,
		Short:   "Serve a Pixlet app in a web server",
		Args:    cobra.MaximumNArgs(1),
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
	cmd.Flags().BoolVar(&opts.watch, "watch", opts.watch, "Reload scripts on change. Does not recurse sub-directories.")
	cmd.Flags().DurationVarP(&opts.maxDuration, "max-duration", "d", opts.maxDuration, "Maximum allowed animation duration")
	_ = cmd.RegisterFlagCompletionFunc("max-duration", cobra.NoFileCompletions)
	cmd.Flags().DurationVarP(&opts.timeout, "timeout", "", opts.timeout, "Timeout for execution")
	_ = cmd.RegisterFlagCompletionFunc("timeout", cobra.NoFileCompletions)
	cmd.Flags().StringVarP(&opts.format, "format", "", opts.format, "Image format (one of "+strings.Join(loader.ImageFormatStrings(), ", ")+")")
	_ = cmd.RegisterFlagCompletionFunc("format", cobra.FixedCompletions(loader.ImageFormatStrings(), cobra.ShellCompDirectiveNoFileComp))
	cmd.Flags().StringVarP(&opts.path, "path", "", opts.path, "Path to serve the app on")
	_ = cmd.RegisterFlagCompletionFunc("path", cobra.NoFileCompletions)
	cmd.Flags().BoolVar(&opts.noBrowser, "no-browser", false, "Don't try to open a browser")
	cmd.Flags().Int32VarP(
		&opts.webpLevel,
		webpLevelFlag,
		"z",
		encode.WebPLevelDefault,
		"WebP compression level (0–9): 0 fast/large, 9 slow/small",
	)
	_ = cmd.RegisterFlagCompletionFunc(webpLevelFlag, completeWebPLevel)

	opts.meta.Register(cmd)
	opts.cache.Register(cmd)

	return cmd
}

func serveRun(cmd *cobra.Command, args []string, opts *serveOptions) error {
	appletPath := "."
	if len(args) != 0 {
		appletPath = args[0]
	}

	imageFormat, err := loader.ImageFormatString(opts.format)
	if err != nil {
		slog.Warn("Invalid image format; defaulting to WebP.", "format", opts.format)
		imageFormat = loader.ImageWebP
	}

	if imageFormat == loader.ImageWebP {
		if flag := cmd.Flags().Lookup(webpLevelFlag); flag != nil && flag.Changed {
			encode.SetWebPLevel(opts.webpLevel)
		}
	}

	cache, err := opts.cache.Load(cmd.Context())
	if err != nil {
		return err
	}
	defer cache.Close()

	s, err := server.NewServer(
		opts.host,
		opts.port,
		opts.path,
		opts.watch,
		appletPath,
		opts.configOutFile,
		!opts.noBrowser,
		loader.WithMeta(opts.meta.Metadata),
		loader.WithMaxDuration(opts.maxDuration),
		loader.WithTimeout(opts.timeout),
		loader.WithImageFormat(imageFormat),
	)
	if err != nil {
		return err
	}
	return s.Run(cmd.Context())
}
