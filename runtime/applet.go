package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	starlibbsoup "github.com/qri-io/starlib/bsoup"
	starlibgzip "github.com/qri-io/starlib/compress/gzip"
	starlibbase64 "github.com/qri-io/starlib/encoding/base64"
	starlibcsv "github.com/qri-io/starlib/encoding/csv"
	starlibhash "github.com/qri-io/starlib/hash"
	starlibhtml "github.com/qri-io/starlib/html"
	starlibre "github.com/qri-io/starlib/re"
	starlibzip "github.com/qri-io/starlib/zipfile"
	"github.com/tronbyt/pixlet/manifest"
	"github.com/tronbyt/pixlet/render"
	"github.com/tronbyt/pixlet/runtime/modules/animation_runtime"
	"github.com/tronbyt/pixlet/runtime/modules/encoding/yaml"
	"github.com/tronbyt/pixlet/runtime/modules/file"
	"github.com/tronbyt/pixlet/runtime/modules/hmac"
	"github.com/tronbyt/pixlet/runtime/modules/humanize"
	"github.com/tronbyt/pixlet/runtime/modules/i18n_runtime"
	"github.com/tronbyt/pixlet/runtime/modules/qrcode"
	"github.com/tronbyt/pixlet/runtime/modules/random"
	"github.com/tronbyt/pixlet/runtime/modules/render_runtime"
	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/canvas"
	"github.com/tronbyt/pixlet/runtime/modules/starlarkhttp"
	"github.com/tronbyt/pixlet/runtime/modules/sunrise"
	"github.com/tronbyt/pixlet/runtime/modules/time_runtime"
	"github.com/tronbyt/pixlet/runtime/modules/xpath"
	"github.com/tronbyt/pixlet/schema"
	"github.com/tronbyt/pixlet/starlarkutil"
	"github.com/tronbyt/pixlet/tools/iterutil"
	starlibjson "go.starlark.net/lib/json"
	starlibmath "go.starlark.net/lib/math"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/starlarktest"
	"go.starlark.net/syntax"
	"golang.org/x/mod/semver"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
)

const (
	ManifestPath = "manifest.yaml"
	LocaleDir    = "locales"
)

type ModuleLoader func(*starlark.Thread, string) (starlark.StringDict, error)

type PrintFunc func(thread *starlark.Thread, msg string)

type AppletOption func(*Applet) error

// ThreadInitializer is called when building a Starlark thread to run an applet
// on. It can customize the thread by overriding behavior or attaching thread
// local data.
type ThreadInitializer func(thread *starlark.Thread) *starlark.Thread

type Applet struct {
	ID       string
	Manifest *manifest.Manifest
	Globals  map[string]starlark.StringDict
	MainFile string
	Root     *os.Root

	loader       ModuleLoader
	initializers []ThreadInitializer
	loadedPaths  map[string]bool

	mainFun    *starlark.Function
	schemaFile string

	Schema *schema.Schema
}

func WithModuleLoader(loader ModuleLoader) AppletOption {
	return func(a *Applet) error {
		a.loader = loader
		return nil
	}
}

func WithThreadInitializer(init ThreadInitializer) AppletOption {
	return func(a *Applet) error {
		a.initializers = append(a.initializers, init)
		return nil
	}
}

func WithSecretDecryptionKey(key *SecretDecryptionKey) AppletOption {
	return func(a *Applet) error {
		if decrypter, err := key.decrypterForApp(a); err != nil {
			return fmt.Errorf("preparing secret key: %w", err)
		} else {
			a.initializers = append(a.initializers, func(t *starlark.Thread) *starlark.Thread {
				decrypter.attachToThread(t)
				return t
			})
			return nil
		}
	}
}

func WithPrintFunc(print PrintFunc) AppletOption {
	return func(a *Applet) error {
		a.initializers = append(a.initializers, func(t *starlark.Thread) *starlark.Thread {
			t.Print = print
			return t
		})
		return nil
	}
}

func WithPrintDisabled() AppletOption {
	return WithPrintFunc(func(thread *starlark.Thread, msg string) {})
}

func WithCanvasMeta(m canvas.Metadata) AppletOption {
	return func(a *Applet) error {
		a.initializers = append(a.initializers, func(t *starlark.Thread) *starlark.Thread {
			canvas.AttachToThread(t, m)
			return t
		})
		return nil
	}
}

func WithLocation(tz *time.Location) AppletOption {
	return func(a *Applet) error {
		a.initializers = append(a.initializers, func(t *starlark.Thread) *starlark.Thread {
			time_runtime.SetLocation(t, tz)
			return t
		})
		return nil
	}
}

