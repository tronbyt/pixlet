package main

import (
	"bufio"
	"bytes"
	"cmp"
	_ "embed"
	"fmt"
	"go/format"
	"image"
	"image/draw"
	"image/png"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/tronbyt/pixlet/assets/emoji"
)

const (
	sheetWidthLimit = 640 // soft limit for packing rows; final sheet may be narrower
	glyphPadding    = 1   // pixels of transparent padding on each side of a glyph
	variationsFile  = "emoji-variation-sequences.txt"
	rawDir          = "raw"
	fallbackFile    = "fallback.png"
)

var (
	//go:embed emoji.tmpl
	emojiTmpl string
	outFile   string
	pngFile   string
)

type glyph struct {
	Runes      []rune
	Path       string
	IsFallback bool
}

// Emoji sprite sheet generator.
//
// Scans assets/emoji/raw for PNG files named like:
//
//	U+1F600.png (single codepoint) OR
//	U+1F1FA_U+1F1F8.png (multi-codepoint sequence)
//
// Packs all images into a sprite sheet sized by each glyph's intrinsic bounds and emits
// `fonts/emoji/sprites.go` containing:
//   - embedded sprite sheet bytes
//   - a map[string]Glyph (keys are the actual Unicode sequence string)
//   - constants for sheet dimensions, glyph extrema, and max sequence length
//
// Multi-codepoint sequences (e.g. flags, keycap sequences) are supported.
// ZWJ (\u200D) sequences are included only if PNG assets are present; no special
// shaping is done—longest-match logic at runtime handles them.
func genEmoji() {
	fontsDir := "fonts"
	emojiDir := filepath.Join(fontsDir, "emoji")
	must(os.MkdirAll(emojiDir, 0o755))
	outFile = filepath.Join(emojiDir, "sprites.go")
	pngFile = filepath.Join(emojiDir, "sprites.png")

	emojiVariations := must2(collectEmojiVariations())

	glyphs, maxSeq := collect(emojiVariations)
	if len(glyphs) == 0 {
		panic("no emoji assets found")
	}
	glyphs = append(glyphs, glyph{Path: fallbackFile, IsFallback: true})

	sheet, index, fallbackRect, sheetW, sheetH, maxHeight, maxWidth := buildSheet(glyphs)

	var pngBuf bytes.Buffer
	must(png.Encode(&pngBuf, sheet))
	must(os.WriteFile(pngFile, pngBuf.Bytes(), 0o644))

	writeOutput(filepath.Base(pngFile), index, fallbackRect, maxSeq, sheetW, sheetH, maxHeight, maxWidth)
}

// collect scans rawDir for U+....png names.
func collect(emojiVariations map[string]struct{}) ([]glyph, int) {
	out := []glyph{}
	maxSeq := 1
	must(fs.WalkDir(emoji.FS, rawDir, func(filepath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		base := path.Base(filepath)
		if !strings.HasPrefix(base, "U+") || !strings.HasSuffix(strings.ToLower(base), ".png") {
			return nil
		}
		name := strings.TrimSuffix(base, ".png")
		parts := strings.Split(name, "_")
		runes := []rune{}
		for _, p := range parts {
			if !strings.HasPrefix(p, "U+") { // skip malformed
				return nil
			}
			hexPart := strings.TrimPrefix(p, "U+")
			v, err := strconv.ParseInt(hexPart, 16, 32)
			if err != nil {
				return err
			}
			runes = append(runes, rune(v))
		}
		if len(runes) > 0 {
			originalSeq := string(runes)
			if _, ok := emojiVariations[originalSeq]; ok && runes[len(runes)-1] != '\uFE0F' {
				runes = append(runes, '\uFE0F')
			}
		}
		if l := len(runes); l > maxSeq {
			maxSeq = l
		}
		out = append(out, glyph{Runes: runes, Path: filepath})
		return nil
	}))
	// Stable ordering: by sequence length (desc) then lexicographically by runes.
	slices.SortFunc(out, func(a, b glyph) int {
		if c := cmp.Compare(len(b.Runes), len(a.Runes)); c != 0 {
			return c // longer first
		}
		return slices.Compare(a.Runes, b.Runes) // lexicographic on rune slices
	})
	return out, maxSeq
}

