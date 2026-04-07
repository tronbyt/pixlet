// Package loader provides primitives to load an applet both when the underlying
// file changes and on demand when an update is requested.
package loader

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/tronbyt/pixlet/encode"
	"github.com/tronbyt/pixlet/runtime"
	"github.com/tronbyt/pixlet/runtime/modules/i18n_runtime"
	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/canvas"
	"github.com/tronbyt/pixlet/schema"
	"go.starlark.net/starlark"
	"golang.org/x/text/language"
)

type metaState struct {
	is2x    bool
	loc     *time.Location
	langTag language.Tag
}

type metaUpdate struct {
	is2x *bool
	loc  *time.Location
	lang *language.Tag
	resp chan metaState
}

// Loader is a structure to provide applet loading when a file changes or on
// demand.
type Loader struct {
	root             *os.Root
	conf             *RenderConfig
	fileChanges      chan struct{}
	watch            bool
	applet           runtime.Applet
	configChanges    chan map[string]any
	requestedChanges chan struct{}
	updatesChan      chan Update
	resultsChan      chan Update
	initialLoad      chan struct{}
	configOutFile    string
	metaUpdates      chan metaUpdate
}

type Update struct {
	canvas.Metadata

	Image     string
	ImageType string
	Schema    string
	Err       error
}

// NewLoader instantiates a new loader structure. The loader will read off of
// fileChanges channel and write updates to the updatesChan. Updates are base64
// encoded WebP strings. If watch is enabled, both file changes and on demand
// requests will send updates over the updatesChan.
func NewLoader(
	id string,
	root *os.Root,
	watch bool,
	fileChanges chan struct{},
	updatesChan chan Update,
	configOutFile string,
	options ...Option,
) (*Loader, error) {
	conf := NewRenderConfig(id, nil, options...)
	l := &Loader{
		conf:             conf,
		root:             root,
		fileChanges:      fileChanges,
		watch:            watch,
		applet:           runtime.Applet{},
		updatesChan:      updatesChan,
		configChanges:    make(chan map[string]any, 100),
		requestedChanges: make(chan struct{}, 100),
		resultsChan:      make(chan Update, 100),
		initialLoad:      make(chan struct{}),
		configOutFile:    configOutFile,
		metaUpdates:      make(chan metaUpdate, 10),
	}

	if err := l.loadApplet(); err != nil {
		slog.Error("Loading applet", "error", err)
	}

	return l, nil
}

// Run executes the main loop. If there are config changes, those are recorded.
// If there is an on-demand request, it's processed and sent back to the caller
// and sent out as an update. If there is a file change, we update the applet
// and send out the update over the updatesChan.
func (l *Loader) Run(ctx context.Context) error {
	config := make(map[string]any)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case c := <-l.configChanges:
			config = c
		case mu := <-l.metaUpdates:
			if mu.is2x != nil {
				l.conf.Meta.Is2x = *mu.is2x
			}
			if mu.loc != nil {
				l.conf.Location = mu.loc
			}
			if mu.lang != nil {
				l.conf.Language = *mu.lang
			}
			if mu.resp != nil {
				mu.resp <- metaState{
					is2x:    l.conf.Meta.Is2x,
					loc:     l.conf.Location,
					langTag: l.conf.Language,
				}
			}
		case <-l.requestedChanges:
			up := Update{}

			byteSlice, err := json.Marshal(config)
			if err != nil {
				panic(err)
			}

			if l.configOutFile != "" {
				// Write the byte slice to the file.
				slog.Debug("Writing to config file", "path", l.configOutFile)
				err = os.WriteFile(l.configOutFile, byteSlice, 0644)
				if err != nil {
					panic(err)
				}
			}

			img, err := l.renderApplet(ctx, config)
			if err != nil {
				slog.Error("Loading applet", "error", err)
				up.Err = err
			} else {
				up.Image = img
				up.ImageType = l.conf.ImageFormat.String()
			}
			up.Metadata = l.conf.Meta

			l.resultsChan <- up
		case <-l.fileChanges:
			slog.Info("Detected updates; reloading")
			up := Update{}

			img, err := l.renderApplet(ctx, config)
			if err != nil {
				slog.Error("Loading applet", "error", err)
				up.Err = err
			} else {
				up.Image = img
				up.ImageType = l.conf.ImageFormat.String()
				up.Schema = string(l.GetSchema())
			}
			up.Metadata = l.conf.Meta

			l.updatesChan <- up
		}
	}
}

func (l *Loader) Close() error {
	return l.applet.Close()
}

func (l *Loader) SetIs2x(is2x bool) {
	if l.conf.Meta.Is2x == is2x {
		return
	}
	resp := make(chan metaState, 1)
	l.metaUpdates <- metaUpdate{
		is2x: &is2x,
		resp: resp,
	}
	<-resp
}

func (l *Loader) SetLocale(loc string) error {
	tag := i18n_runtime.DefaultLanguage
	if loc != "" {
		parsed, err := language.Parse(loc)
		if err != nil {
			return err
		}
		tag = parsed
	}
	resp := make(chan metaState, 1)
	l.metaUpdates <- metaUpdate{
		lang: &tag,
		resp: resp,
	}
	<-resp
	return nil
}

func (l *Loader) SetTimezone(tz string) error {
	var loc *time.Location
	if tz == "" {
		loc = time.Local
	} else {
		loaded, err := time.LoadLocation(tz)
		if err != nil {
			return err
		}
		loc = loaded
	}
	resp := make(chan metaState, 1)
	l.metaUpdates <- metaUpdate{
		loc:  loc,
		resp: resp,
	}
	<-resp
	return nil
}

