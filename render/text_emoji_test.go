package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/tronbyt/gg"
)

func TestTextWidgetWithEmojis(t *testing.T) {
	tests := []struct {
		name    string
		content string
		font    string
		wantErr bool
	}{
		{
			name:    "single emoji with default font",
			content: "😀",
			font:    "",
			wantErr: false,
		},
		{
			name:    "multiple emojis with 5x8 font",
			content: "😀😂😍",
			font:    "5x8",
			wantErr: false,
		},
		{
			name:    "mixed content with 6x10 font",
			content: "Hello 😎 World",
			font:    "6x10",
			wantErr: false,
		},
		{
			name:    "plain text (no emojis)",
			content: "Plain text",
			font:    "tb-8",
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "",
			font:    "tb-8",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := &Text{
				Content: tt.content,
				Font:    tt.font,
				Color:   color.White,
			}

			err := text.Init()
			if (err != nil) != tt.wantErr {
				t.Errorf("Text.Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify that the image was created
				if text.img == nil {
					t.Error("Text.Init() succeeded but img is nil")
				}

				// Verify image has reasonable dimensions (empty content can have zero width)
				bounds := text.img.Bounds()
				if tt.content == "" {
					// Empty content should have zero or minimal width but positive height
					if bounds.Dy() <= 0 {
						t.Errorf("Text image has invalid height for empty content: %dx%d", bounds.Dx(), bounds.Dy())
					}
				} else {
					// Non-empty content should have positive dimensions
					if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
						t.Errorf("Text image has invalid dimensions: %dx%d", bounds.Dx(), bounds.Dy())
					}
				}
			}
		})
	}
}

func TestTextWidgetSize(t *testing.T) {
	tests := []struct {
		name    string
		content string
		font    string
	}{
		{
			name:    "emoji only",
			content: "😀",
			font:    "5x8",
		},
		{
			name:    "mixed content",
			content: "Hi 😎",
			font:    "6x10",
		},
		{
			name:    "multiple emojis",
			content: "😀😂😍",
			font:    "tb-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := &Text{
				Content: tt.content,
				Font:    tt.font,
				Color:   color.White,
			}

			err := text.Init()
			if err != nil {
				t.Fatalf("Text.Init() failed: %v", err)
			}

			width, height := text.Size()

			// Size should be positive
			if width <= 0 || height <= 0 {
				t.Errorf("Text.Size() = (%d, %d), want positive dimensions", width, height)
			}

			// Size should match image bounds
			bounds := text.img.Bounds()
			if width != bounds.Dx() || height != bounds.Dy() {
				t.Errorf("Text.Size() = (%d, %d), but image bounds = %dx%d",
					width, height, bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestTextWidgetPaint(t *testing.T) {
	text := &Text{
		Content: "Test 😀",
		Font:    "6x10",
		Color:   color.White,
	}

	err := text.Init()
	if err != nil {
		t.Fatalf("Text.Init() failed: %v", err)
	}

	// Create a test canvas
	bounds := image.Rect(0, 0, 100, 20)
	ctx := gg.NewContext(bounds.Dx(), bounds.Dy())

	// Paint should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Text.Paint() panicked: %v", r)
		}
	}()

	text.Paint(ctx, bounds, 0)
}

func TestTextWidgetWithCustomHeight(t *testing.T) {
	text := &Text{
		Content: "😀😂",
		Font:    "5x8",
		Height:  15,
		Color:   color.White,
	}

	err := text.Init()
	if err != nil {
		t.Fatalf("Text.Init() failed: %v", err)
	}

	_, height := text.Size()
	if height != 15 {
		t.Errorf("Expected height 15, got %d", height)
	}
}

func TestTextWidgetWithOffset(t *testing.T) {
	text := &Text{
		Content: "Test 😎",
		Font:    "6x10",
		Offset:  3,
		Color:   color.White,
	}

	err := text.Init()
	if err != nil {
		t.Fatalf("Text.Init() failed: %v", err)
	}

	// Should not error - offset is applied during rendering
	if text.img == nil {
		t.Error("Expected image to be created")
	}
}

func TestTextWidgetMaxWidth(t *testing.T) {
	// Create very long content that should be truncated
	longContent := "😀😂😍😎🌈🎉🎊🎁🎈🎂🍰🎉😀😂😍😎🌈🎉🎊🎁🎈🎂🍰🎉😀😂😍😎🌈🎉🎊🎁🎈🎂🍰🎉"

	text := &Text{
		Content: longContent,
		Font:    "5x8",
		Color:   color.White,
	}

	err := text.Init()
	if err != nil {
		t.Fatalf("Text.Init() failed: %v", err)
	}

	width, _ := text.Size()
	if width > MaxWidth {
		t.Errorf("Text width %d exceeds MaxWidth %d", width, MaxWidth)
	}
}

func BenchmarkTextWithEmojis(b *testing.B) {
	tests := []struct {
		name    string
		content string
	}{
		{"plain_text", "Hello World This Is Plain Text"},
		{"single_emoji", "😀"},
		{"mixed_content", "Hello 😀 World 😎 Test 🌈"},
		{"multiple_emojis", "😀😂😍😎🌈🎉🎊🎁"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				text := &Text{
					Content: tt.content,
					Font:    "6x10",
					Color:   color.White,
				}

				err := text.Init()
				if err != nil {
					b.Fatalf("Text.Init() failed: %v", err)
				}
			}
		})
	}
}
