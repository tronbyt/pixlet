// Package loader provides primitives to load an applet both when the underlying
// file changes and on demand when an update is requested.
package loader

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/tronbyt/pixlet/encode"
	"github.com/tronbyt/pixlet/runtime"
	"github.com/tronbyt/pixlet/runtime/modules/render_runtime"
	"github.com/tronbyt/pixlet/schema"
	"go.starlark.net/starlark"
)

type ImageFormat int

const (
	ImageWebP ImageFormat = iota
	ImageGIF
	ImageAVIF
)

// Loader is a structure to provide applet loading when a file changes or on
// demand.
type Loader struct {
	id               string
	fs               fs.FS
	fileChanges      chan bool
	watch            bool
	applet           runtime.Applet
	configChanges    chan map[string]string
	requestedChanges chan bool
	updatesChan      chan Update
	resultsChan      chan Update
	maxDuration      int
	initialLoad      chan bool
	timeout          int
	imageFormat      ImageFormat
	configOutFile    string
	width            int
	height           int
	output2x         bool
}

type Update struct {
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
	fs fs.FS,
	watch bool,
	fileChanges chan bool,
	updatesChan chan Update,
	width, height, maxDuration int,
	timeout int,
	imageFormat ImageFormat,
	configOutFile string,
	output2x bool,
) (*Loader, error) {
	l := &Loader{
		id:               id,
		fs:               fs,
		fileChanges:      fileChanges,
		watch:            watch,
		applet:           runtime.Applet{},
		updatesChan:      updatesChan,
		configChanges:    make(chan map[string]string, 100),
		requestedChanges: make(chan bool, 100),
		resultsChan:      make(chan Update, 100),
		maxDuration:      maxDuration,
		initialLoad:      make(chan bool),
		timeout:          timeout,
		imageFormat:      imageFormat,
		configOutFile:    configOutFile,
		width:            width,
		height:           height,
		output2x:         output2x,
	}

	cache := runtime.NewInMemoryCache()
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)

	if !l.watch {
		if err := l.loadApplet(); err != nil {
			return nil, err
		}
	}

	return l, nil
}

// Run executes the main loop. If there are config changes, those are recorded.
// If there is an on-demand request, it's processed and sent back to the caller
// and sent out as an update. If there is a file change, we update the applet
// and send out the update over the updatesChan.
func (l *Loader) Run() error {
	config := make(map[string]string)

	for {
		select {
		case c := <-l.configChanges:
			config = c
		case <-l.requestedChanges:
			up := Update{}

			byteSlice, err := json.Marshal(config)
			if err != nil {
				panic(err)
			}

			if l.configOutFile != "" {
				// Write the byte slice to the file.
				//log.Printf("writing to %v",l.configOutFile)
				err = os.WriteFile(l.configOutFile, byteSlice, 0644)
				if err != nil {
					panic(err)
				}
			}

			img, err := l.renderApplet(config)
			if err != nil {
				log.Printf("error loading applet: %v", err)
				up.Err = err
			} else {
				up.Image = img
				switch l.imageFormat {
				default:
					fallthrough
				case ImageWebP:
					up.ImageType = "webp"
				case ImageGIF:
					up.ImageType = "gif"
				case ImageAVIF:
					up.ImageType = "avif"
				}
			}

			l.updatesChan <- up
			l.resultsChan <- up
		case <-l.fileChanges:
			log.Println("detected updates, reloading")
			up := Update{}

			img, err := l.renderApplet(config)
			if err != nil {
				log.Printf("error loading applet: %v", err)
				up.Err = err
			} else {
				up.Image = img
				switch l.imageFormat {
				default:
					fallthrough
				case ImageWebP:
					up.ImageType = "webp"
				case ImageGIF:
					up.ImageType = "gif"
				case ImageAVIF:
					up.ImageType = "avif"
				}
				up.Schema = string(l.applet.SchemaJSON)
			}

			l.updatesChan <- up
		}
	}
}

// LoadApplet loads the applet on demand.
//
// TODO: This method is thread safe, but has a pretty glaring race condition. If
// two callers request an update at the same time, they have the potential to
// get each others update. At the time of writing, this method is only called
// when you refresh a webpage during app development - so it doesn't seem likely
// that it's going to cause issues in the short term.
func (l *Loader) LoadApplet(config map[string]string) (string, error) {
	l.configChanges <- config
	l.requestedChanges <- true
	result := <-l.resultsChan
	return result.Image, result.Err
}