func WithLanguage(lang language.Tag) AppletOption {
	return func(a *Applet) error {
		a.initializers = append(a.initializers, func(t *starlark.Thread) *starlark.Thread {
			i18n_runtime.AttachLanguageToThread(t, lang)
			return t
		})
		return nil
	}
}

func WithTests(t testing.TB) AppletOption {
	return func(a *Applet) error {
		a.initializers = append(a.initializers, func(thread *starlark.Thread) *starlark.Thread {
			starlarktest.SetReporter(thread, t)
			return thread
		})
		return nil
	}
}

func NewApplet(id string, src []byte, opts ...AppletOption) (*Applet, error) {
	fn := id
	if !strings.HasSuffix(fn, ".star") {
		fn += ".star"
	}

	vfs := fstest.MapFS{
		fn: &fstest.MapFile{
			Data: src,
		},
	}

	return NewAppletFromFS(id, vfs, opts...)
}

func NewAppletFromFS(id string, fsys fs.FS, opts ...AppletOption) (*Applet, error) {
	a := &Applet{
		ID:          id,
		Globals:     make(map[string]starlark.StringDict),
		loadedPaths: make(map[string]bool),
	}

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	if err := a.load(fsys); err != nil {
		return nil, err
	}

	return a, nil
}

func NewAppletFromRoot(id string, root *os.Root, opts ...AppletOption) (*Applet, error) {
	a, err := NewAppletFromFS(id, root.FS(), opts...)
	if err != nil {
		return nil, err
	}

	a.Root = root
	return a, nil
}

var ErrStarSuffix = fmt.Errorf("script file must have suffix .star")

func NewAppletFromPath(path string, opts ...AppletOption) (*Applet, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s: %w", path, err)
	}

	dir := path
	if !info.IsDir() {
		if !strings.HasSuffix(path, ".star") {
			return nil, fmt.Errorf("%w: %s", ErrStarSuffix, path)
		}

		dir = filepath.Dir(path)
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open root for %s: %w", path, err)
	}

	a, err := NewAppletFromRoot(filepath.Base(path), root, opts...)
	if err != nil {
		_ = root.Close()
		return nil, err
	}

	return a, nil
}

func (a *Applet) Close() error {
	if a.Root == nil {
		return nil
	}
	return a.Root.Close()
}

// Run executes the applet's main function. It returns the render roots that are
// returned by the applet.
func (a *Applet) Run(ctx context.Context) (roots []render.Root, err error) {
	return a.RunWithConfig(ctx, nil)
}

// ExtractRoots extracts render roots from a Starlark value. It expects the value
// to be either a single render root or a list of render roots.
//
// It's used internally by RunWithConfig to extract the roots returned by the applet.
func ExtractRoots(val starlark.Value) ([]render.Root, error) {
	var roots []render.Root

	if val == starlark.None {
		// no roots returned
	} else if returnRoot, ok := val.(render_runtime.Rootable); ok {
		roots = []render.Root{returnRoot.AsRenderRoot()}
	} else if returnList, ok := val.(*starlark.List); ok {
		roots = make([]render.Root, returnList.Len())

		for i, listVal := range iterutil.Enumerate(returnList.Elements()) {
			if listValRoot, ok := listVal.(render_runtime.Rootable); ok {
				roots[i] = listValRoot.AsRenderRoot()
			} else {
				return nil, fmt.Errorf(
					"expected app implementation to return Root(s) but found: %s (at index %d)",
					listVal.Type(), i,
				)
			}
		}
	} else {
		return nil, fmt.Errorf("expected app implementation to return Root(s) but found: %s", val.Type())
	}

	return roots, nil
}

// RunWithConfig exceutes the applet's main function, passing it configuration as a
// starlark dict. It returns the render roots that are returned by the applet.
func (a *Applet) RunWithConfig(ctx context.Context, config map[string]string) (roots []render.Root, err error) {
	var args starlark.Tuple
	if a.mainFun.NumParams() > 0 {
		starlarkConfig := AppletConfig(config)
		args = starlark.Tuple{starlarkConfig}
	}

	returnValue, err := a.Call(ctx, a.mainFun, args...)
	if err != nil {
		return nil, err
	}

	roots, err = ExtractRoots(returnValue)
	if err != nil {
		return nil, err
	}

	return roots, nil
}

