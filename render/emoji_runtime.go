package render

//go:generate go run ./gen/emoji_pack.go

import (
    "image"
    "image/draw"
    "sync"
)

// hasAnyEmojiSequence returns true if content contains any sequence that could be an emoji.
// Cheap prefilter to decide whether to invoke the more expensive segmentation.
var (
    buildFirstRuneIndexOnce sync.Once
    emojiFirstRune map[rune]struct{}
)

// hasAnyEmojiSequence quickly determines if a string contains at least one
// candidate emoji start rune (including multi-codepoint sequences) by
// consulting a precomputed first-rune index built from emojiIndex keys.
func hasAnyEmojiSequence(s string) bool {
    if len(s) == 0 || len(emojiIndex) == 0 { return false }
    buildFirstRuneIndexOnce.Do(func() {
        emojiFirstRune = make(map[rune]struct{})
        for k := range emojiIndex {
            rs := []rune(k)
            if len(rs) > 0 {
                emojiFirstRune[rs[0]] = struct{}{}
            }
        }
    })
    for _, r := range s {
        if _, ok := emojiFirstRune[r]; ok { return true }
    }
    return false
}

// segmentEmoji breaks a string into a sequence of tokens, where each token is either
// an emoji sequence key present in emojiIndex, or a plain text segment (no emoji inside).
// Longest-match strategy for sequences.
func segmentEmoji(s string) []segment {
    if len(s) == 0 || len(emojiIndex) == 0 { return []segment{} }
    runes := []rune(s)
    tokens := []segment{}
    for i := 0; i < len(runes); {
        matched := false
        maxL := emojiMaxSequence
        if maxL > len(runes)-i { maxL = len(runes)-i }
        for l := maxL; l >= 1; l-- { // longest match first
            key := string(runes[i:i+l])
            if _, ok := emojiIndex[key]; ok {
                tokens = append(tokens, segment{emoji: true, text: key})
                i += l
                matched = true
                break
            }
        }
        if matched { continue }
        // Accumulate plain text until next possible emoji start.
        start := i
        i++
        for i < len(runes) {
            if _, ok := emojiIndex[string(runes[i])]; ok { break }
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
    g, ok := emojiIndex[seq]
    if !ok { return 0 }
    sheet := getEmojiSheet()
    cellX := g.X * emojiCellW
    cellY := g.Y * emojiCellH
    y := baselineY - emojiCellH
    r := image.Rect(cellX, cellY, cellX+emojiCellW, cellY+emojiCellH)
    draw.Draw(dst, r.Add(image.Pt(x, y)).Sub(r.Min), sheet, r.Min, draw.Over)
    return emojiCellW
}
