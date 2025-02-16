package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"tidbyt.dev/pixlet/encode"
	"tidbyt.dev/pixlet/globals"
	"tidbyt.dev/pixlet/runtime"
	"tidbyt.dev/pixlet/tools"
)

func init() {
	ApiCmd.Flags().StringVarP(&host, "host", "i", "127.0.0.1", "Host interface for serving rendered images")
	ApiCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port for serving rendered images")
	ApiCmd.Flags().BoolVarP(&renderGif, "gif", "", false, "Generate GIF instead of WebP")
	ApiCmd.Flags().BoolVarP(&silenceOutput, "silent", "", false, "Silence print statements when rendering app")
}

var ApiCmd = &cobra.Command{
	Use:   "api",
	Short: "Run a Pixlet API server",
	Args:  cobra.MinimumNArgs(0),
	RunE:  api,
	Long: `Start an HTTP server that runs a Pixlet app in response to API requests.
	`,
}

type renderRequest struct {
	Path    string            `json:"path"`
	Config  map[string]string `json:"config"`
	Width   int               `json:"width"`
	Height  int               `json:"height"`
	Magnify int               `json:"magnify"`
}

func renderApplet(path string, config map[string]string, width, height, magnify int) ([]byte, error) {
	// check if path exists, and whether it is a directory or a file
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s: %w", path, err)
	}

	var fs fs.FS
	if info.IsDir() {
		fs = os.DirFS(path)
	} else {
		if !strings.HasSuffix(path, ".star") {
			return nil, fmt.Errorf("script file must have suffix .star: %s", path)
		}

		fs = tools.NewSingleFileFS(path)
	}

	if width > 0 {
		globals.Width = width
	}
	if height > 0 {
		globals.Height = height
	}
	if magnify == 0 {
		magnify = 1
	}

	// Remove the print function from the starlark thread if the silent flag is
	// passed.
	var opts []runtime.AppletOption
	if silenceOutput {
		opts = append(opts, runtime.WithPrintDisabled())
	}

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeoutCause(
			ctx,
			time.Duration(timeout)*time.Millisecond,
			fmt.Errorf("timeout after %d ms", timeout),
		)
		defer cancel()
	}

	applet, err := runtime.NewAppletFromFS(filepath.Base(path), fs, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load applet: %w", err)
	}

	roots, err := applet.RunWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error running script: %w", err)
	}
	screens := encode.ScreensFromRoots(roots)

	filter := func(input image.Image) (image.Image, error) {
		if magnify <= 1 {
			return input, nil
		}
		in, ok := input.(*image.RGBA)
		if !ok {
			return nil, fmt.Errorf("image not RGBA, very weird")
		}

		out := image.NewRGBA(
			image.Rect(
				0, 0,
				in.Bounds().Dx()*magnify,
				in.Bounds().Dy()*magnify),
		)
		for x := 0; x < in.Bounds().Dx(); x++ {
			for y := 0; y < in.Bounds().Dy(); y++ {
				for xx := 0; xx < magnify; xx++ {
					for yy := 0; yy < magnify; yy++ {
						out.SetRGBA(
							x*magnify+xx,
							y*magnify+yy,
							in.RGBAAt(x, y),
						)
					}
				}
			}
		}

		return out, nil
	}

	var buf []byte

	if screens.ShowFullAnimation {
		maxDuration = 0
	}

	if renderGif {
		buf, err = screens.EncodeGIF(maxDuration, filter)
	} else {
		buf, err = screens.EncodeWebP(maxDuration, filter)
	}
	if err != nil {
		return nil, fmt.Errorf("error rendering: %w", err)
	}

	return buf, nil
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

	buf, err := renderApplet(r.Path, r.Config, r.Width, r.Height, r.Magnify)
	if err != nil {
		http.Error(w, fmt.Sprintf("error rendering: %v", err), http.StatusInternalServerError)
		return
	}

	if renderGif {
		w.Header().Set("Content-Type", "image/gif")
	} else {
		w.Header().Set("Content-Type", "image/webp")
	}
	w.Write(buf)
}

func api(cmd *cobra.Command, args []string) error {
	cache := runtime.NewInMemoryCache()
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)

	addr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("listening at http://%s\n", addr)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/render", renderHandler)
	return http.ListenAndServe(addr, mux)
}