// CallSchemaHandler calls a schema handler, passing it a single
// string parameter and returning a single string value.
func (app *Applet) CallSchemaHandler(ctx context.Context, handlerName, parameter string, config map[string]string) (result string, err error) {
	handler, found := app.Schema.Handlers[handlerName]
	if !found {
		return "", fmt.Errorf("no exported handler named '%s'", handlerName)
	}

	args := starlark.Tuple{
		starlark.String(parameter),
	}

	if handler.Function.NumParams() > 1 {
		args = append(args, AppletConfig(config))
	}

	resultVal, err := app.Call(ctx, handler.Function, args...)
	if err != nil {
		return "", fmt.Errorf("calling schema handler %s: %v", handlerName, err)
	}

	switch handler.ReturnType {
	case schema.ReturnOptions:
		options, err := schema.EncodeOptions(resultVal)
		if err != nil {
			return "", err
		}
		return options, nil

	case schema.ReturnSchema:
		sch, err := schema.FromStarlark(resultVal, app.Globals[app.schemaFile])
		if err != nil {
			return "", err
		}

		s, err := json.Marshal(sch)
		if err != nil {
			return "", fmt.Errorf("serializing schema to JSON: %w", err)
		}

		return string(s), nil

	case schema.ReturnString:
		str, ok := starlark.AsString(resultVal)
		if !ok {
			return "", fmt.Errorf(
				"expected %s to return a string or string-like value",
				handler.Function.Name(),
			)
		}
		return str, nil
	}

	return "", fmt.Errorf("a very unexpected error happened for handler \"%s\"", handlerName)
}

// RunTests runs all test functions that are defined in the applet source.
func (app *Applet) RunTests(t *testing.T) {
	app.initializers = append(app.initializers, func(thread *starlark.Thread) *starlark.Thread {
		starlarktest.SetReporter(thread, t)
		return thread
	})

	for file, globals := range app.Globals {
		for name, global := range globals {
			if !strings.HasPrefix(name, "test_") {
				continue
			}

			if fun, ok := global.(*starlark.Function); ok {
				t.Run(fmt.Sprintf("%s/%s", file, name), func(t *testing.T) {
					if _, err := app.Call(context.Background(), fun); err != nil {
						t.Error(err)
					}
				})
			}
		}
	}
}

// Calls any callable from Applet.Globals. Pass args and receive a
// starlark Value, or an error if you're unlucky.
func (a *Applet) Call(ctx context.Context, callable *starlark.Function, args ...starlark.Value) (val starlark.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while running %s: %v\n%s", a.ID, r, debug.Stack())
		}
	}()

	t := a.newThread(ctx)
	defer starlarkutil.RunOnExitFuncs(t)

	context.AfterFunc(ctx, func() {
		t.Cancel(context.Cause(ctx).Error())
	})

	resultVal, err := starlark.Call(t, callable, args, nil)
	if err != nil {
		evalErr, ok := err.(*starlark.EvalError)
		if ok {
			return nil, fmt.Errorf("%s", evalErr.Backtrace())
		}
		return nil, fmt.Errorf(
			"in %s at %s: %s",
			callable.Name(),
			callable.Position().String(),
			err,
		)
	}

	return resultVal, nil
}

// PathsForBundle returns a list of all the paths that have been loaded by the
// applet. This is useful for creating a bundle of the applet.
func (a *Applet) PathsForBundle() []string {
	paths := make([]string, 0, len(a.loadedPaths))
	for path := range a.loadedPaths {
		paths = append(paths, path)
	}
	return paths
}

func (a *Applet) load(fsys fs.FS) (err error) {
	// list files in the root directory of fsys
	rootDir, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return fmt.Errorf("reading root directory: %v", err)
	}

	if err := a.loadManifest(fsys, ManifestPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	singleFile := path.Ext(a.ID) == ".star"

	for _, d := range rootDir {
		switch {
		case d.IsDir():
			if d.Name() == LocaleDir {
				if err := a.loadCatalog(fsys); err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}
			}
		case path.Ext(d.Name()) == ".star":
			if singleFile && a.ID != d.Name() {
				// Skip loading other star files when in single-file mode
				continue
			}
			if err := a.ensureLoaded(fsys, d.Name()); err != nil {
				return err
			}
		}
	}

	if a.mainFun == nil {
		return fmt.Errorf("no main() function found in %s", a.ID)
	}

	return nil
}

var (
	ErrMinPixletVersionInvalid = errors.New("manifest minPixletVersion is not a valid version string")
	ErrMinPixletVersion        = errors.New("app requires a newer pixlet version")
)

