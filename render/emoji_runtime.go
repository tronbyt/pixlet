package render

//go:generate go run ./gen/emoji_pack.go

import (
	"image"
	"image/draw"

	"tidbyt.dev/pixlet/fonts/emoji"
)

// containsEmoji determines if a string contains at least one supported emoji
func containsEmoji(s string) bool {
	idx := emoji.Index
	if len(s) == 0 || len(idx) == 0 {
		return false
	}
	runes := []rune(s)
	for i := range len(runes) {
		maxL := min(len(runes)-i, emoji.MaxSequence)
		for l := maxL; l >= 1; l-- {
			if _, ok := idx[string(runes[i:i+l])]; ok {
				return true
			}
		}
	}
	return false
}

// segmentEmoji breaks a string into a sequence of tokens, where each token is either
// an emoji sequence key present in Index, or a plain text segment (no emoji inside).
// Longest-match strategy for sequences.
func segmentEmoji(s string) []segment {
	idx := emoji.Index
	if len(s) == 0 || len(idx) == 0 {
		return []segment{}
	}
	runes := []rune(s)
	tokens := []segment{}
	for i := 0; i < len(runes); {
		matched := false
		maxL := emoji.MaxSequence
		if maxL > len(runes)-i {
			maxL = len(runes) - i
		}
		for l := maxL; l >= 1; l-- { // longest match first
			key := string(runes[i : i+l])
			if _, ok := idx[key]; ok {
				tokens = append(tokens, segment{emoji: true, text: key})
				i += l
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		// Accumulate plain text until next possible emoji start.
		start := i
		i++
		for i < len(runes) {
			if _, ok := idx[string(runes[i])]; ok {
				break
			}
			i++
		}
		tokens = append(tokens, segment{emoji: false, text: string(runes[start:i])})
	}
	return tokens
}

type segment struct {
	emoji bool
	text  string // either a plain text run or the exact emoji sequence string
}

// drawEmojiSequence renders an emoji (sequence) at x, baseline-aligned.
// Returns advance width in pixels.
func drawEmojiSequence(dst draw.Image, seq string, x, baselineY int) int {
	g, ok := emoji.Index[seq]
	if !ok {
		return 0
	}
	sheet, err := emoji.Sheet()
	if err != nil {
		return 0
	}
	cellX := g.X * emoji.CellW
	cellY := g.Y * emoji.CellH
	y := baselineY - emoji.CellH
	r := image.Rect(cellX, cellY, cellX+emoji.CellW, cellY+emoji.CellH)
	draw.Draw(dst, r.Add(image.Pt(x, y)).Sub(r.Min), sheet, r.Min, draw.Over)
	return emoji.CellW
}
