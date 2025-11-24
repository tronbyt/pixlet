package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tronbyt/pixlet/runtime"
	"github.com/tronbyt/pixlet/server/browser"
	"github.com/tronbyt/pixlet/server/loader"
	"golang.org/x/sync/errgroup"
)

// Server provides functionality to serve Starlark over HTTP. It has
// functionality to watch a file and hot reload the browser on changes.
type Server struct {
	watcher *Watcher
	browser *browser.Browser
	loader  *loader.Loader
	watch   bool
}

// NewServer creates a new server initialized with the applet.
func NewServer(
	host string,
	port int,
	servePath string,
	watch bool,
	path string,
	configOutFile string,
	options ...loader.Option,
) (*Server, error) {
	fileChanges := make(chan bool, 100)

	// check if path exists, and whether it is a directory or a file
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat'ing %s: %w", path, err)
	}

	dir := path
	if !info.IsDir() {
		if !strings.HasSuffix(path, ".star") {
			return nil, fmt.Errorf("%w: %s", runtime.ErrStarSuffix, path)
		}

		dir = filepath.Dir(path)
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open root for %s: %w", path, err)
	}

	w := NewWatcher(dir, fileChanges)

	updatesChan := make(chan loader.Update, 100)
	l, err := loader.NewLoader(filepath.Base(path), root, watch, fileChanges, updatesChan, configOutFile, options...)
	if err != nil {
		_ = root.Close()
		return nil, err
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	b, err := browser.NewBrowser(addr, servePath, filepath.Base(path), watch, updatesChan, l, false)
	if err != nil {
		return nil, err
	}

	return &Server{
		watcher: w,
		browser: b,
		loader:  l,
		watch:   watch,
	}, nil
}

// Run serves the http server and runs forever in a blocking fashion.
func (s *Server) Run(ctx context.Context) error {
	defer s.loader.Close()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.loader.Run(ctx)
	})

	g.Go(func() error {
		return s.browser.Run(ctx)
	})

	if s.watch {
		g.Go(func() error {
			return s.watcher.Run(ctx)
		})
		s.loader.LoadApplet(make(map[string]string))
	}

	return g.Wait()
}
