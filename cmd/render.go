package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/flags"
	"github.com/tronbyt/pixlet/cmd/groups"
	"github.com/tronbyt/pixlet/encode"
	"github.com/tronbyt/pixlet/server/loader"
	"golang.org/x/text/language"
)

const webpLevelFlag = "webp-level"

type renderOptions struct {
	meta  *flags.Meta
	cache *flags.Cache

	log               *slog.Logger
	configJSON        string
	output            string
	magnify           int
	imageOutputFormat string
	maxDuration       time.Duration
	silenceOutput     bool
	timeout           time.Duration
	colorFilter       string
	webpLevel         int32
	locale            string
}

func newRenderOptions() *renderOptions {
	return &renderOptions{
		meta:              flags.NewMeta(),
		cache:             flags.NewCache(),
		log:               slog.Default(),
		magnify:           1,
		imageOutputFormat: loader.ImageWebP.String(),
		maxDuration:       15 * time.Second,
		timeout:           30 * time.Second,
		webpLevel:         encode.WebPLevelDefault,
	}
}

func NewRenderCmd() *cobra.Command {
	opts := newRenderOptions()

	cmd := &cobra.Command{
		Use:     "render [PATH] [KEY=VALUE]...",
		GroupID: groups.Applet,
		Short:   "Run a Pixlet app with provided config parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return renderRun(cmd, args, opts)
		},
		Long: `Render a Pixlet app with provided config parameters.

The path argument should be the path to the Pixlet app to run. The
app can be a single file with the .star extension, or a directory
containing multiple Starlark files and resources.
	`,
		ValidArgsFunction: completeRender(opts.meta),
	}

	cmd.Flags().StringVarP(&opts.configJSON, "config", "c", opts.configJSON, "Config file in json format")
	_ = cmd.RegisterFlagCompletionFunc("config", cobra.FixedCompletions([]string{"json"}, cobra.ShellCompDirectiveFilterFileExt))

	cmd.Flags().StringVarP(&opts.output, "output", "o", opts.output, "Path for rendered image")
	_ = cmd.RegisterFlagCompletionFunc("output", cobra.FixedCompletions(loader.ImageFormatStrings(), cobra.ShellCompDirectiveFilterFileExt))

	cmd.Flags().StringVarP(&opts.imageOutputFormat, "format", "", opts.imageOutputFormat, "Output format (one of "+strings.Join(loader.ImageFormatStrings(), ", ")+")")
	_ = cmd.RegisterFlagCompletionFunc("format", cobra.FixedCompletions(loader.ImageFormatStrings(), cobra.ShellCompDirectiveNoFileComp))

	cmd.Flags().BoolVarP(&opts.silenceOutput, "silent", "", opts.silenceOutput, "Silence print statements when rendering app")
	cmd.Flags().IntVarP(
		&opts.magnify,
		"magnify",
		"m",
		opts.magnify,
		"Increase image dimension by a factor (useful for debugging)",
	)
	_ = cmd.RegisterFlagCompletionFunc("magnify", cobra.NoFileCompletions)

	cmd.Flags().StringVar(
		&opts.colorFilter,
		"color-filter",
		opts.colorFilter,
		`Apply a color filter. (See "pixlet community list-color-filters")`,
	)
	_ = cmd.RegisterFlagCompletionFunc("color-filter", func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		s := make([]string, 0, len(encode.ColorFilterValues()))
		for _, v := range encode.ColorFilterValues() {
			desc, _ := v.Description()
			s = append(s, v.String()+"\t"+desc)
		}
		return s, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.Flags().DurationVarP(
		&opts.maxDuration,
		"max-duration",
		"d",
		opts.maxDuration,
		"Maximum allowed animation duration",
	)
	_ = cmd.RegisterFlagCompletionFunc("max-duration", cobra.NoFileCompletions)

	cmd.Flags().DurationVarP(
		&opts.timeout,
		"timeout",
		"",
		opts.timeout,
		"Timeout for execution",
	)
	_ = cmd.RegisterFlagCompletionFunc("timeout", cobra.NoFileCompletions)

	cmd.Flags().Int32VarP(
		&opts.webpLevel,
		webpLevelFlag,
		"z",
		opts.webpLevel,
		"WebP compression level (0–9): 0 fast/large, 9 slow/small",
	)
	_ = cmd.RegisterFlagCompletionFunc(webpLevelFlag, completeWebPLevel)

	cmd.Flags().StringVar(&opts.locale, "locale", opts.locale, "Locale to use for rendering")
	_ = cmd.RegisterFlagCompletionFunc("locale", cobra.NoFileCompletions)

	opts.meta.Register(cmd)
	opts.cache.Register(cmd)

	return cmd
}

