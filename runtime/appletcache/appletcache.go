package appletcache

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"strconv"

	"go.uber.org/atomic"
)

const (
	enabledEnv = "PIXLET_APPLET_CACHE_ENABLED"
	dirEnv     = "PIXLET_APPLET_CACHE_DIR"
)

var (
	Enabled = atomic.NewBool(true)
	Dir     = atomic.String{}
)

func init() { //nolint:gochecknoinits
	if env := os.Getenv(enabledEnv); env != "" {
		if val, err := strconv.ParseBool(env); err == nil {
			Enabled.Store(val)
		} else {
			slog.Warn(enabledEnv+" is invalid; using default.", "error", err)
		}
	}

	if env := os.Getenv(dirEnv); env != "" {
		Dir.Store(env)
	}
}

const starCacheDir = ".starcache"

var ErrDisabled = errors.New("applet cache is disabled")

func New(root *os.Root) (*Cache, error) {
	if !Enabled.Load() {
		return nil, ErrDisabled
	}

	cache := &Cache{}
	if dir := Dir.Load(); dir != "" {
		cache.nest = true

		if err := os.Mkdir(dir, 0755); err != nil && !errors.Is(err, fs.ErrExist) {
			return nil, fmt.Errorf("creating starcache dir: %w", err)
		}

		var err error
		if cache.root, err = os.OpenRoot(dir); err != nil {
			return nil, fmt.Errorf("opening starcache root: %w", err)
		}
	} else {
		if err := root.Mkdir(starCacheDir, 0755); err != nil && !errors.Is(err, fs.ErrExist) {
			return nil, fmt.Errorf("creating starcache dir: %w", err)
		}

		var err error
		if cache.root, err = root.OpenRoot(starCacheDir); err != nil {
			return nil, fmt.Errorf("opening starcache root: %w", err)
		}

		if err := cache.root.WriteFile(".gitignore", []byte("*\n"), 0644); err != nil && !errors.Is(err, fs.ErrExist) {
			_ = cache.root.Close()
			return nil, fmt.Errorf("writing starcache .gitignore: %w", err)
		}
	}

	return cache, nil
}

type Cache struct {
	root *os.Root
	nest bool
}

type Header struct {
	Size    int64
	ModTime int64
}

func (c Cache) NewWriter(_ context.Context, id, key string, sourceFile fs.File) (io.WriteCloser, error) {
	stat, err := sourceFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("stating source file: %w", err)
	}

	cachePath := c.cachePath(id, key)

	if dir := path.Dir(cachePath); dir != "." {
		if err := c.root.Mkdir(id, 0755); err != nil && !errors.Is(err, fs.ErrExist) {
			return nil, fmt.Errorf("creating cache dir: %w", err)
		}
	}

	f, err := c.root.Create(cachePath)
	if err != nil {
		return nil, fmt.Errorf("creating cache file: %w", err)
	}
	var success bool
	defer func() {
		if !success {
			_ = f.Close()
			_ = c.root.Remove(cachePath)
		}
	}()

	data := Header{
		Size:    stat.Size(),
		ModTime: stat.ModTime().UnixNano(),
	}

	if err := binary.Write(f, binary.LittleEndian, &data); err != nil {
		return nil, fmt.Errorf("writing cache file: %w", err)
	}

	success = true
	return f, nil
}

func (c Cache) NewReader(_ context.Context, id, key string, sourceFile fs.File) (io.ReadCloser, bool, error) {
	sourceStat, err := sourceFile.Stat()
	if err != nil {
		return nil, false, fmt.Errorf("stating source file: %w", err)
	}

	cacheFile, err := c.root.Open(c.cachePath(id, key))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("opening cache file: %w", err)
	}
	var success bool
	defer func() {
		if !success {
			_ = cacheFile.Close()
		}
	}()

	var data Header

	if err := binary.Read(cacheFile, binary.LittleEndian, &data); err != nil {
		return nil, false, fmt.Errorf("reading cache file: %w", err)
	}

	if data.Size != sourceStat.Size() || data.ModTime != sourceStat.ModTime().UnixNano() {
		return nil, false, nil
	}

	success = true
	return cacheFile, true, nil
}

func (c Cache) Close() error {
	if c.root != nil {
		return c.root.Close()
	}
	return nil
}

func (c Cache) cachePath(id, key string) string {
	key += ".bin"
	if c.nest {
		return path.Join(id, key)
	}
	return key
}