func (a *Applet) loadManifest(fsys fs.FS, pathToLoad string) (err error) {
	r, err := fsys.Open(pathToLoad)
	if err != nil {
		return fmt.Errorf("opening %s: %w", pathToLoad, err)
	}
	defer r.Close()

	a.Manifest, err = manifest.LoadManifest(r)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	if Version != "" && a.Manifest.MinPixletVersion != "" && semver.IsValid(Version) {
		if !semver.IsValid(a.Manifest.MinPixletVersion) {
			return fmt.Errorf("%w: %s", ErrMinPixletVersionInvalid, a.Manifest.MinPixletVersion)
		}

		if semver.Compare(Version, a.Manifest.MinPixletVersion) < 0 {
			return fmt.Errorf("%w: needs %s, got %s", ErrMinPixletVersion, a.Manifest.MinPixletVersion, Version)
		}
	}

	a.initializers = append(a.initializers, func(t *starlark.Thread) *starlark.Thread {
		manifest.AttachToThread(t, a.Manifest)
		return t
	})

	return nil
}

func (a *Applet) ensureLoaded(fsys fs.FS, pathToLoad string, currentlyLoading ...string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while executing %s: %v\n%s", a.ID, r, debug.Stack())
		}
	}()

	// normalize path so that it can be used as a key
	pathToLoad = path.Clean(pathToLoad)
	if _, ok := a.Globals[pathToLoad]; ok {
		// already loaded, good to go
		return nil
	}

	// use the currentlyLoading slice to detect circular dependencies
	if slices.Contains(currentlyLoading, pathToLoad) {
		return fmt.Errorf("circular dependency detected: %s -> %s", strings.Join(currentlyLoading, " -> "), pathToLoad)
	} else {
		// mark this file as currently loading. if we encounter it again,
		// we have a circular dependency.
		currentlyLoading = append(currentlyLoading, pathToLoad)

		// also mark the file as loaded to keep track of all of the files
		// that have been loaded
		a.loadedPaths[pathToLoad] = true
	}

	src, err := fs.ReadFile(fsys, pathToLoad)
	if err != nil {
		return fmt.Errorf("reading %s: %v", pathToLoad, err)
	}

	predeclared := starlark.StringDict{
		"struct": starlark.NewBuiltin("struct", starlarkstruct.Make),
	}

	thread := a.newThread(context.Background())
	defer starlarkutil.RunOnExitFuncs(thread)

	// override loader to allow loading starlark files
	thread.Load = func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
		// normalize module path
		modulePath := path.Clean(module)

		// if the module exists on the filesystem, load it
		if _, err := fs.Stat(fsys, modulePath); err == nil {
			// ensure the module is loaded, and pass the currentlyLoading slice
			// to detect circular dependencies
			if err := a.ensureLoaded(fsys, modulePath, currentlyLoading...); err != nil {
				return nil, err
			}

			if g, ok := a.Globals[modulePath]; !ok {
				return nil, fmt.Errorf("module %s not loaded", modulePath)
			} else {
				return g, nil
			}
		}

		// fallback to default loader
		return a.loadModule(thread, module)
	}

	switch path.Ext(pathToLoad) {
	case ".star":
		globals, err := starlark.ExecFileOptions(
			&syntax.FileOptions{
				Set:       true,
				Recursion: true,
			},
			thread,
			path.Join(a.ID, pathToLoad),
			src,
			predeclared,
		)
		if err != nil {
			return fmt.Errorf("starlark.ExecFile: %v", err)
		}
		a.Globals[pathToLoad] = globals

		// if the file is in the root directory, check for the main function
		// and schema function
		mainFun, _ := globals["main"].(*starlark.Function)
		if mainFun != nil {
			if a.MainFile != "" {
				return fmt.Errorf("multiple files with a main() function:\n- %s\n- %s", pathToLoad, a.MainFile)
			}

			a.MainFile = pathToLoad
			a.mainFun = mainFun
		}

		schemaFun, _ := globals[schema.SchemaFunctionName].(*starlark.Function)
		if schemaFun != nil {
			if a.schemaFile != "" {
				return fmt.Errorf("multiple files with a %s() function:\n- %s\n- %s", schema.SchemaFunctionName, pathToLoad, a.schemaFile)
			}
			a.schemaFile = pathToLoad

			schemaVal, err := a.Call(context.Background(), schemaFun)
			if err != nil {
				return fmt.Errorf("calling schema function for %s: %w", a.ID, err)
			}

			a.Schema, err = schema.FromStarlark(schemaVal, globals)
			if err != nil {
				return fmt.Errorf("parsing schema for %s: %w", a.ID, err)
			}
		}

	default:
		a.Globals[pathToLoad] = starlark.StringDict{
			"file": &file.File{
				FS:   fsys,
				Path: pathToLoad,
			},
		}
	}

	return nil
}