func renderRun(cmd *cobra.Command, args []string, opts *renderOptions) error {
	path, config, _, err := loadConfig(opts.configJSON, args)
	if err != nil {
		return err
	}

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

	imageFormat, err := loader.ImageFormatString(opts.imageOutputFormat)
	if err != nil {
		opts.log.Warn("Invalid image format; defaulting to WebP.", "format", opts.imageOutputFormat)
		imageFormat = loader.ImageWebP
	}

	if imageFormat == loader.ImageWebP {
		if flag := cmd.Flags().Lookup(webpLevelFlag); flag != nil && flag.Changed {
			encode.SetWebPLevel(opts.webpLevel)
		}
	}

	if opts.output != "" {
		outPath = opts.output
	} else {
		if opts.meta.Is2x {
			outPath += "@2x"
		}
		outPath += "." + imageFormat.String()
	}

	cache, err := opts.cache.Load(cmd.Context())
	if err != nil {
		return err
	}
	defer cache.Close()

	filters := encode.RenderFilters{Magnify: opts.magnify}
	if opts.colorFilter != "" {
		if filters.ColorFilter, err = encode.ColorFilterString(opts.colorFilter); err != nil {
			return err
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

	buf, _, err := loader.RenderApplet(
		cmd.Context(),
		path,
		config,
		loader.WithMeta(opts.meta.Metadata),
		loader.WithMaxDuration(opts.maxDuration),
		loader.WithTimeout(opts.timeout),
		loader.WithImageFormat(imageFormat),
		loader.WithSilenceOutput(opts.silenceOutput),
		loader.WithLanguage(lang),
		loader.WithFilters(filters),
	)
	if err != nil {
		return fmt.Errorf("error rendering: %w", err)
	}

	if outPath == "-" {
		if _, err := os.Stdout.Write(buf); err != nil {
			return fmt.Errorf("writing to stdout: %w", err)
		}
	} else {
		if err := os.WriteFile(outPath, buf, 0644); err != nil {
			return fmt.Errorf("writing to file: %w", err)
		}

		if wd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(wd, outPath); err == nil {
				outPath = rel
			}
		}
	}

	opts.log.Info("Rendered image", "path", outPath)
	return nil
}

func loadConfig(configPath string, args []string) (string, map[string]any, []string, error) {
	starPath := "."
	if len(args) != 0 {
		if !strings.Contains(args[0], "=") {
			starPath = args[0]
			args = args[1:]
		} else if _, err := os.Stat(args[0]); err == nil || !errors.Is(err, os.ErrNotExist) {
			starPath = args[0]
			args = args[1:]
		}
	}

	starPath, err := filepath.Abs(starPath)
	if err != nil {
		return "", nil, args, fmt.Errorf("failed to get absolute path for %s: %w", starPath, err)
	}

	config := map[string]any{}

	if configPath != "" {
		f, err := os.Open(configPath)
		if err != nil {
			return "", nil, args, fmt.Errorf("file open error: %w", err)
		}
		defer func() { _ = f.Close() }()

		err = json.NewDecoder(f).Decode(&config)
		if err != nil {
			return "", nil, args, fmt.Errorf("failed to unmarshal JSON %v: %w", configPath, err)
		}
	}

	for _, param := range args {
		key, val, ok := strings.Cut(param, "=")
		if !ok {
			return "", nil, args, fmt.Errorf("parameters must be in form <key>=<value>, found %s", param)
		}
		config[key] = val
	}

	return starPath, config, args, nil
}
