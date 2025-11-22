package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/encode"
	"github.com/tronbyt/pixlet/runtime"
	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/canvas"
	"github.com/tronbyt/pixlet/server/loader"
)

func init() {
	ApiCmd.Flags().StringVarP(&host, "host", "i", "127.0.0.1", "Host interface for serving rendered images")
	_ = ApiCmd.RegisterFlagCompletionFunc("host", cobra.NoFileCompletions)
	ApiCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port for serving rendered images")
	_ = ApiCmd.RegisterFlagCompletionFunc("port", cobra.NoFileCompletions)
	ApiCmd.Flags().StringVarP(&imageOutputFormat, "format", "", "webp", "Output format. One of webp|gif|avif")
	_ = ApiCmd.RegisterFlagCompletionFunc("format", cobra.FixedCompletions(formats, cobra.ShellCompDirectiveNoFileComp))
	_ = ApiCmd.RegisterFlagCompletionFunc("format", cobra.NoFileCompletions)
	ApiCmd.Flags().BoolVarP(&silenceOutput, "silent", "", false, "Silence print statements when rendering app")
}

var imageFormat loader.ImageFormat

var ApiCmd = &cobra.Command{
	Use:   "api",
	Short: "Run a Pixlet API server",
	Args:  cobra.MinimumNArgs(0),
	RunE:  api,
	Long: `Start an HTTP server that runs a Pixlet app in response to API requests.
	`,
	ValidArgsFunction: cobra.NoFileCompletions,
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
func renderHandler(w http.ResponseWriter, req *http.Request) {
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

	buf, _, err := loader.RenderApplet(r.Path, r.Config, meta, maxDuration, timeout, imageFormat, silenceOutput, filters)
	if err != nil {
		http.Error(w, fmt.Sprintf("error rendering: %v", err), http.StatusInternalServerError)
		return
	}

	switch imageFormat {
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

func api(cmd *cobra.Command, args []string) error {
	cache := runtime.NewInMemoryCache()
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)

	imageFormat = loader.ImageWebP
	switch imageOutputFormat {
	case "webp":
		imageFormat = loader.ImageWebP
	case "gif":
		imageFormat = loader.ImageGIF
	case "avif":
		imageFormat = loader.ImageAVIF
	default:
		slog.Warn("Invalid image format; defaulting to WebP.", "format", imageOutputFormat)
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	slog.Info("Starting HTTP server", "address", addr)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/render", renderHandler)
	return http.ListenAndServe(addr, mux)
}
