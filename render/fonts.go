package render

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/tronbyt/pixlet/fonts"
	"github.com/zachomedia/go-bdf"
	"go.uber.org/atomic"
	"golang.org/x/image/font"
)

const fontCacheTTLEnv = "PIXLET_FONT_CACHE_TTL"

var (
	FontCacheTTL   = atomic.NewDuration(time.Hour)
	fontCache      = ttlcache.New[string, font.Face]()
	fontCacheMutex = &sync.Mutex{}
)

func init() { //nolint:gochecknoinits
	if val := os.Getenv(fontCacheTTLEnv); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			FontCacheTTL.Store(d)
		} else {
			slog.Warn(fontCacheTTLEnv+" is invalid; using default.", "error", err)
		}
	}

	go fontCache.Start()
}

func GetFontList() ([]string, error) {
	entries, err := fonts.FS.ReadDir(".")
	if err != nil {
		return nil, err
	}

	fontNames := make([]string, 0, len(entries))
	for _, e := range entries {
		fontNames = append(fontNames, strings.TrimSuffix(e.Name(), fonts.Ext))
	}

	return fontNames, nil
}

func GetFont(name string) (font.Face, error) {
	fontCacheMutex.Lock()
	defer fontCacheMutex.Unlock()

	if item := fontCache.Get(name); item != nil {
		return item.Value(), nil
	}

	data, err := fonts.GetBytes(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("unknown font %q", name)
		}
		return nil, fmt.Errorf("reading font %q: %w", name, err)
	}

	fnt, err := bdf.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing font %q: %w", name, err)
	}

	face := fnt.NewFace()
	fontCache.Set(name, face, FontCacheTTL.Load())
	return face, nil
}
