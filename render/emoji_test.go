package render

import (
	"math"
	"testing"

	font "github.com/tronbyt/pixlet/fonts/emoji"
)

func scaledWidth(seq string, height int) int {
	glyph, ok := font.Index[seq]
	if !ok || glyph.Empty() {
		if fb := font.Fallback; !fb.Empty() {
			glyph = fb
		} else {
			return height
		}
	}
	if height == 0 {
		return glyph.Dx()
	}
	innerH := glyph.Dy()
	if innerH == 0 {
		return height
	}
	totalW := glyph.Dx()
	ratio := float64(totalW) / float64(innerH)
	width := int(math.Round(float64(height) * ratio))
	if width <= 0 {
		width = 1
	}
	return width
}

func TestEmojiWidget(t *testing.T) {
	tests := []struct {
		name    string
		emoji   string
		height  int
		wantErr bool
	}{
		{
			name:    "valid emoji with small height",
			emoji:   "😀",
			height:  8,
			wantErr: false,
		},
		{
			name:    "valid emoji with medium height",
			emoji:   "🚀",
			height:  16,
			wantErr: false,
		},
		{
			name:    "valid emoji with large height",
			emoji:   "🎉",
			height:  32,
			wantErr: false,
		},
		{
			name:    "complex emoji sequence",
			emoji:   "👨‍👩‍👧‍👦",
			height:  20,
			wantErr: false,
		},
		{
			name:    "flag emoji",
			emoji:   "🇺🇸",
			height:  12,
			wantErr: false,
		},
		{
			name:    "empty emoji string",
			emoji:   "",
			height:  16,
			wantErr: true,
		},
		{
			name:    "zero height",
			emoji:   "😀",
			height:  0,
			wantErr: false,
		},
		{
			name:    "negative height",
			emoji:   "😀",
			height:  -5,
			wantErr: true,
		},
		{
			name:    "unknown emoji",
			emoji:   "🤷‍♀️",
			height:  16,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emoji := &Emoji{
				EmojiStr: tt.emoji,
				Height:   tt.height,
			}

			err := emoji.Init()
			if (err != nil) != tt.wantErr {
				t.Errorf("Emoji.Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify that the image was created
				if emoji.img == nil {
					t.Error("Emoji.Init() succeeded but img is nil")
				}

				// Verify image has correct dimensions
				width, height := emoji.Size()
				if tt.height != 0 && height != tt.height {
					t.Errorf("Expected height %d, got %d", tt.height, height)
				}

				expectedWidth := scaledWidth(tt.emoji, tt.height)
				if width != expectedWidth {
					t.Errorf("Expected width %d based on glyph aspect ratio, got %d", expectedWidth, width)
				}

				// Verify dimensions are positive
				if width <= 0 || height <= 0 {
					t.Errorf("Emoji has invalid dimensions: %dx%d", width, height)
				}
			}
		})
	}
}

func TestEmojiWidgetSize(t *testing.T) {
	tests := []struct {
		name   string
		emoji  string
		height int
	}{
		{"small emoji", "😀", 8},
		{"medium emoji", "🚀", 16},
		{"large emoji", "🎉", 32},
		{"very large emoji", "😁", 48},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emoji := &Emoji{
				EmojiStr: tt.emoji,
				Height:   tt.height,
			}

			err := emoji.Init()
			if err != nil {
				t.Fatalf("Emoji.Init() failed: %v", err)
			}

			width, height := emoji.Size()

			if height != tt.height {
				t.Errorf("Emoji.Size() height = %d, want %d", height, tt.height)
			}

			expectedWidth := scaledWidth(tt.emoji, tt.height)
			if width != expectedWidth {
				t.Errorf("Emoji.Size() width = %d, want %d", width, expectedWidth)
			}

			// Size should match image bounds
			bounds := emoji.img.Bounds()
			if width != bounds.Dx() || height != bounds.Dy() {
				t.Errorf("Emoji.Size() = (%d, %d), but image bounds = %dx%d",
					width, height, bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestEmojiWidgetFrameCount(t *testing.T) {
	emoji := &Emoji{
		EmojiStr: "😀",
		Height:   16,
	}

	err := emoji.Init()
	if err != nil {
		t.Fatalf("Emoji.Init() failed: %v", err)
	}

	// Emoji widgets should always have exactly 1 frame
	frames := emoji.FrameCount(emoji.img.Bounds())
	if frames != 1 {
		t.Errorf("Expected 1 frame, got %d", frames)
	}
}

func BenchmarkEmojiWidget(b *testing.B) {
	tests := []struct {
		name   string
		emoji  string
		height int
	}{
		{"small_emoji", "😀", 8},
		{"medium_emoji", "🚀", 16},
		{"large_emoji", "🎉", 32},
		{"xlarge_emoji", "⚡", 64},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				emoji := &Emoji{
					EmojiStr: tt.emoji,
					Height:   tt.height,
				}

				err := emoji.Init()
				if err != nil {
					b.Fatalf("Emoji.Init() failed: %v", err)
				}
			}
		})
	}
}