func (l *Loader) Locale() language.Tag {
	return l.conf.Language
}

func (l *Loader) Location() *time.Location {
	return l.conf.Location
}

// LoadApplet loads the applet on demand.
//
// TODO: This method is thread safe, but has a pretty glaring race condition. If
// two callers request an update at the same time, they have the potential to
// get each others update. At the time of writing, this method is only called
// when you refresh a webpage during app development - so it doesn't seem likely
// that it's going to cause issues in the short term.
func (l *Loader) LoadApplet(config map[string]any) (string, error) {
	l.configChanges <- config
	l.requestedChanges <- struct{}{}
	result := <-l.resultsChan
	return result.Image, result.Err
}

func (l *Loader) GetSchema() []byte {
	<-l.initialLoad

	s := l.applet.Schema
	if s == nil {
		s = &schema.Schema{}
	}

	b, err := json.Marshal(s)
	if err != nil {
		slog.Error("Marshalling schema", "error", err)
	}
	return b
}

func (l *Loader) CallSchemaHandler(ctx context.Context, config map[string]any, handlerName, parameter string) (string, error) {
	<-l.initialLoad
	return l.applet.CallSchemaHandler(ctx, handlerName, parameter, config)
}

func (l *Loader) loadApplet() error {
	opts := []runtime.AppletOption{
		runtime.WithCanvasMeta(l.conf.Meta),
		runtime.WithLocation(l.conf.Location),
		runtime.WithLanguage(l.conf.Language),
	}

	app, err := runtime.NewAppletFromRoot(context.Background(), l.root, l.conf.Path, opts...)
	l.markInitialLoadComplete()
	if err != nil {
		return err
	}

	_ = l.applet.Close()
	l.applet = *app
	return nil
}

func (l *Loader) GetMainFile() string {
	return l.applet.MainFile
}

func (l *Loader) renderApplet(ctx context.Context, config map[string]any) (string, error) {
	if l.watch {
		if err := l.loadApplet(); err != nil {
			return "", err
		}
	}

	l.conf.Config = config

	img, err := renderApplet(ctx, &l.applet, l.conf)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(img), nil
}

func (l *Loader) markInitialLoadComplete() {
	// safely close the l.initialLoad channel to signal that the initial load is complete
	select {
	case <-l.initialLoad:
	default:
		close(l.initialLoad)
	}
}

func (l *Loader) Meta() canvas.Metadata {
	return l.conf.Meta
}

func RenderApplet(ctx context.Context, path string, config map[string]any, options ...Option) ([]byte, []string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to stat file: %w", err)
	}

	dir := path
	if !info.IsDir() {
		dir = filepath.Dir(path)
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open root: %w", err)
	}
	defer func() { _ = root.Close() }()

	return RenderAppletRoot(ctx, root, filepath.Base(path), config, options...)
}

func RenderAppletRoot(ctx context.Context, root *os.Root, path string, config map[string]any, options ...Option) ([]byte, []string, error) {
	conf := NewRenderConfig(path, config, options...)

	opts := []runtime.AppletOption{
		runtime.WithCanvasMeta(conf.Meta),
		runtime.WithLocation(conf.Location),
		runtime.WithLanguage(conf.Language),
	}

	var output []string
	if conf.SilenceOutput {
		// Replace the print function from the starlark thread if the silent flag is passed.
		opts = append(opts, runtime.WithPrintFunc(func(thread *starlark.Thread, msg string) {
			output = append(output, msg)
		}))
	}

	applet, err := runtime.NewAppletFromRoot(ctx, root, path, opts...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = applet.Close() }()

	img, err := renderApplet(ctx, applet, conf)
	if err != nil {
		return nil, output, err
	}
	return img, output, nil
}

func renderApplet(ctx context.Context, applet *runtime.Applet, conf *RenderConfig) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if conf.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeoutCause(ctx, conf.Timeout, fmt.Errorf("timeout after %s", conf.Timeout))
		defer cancel()
	}

	meta := conf.Meta
	filters := conf.Filters

	if meta.Is2x && (applet.Manifest == nil || !applet.Manifest.Supports2x) {
		meta.Is2x = false
		filters.Magnify *= 2
	}

	roots, err := applet.RunWithConfig(ctx, conf.Config)
	if err != nil {
		return nil, fmt.Errorf("error running script: %w", err)
	}

	screens := encode.ScreensFromRoots(roots, meta.ScaledWidth(), meta.ScaledHeight())

	if conf.ShowFullAnimation != nil {
		screens.ShowFullAnimation = *conf.ShowFullAnimation
	}

	filter := encode.ImageFilter(nil)
	var chain []encode.ImageFilter
	if filters.Magnify > 1 {
		chain = append(chain, encode.Magnify(filters.Magnify))
	}
	if imageFilter, err := filters.ColorFilter.ImageFilter(); err == nil && imageFilter != nil {
		chain = append(chain, imageFilter)
	}

	if len(chain) > 0 {
		filter = encode.Chain(chain...)
	}

	maxDuration := conf.MaxDuration
	if screens.ShowFullAnimation {
		maxDuration = 0
	}

	var img []byte
	switch conf.ImageFormat {
	default:
		fallthrough
	case ImageWebP:
		img, err = screens.EncodeWebP(ctx, maxDuration, filter)
	case ImageGIF:
		img, err = screens.EncodeGIF(ctx, maxDuration, filter)
	}
	if err != nil {
		return nil, fmt.Errorf("error rendering: %w", err)
	}

	return img, nil
}
