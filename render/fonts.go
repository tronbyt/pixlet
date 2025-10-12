package render

import (
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/zachomedia/go-bdf"
	"golang.org/x/image/font"
	"tidbyt.dev/pixlet/fonts"
)

var fontCache = map[string]font.Face{}
var fontMutex = &sync.Mutex{}

func GetFontList() []string {
	entries, err := fonts.Fonts.ReadDir(".")
	if err != nil {
		panic(err)
	}

	fontNames := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		fontNames = append(fontNames, trimExt(e.Name()))
	}

	return fontNames
}

func GetFont(name string) (font.Face, error) {
	fontMutex.Lock()
	defer fontMutex.Unlock()

	if font, ok := fontCache[name]; ok {
		return font, nil
	}

	entries, err := fonts.Fonts.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("listing fonts: %w", err)
	}

	var found string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		if trimExt(e.Name()) == name {
			found = e.Name()
			break
		}
	}
	if found == "" {
		return nil, fmt.Errorf("unknown font %q", name)
	}

	data, err := fonts.Fonts.ReadFile(found)
	if err != nil {
		return nil, fmt.Errorf("reading font %q: %w", found, err)
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