// buildSheet packs glyph images into a sprite sheet using each glyph's intrinsic width, returning
// the sheet along with glyph metadata (pixel coordinates and dimensions) plus aggregate metrics.
func buildSheet(glyphs []glyph) (*image.NRGBA, map[string]image.Rectangle, image.Rectangle, int, int, int, int) {
	type placement struct {
		glyph      glyph
		img        image.Image
		x, y       int
		w, h       int
		boxW, boxH int
	}

	placements := make([]placement, 0, len(glyphs))
	cursorX, cursorY := 0, 0
	rowHeight, rowWidth := 0, 0
	maxRowWidth, maxHeight, maxGlyphWidth := 0, 0, 0

	for _, g := range glyphs {
		f := must2(emoji.FS.Open(g.Path))
		img := must2(png.Decode(f))
		_ = f.Close()

		b := img.Bounds()
		w, h := b.Dx(), b.Dy()
		if w == 0 || h == 0 {
			continue
		}

		totalW := w + glyphPadding*2
		if totalW > maxGlyphWidth {
			maxGlyphWidth = totalW
		}

		boxW := totalW
		boxH := h

		if cursorX != 0 && cursorX+boxW > sheetWidthLimit {
			if rowWidth > maxRowWidth {
				maxRowWidth = rowWidth
			}
			cursorX = 0
			cursorY += rowHeight
			rowHeight = 0
		}

		placements = append(placements, placement{
			glyph: g,
			img:   img,
			x:     cursorX + glyphPadding,
			y:     cursorY,
			w:     w,
			h:     h,
			boxW:  boxW,
			boxH:  boxH,
		})
		cursorX += boxW
		rowWidth = cursorX
		if rowHeight < h {
			rowHeight = h
		}
		if maxHeight < h {
			maxHeight = h
		}
		if rowWidth > maxRowWidth {
			maxRowWidth = rowWidth
		}
	}

	if len(placements) == 0 {
		panic("no glyphs found")
	}

	if rowWidth > maxRowWidth {
		maxRowWidth = rowWidth
	}

	sheetWidth := maxRowWidth
	if sheetWidth == 0 {
		sheetWidth = sheetWidthLimit
	}
	sheetHeight := cursorY + rowHeight

	sheet := image.NewNRGBA(image.Rect(0, 0, sheetWidth, sheetHeight))
	index := make(map[string]image.Rectangle, len(placements))
	var fallbackRect image.Rectangle

	for _, p := range placements {
		boxRect := image.Rect(p.x-glyphPadding, p.y, p.x-glyphPadding+p.boxW, p.y+p.boxH)
		draw.Draw(sheet, boxRect, image.Transparent, image.Point{}, draw.Src)
		glyphRect := image.Rect(p.x, p.y, p.x+p.w, p.y+p.h)
		draw.Draw(sheet, glyphRect, p.img, p.img.Bounds().Min, draw.Over)
		if p.glyph.IsFallback {
			fallbackRect = boxRect
			continue
		}
		seqKey := string(p.glyph.Runes)
		index[seqKey] = boxRect
	}

	if fallbackRect.Empty() {
		panic("failed to generate fallback glyph, is " + fallbackFile + " empty?")
	}

	return sheet, index, fallbackRect, sheetWidth, sheetHeight, maxHeight, maxGlyphWidth
}

func collectEmojiVariations() (map[string]struct{}, error) {
	f, err := emoji.FS.Open(variationsFile)
	if err != nil {
		return nil, fmt.Errorf("open emoji variation sequences: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	sequences := make(map[string]struct{})
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ";")
		if len(parts) < 2 {
			continue
		}
		if strings.TrimSpace(parts[1]) != "emoji style" {
			continue
		}
		codepoints := strings.Fields(parts[0])
		if len(codepoints) == 0 {
			continue
		}
		if !strings.EqualFold(codepoints[len(codepoints)-1], "FE0F") {
			continue
		}
		runes := make([]rune, 0, len(codepoints)-1)
		valid := true
		for _, cp := range codepoints[:len(codepoints)-1] {
			v, err := strconv.ParseInt(cp, 16, 32)
			if err != nil {
				valid = false
				break
			}
			runes = append(runes, rune(v))
		}
		if valid && len(runes) != 0 {
			sequences[string(runes)] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse emoji variation sequences: %w", err)
	}
	return sequences, nil
}

type emojiTemplateData struct {
	PNGFilename string
	SheetWidth  int
	SheetHeight int
	MaxWidth    int
	MaxHeight   int
	MaxSequence int
	Fallback    image.Rectangle
	Index       map[string]image.Rectangle
	Keys        []string
}

func writeOutput(pngFileName string, index map[string]image.Rectangle, fallback image.Rectangle, maxSeq, sheetW, sheetH, maxHeight, maxWidth int) {
	keys := slices.Collect(maps.Keys(index))
	slices.SortFunc(keys, func(a, b string) int {
		return cmp.Or(
			cmp.Compare(utf8.RuneCountInString(b), utf8.RuneCountInString(a)),
			cmp.Compare(a, b),
		)
	})

	data := emojiTemplateData{
		PNGFilename: pngFileName,
		SheetWidth:  sheetW,
		SheetHeight: sheetH,
		MaxWidth:    maxWidth,
		MaxHeight:   maxHeight,
		MaxSequence: maxSeq,
		Fallback:    fallback,
		Index:       index,
		Keys:        keys,
	}

	t := must2(template.New("emoji").Parse(emojiTmpl))

	var b bytes.Buffer
	must(t.Execute(&b, data))

	content := must2(format.Source(b.Bytes()))
	must(os.WriteFile(outFile, content, 0o644))
}