func (l *Loader) GetSchema() []byte {
	<-l.initialLoad

	s := l.applet.SchemaJSON
	if len(s) > 0 {
		return s
	}

	b, _ := json.Marshal(&schema.Schema{})
	return b
}

func (l *Loader) CallSchemaHandler(ctx context.Context, config map[string]string, handlerName, parameter string) (string, error) {
	<-l.initialLoad
	return l.applet.CallSchemaHandler(ctx, handlerName, parameter, config)
}

func (l *Loader) loadApplet() error {
	opts := []runtime.AppletOption{
		runtime.WithMetadata(render_runtime.Metadata{
			Width:  l.width,
			Height: l.height,
			Is2x:   l.output2x,
		}),
	}

	app, err := runtime.NewAppletFromFS(l.id, l.fs, opts...)
	l.markInitialLoadComplete()
	if err != nil {
		return err
	}

	l.applet = *app
	return nil
}

func (l *Loader) renderApplet(config map[string]string) (string, error) {
	if l.watch {
		if err := l.loadApplet(); err != nil {
			return "", err
		}
	}

	ctx, _ := context.WithTimeoutCause(
		context.Background(),
		time.Duration(l.timeout)*time.Millisecond,
		fmt.Errorf("timeout after %dms", l.timeout),
	)

	roots, err := l.applet.RunWithConfig(ctx, config)
	if err != nil {
		return "", fmt.Errorf("error running script: %w", err)
	}

	width, height := l.width, l.height
	magnify := 1

	if l.output2x {
		if l.applet.Manifest != nil && l.applet.Manifest.Supports2x {
			width *= 2
			height *= 2
		} else {
			magnify = 2
		}
	}

	screens := encode.ScreensFromRoots(roots, width, height)

	var chain []encode.ImageFilter
	if magnify > 1 {
		chain = append(chain, encode.Magnify(magnify))
	}
	filter := encode.Chain(chain...)

	maxDuration := l.maxDuration
	if screens.ShowFullAnimation {
		maxDuration = 0
	}

	var img []byte
	switch l.imageFormat {
	default:
		fallthrough
	case ImageWebP:
		img, err = screens.EncodeWebP(maxDuration, filter)
	case ImageGIF:
		img, err = screens.EncodeGIF(maxDuration, filter)
	case ImageAVIF:
		img, err = screens.EncodeAVIF(maxDuration, filter)
	}
	if err != nil {
		return "", fmt.Errorf("error rendering: %w", err)
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

func RenderApplet(path string, config map[string]string, width, height, magnify, maxDuration, timeout int, imageFormat ImageFormat, silenceOutput bool, filters *encode.RenderFilters) ([]byte, []string, error) {
	if filters == nil {
		filters = &encode.RenderFilters{}
	}
	if filters.Magnify == 0 {
		filters.Magnify = magnify
	}

	opts := []runtime.AppletOption{
		runtime.WithMetadata(render_runtime.Metadata{
			Width:  width,
			Height: height,
			Is2x:   filters.Output2x,
		}),
	}

	var output []string
	if silenceOutput {
		// Replace the print function from the starlark thread if the silent flag is passed.
		opts = append(opts, runtime.WithPrintFunc(func(thread *starlark.Thread, msg string) {
			output = append(output, msg)
		}))
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

	applet, err := runtime.NewAppletFromPath(path, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load applet: %w", err)
	}

	roots, err := applet.RunWithConfig(ctx, config)
	if err != nil {
		return nil, output, fmt.Errorf("error running script: %w", err)
	}

	if filters.Output2x {
		if applet.Manifest != nil && applet.Manifest.Supports2x {
			width *= 2
			height *= 2
		} else {
			if filters.Magnify == 0 {
				filters.Magnify = 1
			}
			filters.Magnify *= 2
		}
	}

	screens := encode.ScreensFromRoots(roots, width, height)

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

	var buf []byte

	if screens.ShowFullAnimation {
		maxDuration = 0
	}

	switch imageFormat {
	default:
		fallthrough
	case ImageWebP:
		buf, err = screens.EncodeWebP(maxDuration, filter)
	case ImageGIF:
		buf, err = screens.EncodeGIF(maxDuration, filter)
	case ImageAVIF:
		buf, err = screens.EncodeAVIF(maxDuration, filter)
	}
	if err != nil {
		return nil, output, fmt.Errorf("error rendering: %w", err)
	}

	return buf, output, nil
}

func (l *Loader) Width() int {
	if l.output2x {
		return l.width * 2
	}
	return l.width
}

func (l *Loader) Height() int {
	if l.output2x {
		return l.height * 2
	}
	return l.height
}
