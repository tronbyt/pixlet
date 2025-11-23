package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/encode"
	"github.com/tronbyt/pixlet/runtime"
	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/canvas"
	"github.com/tronbyt/pixlet/server/loader"
)

type apiOptions struct {
	host          string
	port          int
	format        string
	silenceOutput bool
	maxDuration   time.Duration
	timeout       time.Duration
	imageFormat   loader.ImageFormat
}

func NewAPICmd() *cobra.Command {
	opts := &apiOptions{
		host:        "127.0.0.1",
		port:        8080,
		format:      "webp",
		maxDuration: 15 * time.Second,
		timeout:     30 * time.Second,
		imageFormat: loader.ImageWebP,
	}

	cmd := &cobra.Command{
		Use:   "api",
		Short: "Run a Pixlet API server",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return apiRun(cmd, args, opts)
		},
		Long: `Start an HTTP server that runs a Pixlet app in response to API requests.
	`,
		ValidArgsFunction: cobra.NoFileCompletions,
	}

	cmd.Flags().StringVarP(&opts.host, "host", "i", opts.host, "Host interface for serving rendered images")
	_ = cmd.RegisterFlagCompletionFunc("host", cobra.NoFileCompletions)
	cmd.Flags().IntVarP(&opts.port, "port", "p", opts.port, "Port for serving rendered images")
	_ = cmd.RegisterFlagCompletionFunc("port", cobra.NoFileCompletions)
	cmd.Flags().StringVarP(&opts.format, "format", "", opts.format, "Output format. One of webp|gif|avif")
	_ = cmd.RegisterFlagCompletionFunc("format", cobra.FixedCompletions(formats, cobra.ShellCompDirectiveNoFileComp))
	cmd.Flags().BoolVarP(&opts.silenceOutput, "silent", "", opts.silenceOutput, "Silence print statements when rendering app")

	return cmd
}

type renderRequest struct {
	Path        string            `json:"path"`
	Config      map[string]string `json:"config"`
	Width       int               `json:"width"`
	Height      int               `json:"height"`
	Magnify     int               `json:"magnify"`
	ColorFilter string            `json:"color_filter,omitempty"`
	Output2x    bool              `json:"2x,omitempty"`
}

func validatePath(path string) bool {
	return !strings.Contains(path, "..")
}

// Example request
//
//	{
//	   "path": "/workspaces/pixlet/examples/clock",
//	   "config": {
//	       "timezone": "America/New_York"
//	   }
//	}
func (o *apiOptions) renderHandler(w http.ResponseWriter, req *http.Request) {
	var r renderRequest

	if err := json.NewDecoder(req.Body).Decode(&r); err != nil {
		http.Error(w, fmt.Sprintf("failed to decode render request: %v", err), http.StatusBadRequest)
		return
	}

	if !validatePath(r.Path) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	filters := &encode.RenderFilters{Magnify: r.Magnify}
	if r.ColorFilter != "" {
		var err error
		if filters.ColorFilter, err = encode.ColorFilterString(r.ColorFilter); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	meta := canvas.Metadata{
		Width:  r.Width,
		Height: r.Height,
		Is2x:   r.Output2x,
	}

	buf, _, err := loader.RenderApplet(r.Path, r.Config, meta, o.maxDuration, o.timeout, o.imageFormat, o.silenceOutput, nil, filters)
	if err != nil {
		http.Error(w, fmt.Sprintf("error rendering: %v", err), http.StatusInternalServerError)
		return
	}

	switch o.imageFormat {
	default:
		fallthrough
	case loader.ImageWebP:
		w.Header().Set("Content-Type", "image/webp")
	case loader.ImageGIF:
		w.Header().Set("Content-Type", "image/gif")
	case loader.ImageAVIF:
		w.Header().Set("Content-Type", "image/avif")
	}
	w.Write(buf)
}

func apiRun(_ *cobra.Command, _ []string, opts *apiOptions) error {
	cache := runtime.NewInMemoryCache()
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)

	switch opts.format {
	case "gif":
		opts.imageFormat = loader.ImageGIF
	case "avif":
		opts.imageFormat = loader.ImageAVIF
	default:
		if opts.format != "webp" {
			slog.Warn("Invalid image format; defaulting to WebP.", "format", opts.format)
		}
	}

	addr := fmt.Sprintf("%s:%d", opts.host, opts.port)
	slog.Info("Starting HTTP server", "address", addr)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/render", opts.renderHandler)
	return http.ListenAndServe(addr, mux)
}