func (a *Applet) loadCatalog(fsys fs.FS) error {
	dir, err := fs.Sub(fsys, LocaleDir)
	if err != nil {
		return fmt.Errorf("opening locales directory: %w", err)
	}

	d, err := fs.ReadDir(dir, ".")
	if err != nil {
		return fmt.Errorf("listing locales directory: %w", err)
	}

	b := catalog.NewBuilder()

	for _, entry := range d {
		if err := loadLocale(dir, b, entry.Name()); err != nil {
			return err
		}
	}

	a.initializers = append(a.initializers, func(t *starlark.Thread) *starlark.Thread {
		i18n_runtime.AttachCatalogToThread(t, b)
		return t
	})

	return nil
}

func loadLocale(fsys fs.FS, b *catalog.Builder, name string) error {
	if !strings.HasSuffix(name, ".json") {
		return nil
	}

	base := strings.TrimSuffix(name, ".json")
	base = strings.ReplaceAll(base, "_", "-")

	tag, err := language.Parse(base)
	if err != nil {
		return fmt.Errorf("parsing locale %s: %w", name, err)
	}

	f, err := fsys.Open(name)
	if err != nil {
		return fmt.Errorf("opening locale file %s: %w", name, err)
	}
	defer f.Close()

	var msgs map[string]string
	if err := json.NewDecoder(f).Decode(&msgs); err != nil {
		return fmt.Errorf("decoding locale file %s: %w", name, err)
	}

	for key, val := range msgs {
		if err := b.SetString(tag, key, val); err != nil {
			return fmt.Errorf("setting locale string %s: %w", name, err)
		}
	}

	return nil
}

func (a *Applet) newThread(ctx context.Context) *starlark.Thread {
	t := &starlark.Thread{
		Name: a.ID,
		Load: a.loadModule,
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Printf("[%s] %s\n", a.ID, msg)
		},
	}

	starlarkutil.AttachThreadContext(ctx, t)
	random.AttachToThread(t)

	for _, init := range a.initializers {
		t = init(t)
	}

	return t
}

func (a *Applet) loadModule(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	if a.loader != nil {
		mod, err := a.loader(thread, module)
		if err == nil {
			return mod, nil
		}
	}

	switch module {
	case "render.star":
		return render_runtime.LoadRenderModule()

	case i18n_runtime.ModuleName:
		return i18n_runtime.LoadModule()

	case "animation.star":
		return animation_runtime.LoadAnimationModule()

	case "schema.star":
		return schema.LoadModule()

	case "cache.star":
		return LoadCacheModule()

	case "secret.star":
		return LoadSecretModule()

	case "xpath.star":
		return xpath.LoadXPathModule()

	case "bsoup.star":
		return starlibbsoup.LoadModule()

	case "compress/gzip.star":
		return starlark.StringDict{
			starlibgzip.Module.Name: starlibgzip.Module,
		}, nil

	case "compress/zipfile.star":
		// Starlib expects you to load the ZipFile function directly, rather than having it be part of a namespace.
		// Wraps this to be more consistent with other pixlet modules, as follows:
		//   load("compress/zipfile.star", "zipfile")
		//   archive = zipfile.ZipFile("/tmp/foo.zip")
		m, _ := starlibzip.LoadModule()
		return starlark.StringDict{
			"zipfile": &starlarkstruct.Module{
				Name:    "zipfile",
				Members: m,
			},
		}, nil

	case "encoding/base64.star":
		return starlibbase64.LoadModule()

	case "encoding/csv.star":
		return starlibcsv.LoadModule()

	case "encoding/json.star":
		return starlark.StringDict{
			starlibjson.Module.Name: starlibjson.Module,
		}, nil

	case yaml.ModuleName:
		return yaml.LoadModule()

	case "hash.star":
		return starlibhash.LoadModule()

	case "hmac.star":
		return hmac.LoadModule()

	case "http.star":
		return starlarkhttp.LoadModule()

	case "html.star":
		return starlibhtml.LoadModule()

	case "humanize.star":
		return humanize.LoadModule()

	case "math.star":
		return starlark.StringDict{
			starlibmath.Module.Name: starlibmath.Module,
		}, nil

	case "re.star":
		return starlibre.LoadModule()

	case "sunrise.star":
		return sunrise.LoadModule()

	case time_runtime.ModuleName:
		return time_runtime.LoadModule()

	case "random.star":
		return random.LoadModule()

	case "qrcode.star":
		return qrcode.LoadModule()

	case "assert.star":
		return starlarktest.LoadAssertModule()

	default:
		return nil, fmt.Errorf("invalid module: %s", module)
	}
}
