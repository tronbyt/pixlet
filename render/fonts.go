package render

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/zachomedia/go-bdf"
	"golang.org/x/image/font"
	"tidbyt.dev/pixlet/fonts"
)

var fontCache = map[string]font.Face{}
var fontMutex = &sync.Mutex{}

func GetFontList() ([]string, error) {
	entries, err := fonts.Fonts.ReadDir(".")
	if err != nil {
		return nil, err
	}

	fontNames := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		fontNames = append(fontNames, trimExt(e.Name()))
	}

	return fontNames, nil
}

func GetFont(name string) (font.Face, error) {
	fontMutex.Lock()
	defer fontMutex.Unlock()

	if font, ok := fontCache[name]; ok {
		return font, nil
	}

	data, err := fonts.Fonts.ReadFile(name + ".bdf")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("unknown font %q", name)
		}
		return nil, fmt.Errorf("reading font %q: %w", name, err)
	}

	f, err := bdf.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing font %q: %w", name, err)
	}

	fontCache[name] = f.NewFace()
	return fontCache[name], nil
}

func trimExt(filename string) string {
	return strings.TrimSuffix(filename, path.Ext(filename))
}
