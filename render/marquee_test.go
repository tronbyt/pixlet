package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarqueeNoScrollHorizontal(t *testing.T) {
	m := Marquee{
		Width: 6,
		Child: Row{
			Children: []Widget{
				Box{Width: 3, Height: 3, Color: color.RGBA{0xff, 0, 0, 0xff}},
				Box{Width: 2, Height: 2, Color: color.RGBA{0, 0xff, 0, 0xff}},
				Box{Width: 1, Height: 1, Color: color.RGBA{0, 0, 0xff, 0xff}},
			},
		},
	}

	mv := Marquee{
		Height: 3,
		Child: Row{
			Children: []Widget{
				Box{Width: 3, Height: 3, Color: color.RGBA{0xff, 0, 0, 0xff}},
				Box{Width: 2, Height: 2, Color: color.RGBA{0, 0xff, 0, 0xff}},
				Box{Width: 1, Height: 1, Color: color.RGBA{0, 0, 0xff, 0xff}},
			},
		},
		ScrollDirection: "vertical",
	}

	// Child fits so there's just 1 single frame
	assert.Equal(t, 1, m.FrameCount(image.Rect(0, 0, 100, 100)))
	assert.Equal(t, 1, mv.FrameCount(image.Rect(0, 0, 100, 100)))
	im := PaintWidget(m, image.Rect(0, 0, 100, 100), 0)
	imv := PaintWidget(mv, image.Rect(0, 0, 100, 100), 0)
	require.NoError(t, checkImage([]string{
		"rrrggb",
		"rrrgg.",
		"rrr...",
	}, im))
	require.NoError(t, checkImage([]string{
		"rrrggb",
		"rrrgg.",
		"rrr...",
	}, imv))
}

func TestMarqueeNoScrollAlignCenter(t *testing.T) {
	m := Marquee{
		Width: 8,
		Child: Row{
			Children: []Widget{
				Box{Width: 3, Height: 3, Color: color.RGBA{0xff, 0, 0, 0xff}},
				Box{Width: 2, Height: 2, Color: color.RGBA{0, 0xff, 0, 0xff}},
				Box{Width: 1, Height: 1, Color: color.RGBA{0, 0, 0xff, 0xff}},
			},
		},
		Align: "center",
	}

	mv := Marquee{
		Height: 5,
		Child: Row{
			Children: []Widget{
				Box{Width: 3, Height: 3, Color: color.RGBA{0xff, 0, 0, 0xff}},
				Box{Width: 2, Height: 2, Color: color.RGBA{0, 0xff, 0, 0xff}},
				Box{Width: 1, Height: 1, Color: color.RGBA{0, 0, 0xff, 0xff}},
			},
		},
		ScrollDirection: "vertical",
		Align:           "center",
	}

	// Child fits so there's just 1 single frame
	assert.Equal(t, 1, m.FrameCount(image.Rect(0, 0, 100, 100)))
	assert.Equal(t, 1, mv.FrameCount(image.Rect(0, 0, 100, 100)))
	im := PaintWidget(m, image.Rect(0, 0, 100, 100), 0)
	imv := PaintWidget(mv, image.Rect(0, 0, 100, 100), 0)
	require.NoError(t, checkImage([]string{
		".rrrggb.",
		".rrrgg..",
		".rrr....",
	}, im))
	require.NoError(t, checkImage([]string{
		"......",
		"rrrggb",
		"rrrgg.",
		"rrr...",
		"......",
	}, imv))
}

func TestMarqueeNoScrollAlignEnd(t *testing.T) {
	m := Marquee{
		Width: 8,
		Child: Row{
			Children: []Widget{
				Box{Width: 3, Height: 3, Color: color.RGBA{0xff, 0, 0, 0xff}},
				Box{Width: 2, Height: 2, Color: color.RGBA{0, 0xff, 0, 0xff}},
				Box{Width: 1, Height: 1, Color: color.RGBA{0, 0, 0xff, 0xff}},
			},
		},
		Align: "end",
	}

	mv := Marquee{
		Height: 5,
		Child: Row{
			Children: []Widget{
				Box{Width: 3, Height: 3, Color: color.RGBA{0xff, 0, 0, 0xff}},
				Box{Width: 2, Height: 2, Color: color.RGBA{0, 0xff, 0, 0xff}},
				Box{Width: 1, Height: 1, Color: color.RGBA{0, 0, 0xff, 0xff}},
			},
		},
		ScrollDirection: "vertical",
		Align:           "end",
	}

	// Child fits so there's just 1 single frame
	assert.Equal(t, 1, m.FrameCount(image.Rect(0, 0, 100, 100)))
	assert.Equal(t, 1, mv.FrameCount(image.Rect(0, 0, 100, 100)))
	im := PaintWidget(m, image.Rect(0, 0, 100, 100), 0)
	imv := PaintWidget(mv, image.Rect(0, 0, 100, 100), 0)
	require.NoError(t, checkImage([]string{
		"..rrrggb",
		"..rrrgg.",
		"..rrr...",
	}, im))
	require.NoError(t, checkImage([]string{
		"......",
		"......",
		"rrrggb",
		"rrrgg.",
		"rrr...",
	}, imv))
}

// The addition of OffsetStart and OffsetEnd changes the default
// behaviour of Marquee. Passing start==width and end==0 mimics the
// old default.
func TestMarqueeOldBehavior(t *testing.T) {
	m := Marquee{
		Width:       6,
		OffsetStart: 6,
		OffsetEnd:   0,
		Child: Row{
			Children: []Widget{
				Box{Width: 3, Height: 3, Color: color.RGBA{0xff, 0, 0, 0xff}},
				Box{Width: 3, Height: 2, Color: color.RGBA{0, 0xff, 0, 0xff}},
				Box{Width: 3, Height: 1, Color: color.RGBA{0, 0, 0xff, 0xff}},
			},
		},
	}

	// The child's 9 pixels will be scrolled into view (7 frames),
	// scrolled out of view (9 frames) and then finally scrolled
	// back into view again (6 frames). 22 frames in total.
	assert.Equal(t, 22, m.FrameCount(image.Rect(0, 0, 100, 100)))

	// Scrolling into view
	require.NoError(t, checkImage([]string{
		"......",
		"......",
		"......",
	}, PaintWidget(m, image.Rect(0, 0, 100, 100), 0)))

	require.NoError(t, checkImage([]string{
		"....rr",
		"....rr",
		"....rr",
	}, PaintWidget(m, image.Rect(0, 0, 100, 100), 2)))

	require.NoError(t, checkImage([]string{
		"rrrggg",
		"rrrggg",
		"rrr...",
	}, PaintWidget(m, image.Rect(0, 0, 100, 100), 6)))

	// Scrolling out of view
	require.NoError(t, checkImage([]string{
		"rgggbb",
		"rggg..",
		"r.....",
	}, PaintWidget(m, image.Rect(0, 0, 100, 100), 8)))

	require.NoError(t, checkImage([]string{
		"b.....",
		"......",
		"......",
	}, PaintWidget(m, image.Rect(0, 0, 100, 100), 14)))

	require.NoError(t, checkImage([]string{
		"......",
		"......",
		"......",
	}, PaintWidget(m, image.Rect(0, 0, 100, 100), 15)))

	// Scrolling back into view
	require.NoError(t, checkImage([]string{
		"...rrr",
		"...rrr",
		"...rrr",
	}, PaintWidget(m, image.Rect(0, 0, 100, 100), 18)))

	require.NoError(t, checkImage([]string{
		"rrrggg",
		"rrrggg",
		"rrr...",
	}, PaintWidget(m, image.Rect(0, 0, 100, 100), 21)))

	// Later frames keep it fixed in the last frame. This makes
	// multiple simultaneous marquees look nice when they've
	// different length.

	require.NoError(t, checkImage([]string{
		"rrrggg",
		"rrrggg",
		"rrr...",
	}, PaintWidget(m, image.Rect(0, 0, 100, 100), 22)))

	require.NoError(t, checkImage([]string{
		"rrrggg",
		"rrrggg",
		"rrr...",
	}, PaintWidget(m, image.Rect(0, 0, 100, 100), 26)))

	require.NoError(t, checkImage([]string{
		"rrrggg",
		"rrrggg",
		"rrr...",
	}, PaintWidget(m, image.Rect(0, 0, 100, 100), 100000)))
}

func TestMarqueeOffsetIdentical(t *testing.T) {
	child := Row{
		Children: []Widget{
			Box{Width: 1, Height: 1, Color: color.RGBA{0xff, 0, 0, 0xff}},
			Box{Width: 2, Height: 1, Color: color.RGBA{0, 0xff, 0, 0xff}},
			Box{Width: 4, Height: 1, Color: color.RGBA{0, 0, 0xff, 0xff}},
		},
	}
	m := Marquee{
		Width: 6,
		Child: child,
	}
	im := image.Rect(0, 0, 100, 100)

	// Check that identical frames are not repeated after
	// another, if start and end offset are identical.
	require.NoError(t, checkImage([]string{"rggbbb"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"ggbbbb"}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"gbbbb."}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{"bbbb.."}, PaintWidget(m, im, 3)))
	require.NoError(t, checkImage([]string{"bbb..."}, PaintWidget(m, im, 4)))
	require.NoError(t, checkImage([]string{"bb...."}, PaintWidget(m, im, 5)))
	require.NoError(t, checkImage([]string{"b....."}, PaintWidget(m, im, 6)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 7)))
	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 8)))
	require.NoError(t, checkImage([]string{"....rg"}, PaintWidget(m, im, 9)))
	require.NoError(t, checkImage([]string{"...rgg"}, PaintWidget(m, im, 10)))
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 11)))
	require.NoError(t, checkImage([]string{".rggbb"}, PaintWidget(m, im, 12)))
	assert.Equal(t, 13, m.FrameCount(im))

	m.OffsetStart = 3
	m.OffsetEnd = 3
	require.NoError(t, checkImage([]string{"...rgg"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{".rggbb"}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{"rggbbb"}, PaintWidget(m, im, 3)))
	require.NoError(t, checkImage([]string{"ggbbbb"}, PaintWidget(m, im, 4)))
	require.NoError(t, checkImage([]string{"gbbbb."}, PaintWidget(m, im, 5)))
	require.NoError(t, checkImage([]string{"bbbb.."}, PaintWidget(m, im, 6)))
	require.NoError(t, checkImage([]string{"bbb..."}, PaintWidget(m, im, 7)))
	require.NoError(t, checkImage([]string{"bb...."}, PaintWidget(m, im, 8)))
	require.NoError(t, checkImage([]string{"b....."}, PaintWidget(m, im, 9)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 10)))
	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 11)))
	require.NoError(t, checkImage([]string{"....rg"}, PaintWidget(m, im, 12)))
	assert.Equal(t, 13, m.FrameCount(im))
}

func TestMarqueeOffsetStart(t *testing.T) {
	child := Row{
		Children: []Widget{
			Box{Width: 1, Height: 1, Color: color.RGBA{0xff, 0, 0, 0xff}},
			Box{Width: 2, Height: 1, Color: color.RGBA{0, 0xff, 0, 0xff}},
			Box{Width: 4, Height: 1, Color: color.RGBA{0, 0, 0xff, 0xff}},
		},
	}
	m := Marquee{
		Width: 6,
		Child: child,
	}
	im := image.Rect(0, 0, 100, 100)

	// OffsetStart affects the initial position of the child
	m.OffsetStart = 2
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{".rggbb"}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"rggbbb"}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{"ggbbbb"}, PaintWidget(m, im, 3)))
	require.NoError(t, checkImage([]string{"gbbbb."}, PaintWidget(m, im, 4)))
	require.NoError(t, checkImage([]string{"bbbb.."}, PaintWidget(m, im, 5)))
	require.NoError(t, checkImage([]string{"bbb..."}, PaintWidget(m, im, 6)))
	require.NoError(t, checkImage([]string{"bb...."}, PaintWidget(m, im, 7)))
	require.NoError(t, checkImage([]string{"b....."}, PaintWidget(m, im, 8)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 9)))

	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 10)))
	require.NoError(t, checkImage([]string{"....rg"}, PaintWidget(m, im, 11)))
	require.NoError(t, checkImage([]string{"...rgg"}, PaintWidget(m, im, 12)))
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 13)))
	require.NoError(t, checkImage([]string{".rggbb"}, PaintWidget(m, im, 14)))
	require.NoError(t, checkImage([]string{"rggbbb"}, PaintWidget(m, im, 15)))
	assert.Equal(t, 16, m.FrameCount(im))

	// Negative OffsetStart
	m.OffsetStart = -2
	require.NoError(t, checkImage([]string{"gbbbb."}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"bbbb.."}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"bbb..."}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{"bb...."}, PaintWidget(m, im, 3)))
	require.NoError(t, checkImage([]string{"b....."}, PaintWidget(m, im, 4)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 5)))
	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 6)))
	require.NoError(t, checkImage([]string{"....rg"}, PaintWidget(m, im, 7)))
	require.NoError(t, checkImage([]string{"...rgg"}, PaintWidget(m, im, 8)))
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 9)))
	require.NoError(t, checkImage([]string{".rggbb"}, PaintWidget(m, im, 10)))
	require.NoError(t, checkImage([]string{"rggbbb"}, PaintWidget(m, im, 11)))
	assert.Equal(t, 12, m.FrameCount(im))

	// Overly negative OffsetStart is truncated to child width
	m.OffsetStart = -1000
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"....rg"}, PaintWidget(m, im, 2)))
	assert.Equal(t, 7, m.FrameCount(im))
	m.OffsetStart = -7
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 1)))
	assert.Equal(t, 7, m.FrameCount(im))
	m.OffsetStart = -8
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 1)))
	assert.Equal(t, 7, m.FrameCount(im))
	m.OffsetStart = -6
	require.NoError(t, checkImage([]string{"b....."}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 2)))
	assert.Equal(t, 8, m.FrameCount(im))
}

func TestMarqueeOffsetEnd(t *testing.T) {
	child := Row{
		Children: []Widget{
			Box{Width: 1, Height: 1, Color: color.RGBA{0xff, 0, 0, 0xff}},
			Box{Width: 2, Height: 1, Color: color.RGBA{0, 0xff, 0, 0xff}},
			Box{Width: 4, Height: 1, Color: color.RGBA{0, 0, 0xff, 0xff}},
		},
	}
	m := Marquee{
		Width: 6,
		Child: child,
	}
	im := image.Rect(0, 0, 100, 100)

	// OffsetEnd affects the final position of the child
	m.OffsetEnd = 2
	require.NoError(t, checkImage([]string{"rggbbb"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"ggbbbb"}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"gbbbb."}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{"bbbb.."}, PaintWidget(m, im, 3)))
	require.NoError(t, checkImage([]string{"bbb..."}, PaintWidget(m, im, 4)))
	require.NoError(t, checkImage([]string{"bb...."}, PaintWidget(m, im, 5)))
	require.NoError(t, checkImage([]string{"b....."}, PaintWidget(m, im, 6)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 7)))
	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 8)))
	require.NoError(t, checkImage([]string{"....rg"}, PaintWidget(m, im, 9)))
	require.NoError(t, checkImage([]string{"...rgg"}, PaintWidget(m, im, 10)))
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 11)))
	assert.Equal(t, 12, m.FrameCount(im))
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 12)))
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 13)))
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 1024)))

	// Negative offset places child outside of marquee
	m.OffsetEnd = -4
	require.NoError(t, checkImage([]string{"rggbbb"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"ggbbbb"}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"gbbbb."}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{"...rgg"}, PaintWidget(m, im, 10)))
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 11)))
	require.NoError(t, checkImage([]string{".rggbb"}, PaintWidget(m, im, 12)))
	require.NoError(t, checkImage([]string{"rggbbb"}, PaintWidget(m, im, 13)))
	require.NoError(t, checkImage([]string{"ggbbbb"}, PaintWidget(m, im, 14)))
	require.NoError(t, checkImage([]string{"gbbbb."}, PaintWidget(m, im, 15)))
	require.NoError(t, checkImage([]string{"bbbb.."}, PaintWidget(m, im, 16)))
	require.NoError(t, checkImage([]string{"bbb..."}, PaintWidget(m, im, 17)))
	assert.Equal(t, 18, m.FrameCount(im))
	require.NoError(t, checkImage([]string{"bbb..."}, PaintWidget(m, im, 18)))
	require.NoError(t, checkImage([]string{"bbb..."}, PaintWidget(m, im, 19)))
	require.NoError(t, checkImage([]string{"bbb..."}, PaintWidget(m, im, 1024)))

	// Very negative offset is truncated to width of child
	m.OffsetEnd = -133
	require.NoError(t, checkImage([]string{"rggbbb"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"bbb..."}, PaintWidget(m, im, 17)))
	require.NoError(t, checkImage([]string{"bb...."}, PaintWidget(m, im, 18)))
	require.NoError(t, checkImage([]string{"b....."}, PaintWidget(m, im, 19)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 20)))
	assert.Equal(t, 21, m.FrameCount(im))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 21)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 22)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 23)))

	// OffsetEnd >= width means it doesn't scroll back
	m.OffsetEnd = 6
	require.NoError(t, checkImage([]string{"rggbbb"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"b....."}, PaintWidget(m, im, 6)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 7)))
	assert.Equal(t, 8, m.FrameCount(im))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 8)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 9)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 1024)))
}

func TestMarqueeDelayScrollOffsetStart(t *testing.T) {
	child := Row{
		Children: []Widget{
			Box{Width: 1, Height: 1, Color: color.RGBA{0xff, 0, 0, 0xff}},
			Box{Width: 2, Height: 1, Color: color.RGBA{0, 0xff, 0, 0xff}},
			Box{Width: 4, Height: 1, Color: color.RGBA{0, 0, 0xff, 0xff}},
		},
	}
	m := Marquee{
		Width: 6,
		Child: child,
		Delay: 2,
	}
	im := image.Rect(0, 0, 100, 100)

	// OffsetStart affects the initial position of the child
	m.OffsetStart = 2
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{".rggbb"}, PaintWidget(m, im, 3)))
	require.NoError(t, checkImage([]string{"rggbbb"}, PaintWidget(m, im, 4)))
	require.NoError(t, checkImage([]string{"ggbbbb"}, PaintWidget(m, im, 5)))
	require.NoError(t, checkImage([]string{"gbbbb."}, PaintWidget(m, im, 6)))
	require.NoError(t, checkImage([]string{"bbbb.."}, PaintWidget(m, im, 7)))
	require.NoError(t, checkImage([]string{"bbb..."}, PaintWidget(m, im, 8)))
	require.NoError(t, checkImage([]string{"bb...."}, PaintWidget(m, im, 9)))
	require.NoError(t, checkImage([]string{"b....."}, PaintWidget(m, im, 10)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 11)))

	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 12)))
	require.NoError(t, checkImage([]string{"....rg"}, PaintWidget(m, im, 13)))
	require.NoError(t, checkImage([]string{"...rgg"}, PaintWidget(m, im, 14)))
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 15)))
	require.NoError(t, checkImage([]string{".rggbb"}, PaintWidget(m, im, 16)))
	require.NoError(t, checkImage([]string{"rggbbb"}, PaintWidget(m, im, 17)))
	assert.Equal(t, 18, m.FrameCount(im))

	// // Negative OffsetStart
	m.OffsetStart = -2
	require.NoError(t, checkImage([]string{"gbbbb."}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"gbbbb."}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"gbbbb."}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{"bbbb.."}, PaintWidget(m, im, 3)))
	require.NoError(t, checkImage([]string{"bbb..."}, PaintWidget(m, im, 4)))
	require.NoError(t, checkImage([]string{"bb...."}, PaintWidget(m, im, 5)))
	require.NoError(t, checkImage([]string{"b....."}, PaintWidget(m, im, 6)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 7)))
	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 8)))
	require.NoError(t, checkImage([]string{"....rg"}, PaintWidget(m, im, 9)))
	require.NoError(t, checkImage([]string{"...rgg"}, PaintWidget(m, im, 10)))
	require.NoError(t, checkImage([]string{"..rggb"}, PaintWidget(m, im, 11)))
	require.NoError(t, checkImage([]string{".rggbb"}, PaintWidget(m, im, 12)))
	require.NoError(t, checkImage([]string{"rggbbb"}, PaintWidget(m, im, 13)))
	assert.Equal(t, 14, m.FrameCount(im))

	// // Overly negative OffsetStart is truncated to child width
	m.OffsetStart = -1000
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 3)))
	require.NoError(t, checkImage([]string{"....rg"}, PaintWidget(m, im, 4)))
	assert.Equal(t, 9, m.FrameCount(im))
	m.OffsetStart = -7
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 3)))
	assert.Equal(t, 9, m.FrameCount(im))
	m.OffsetStart = -8
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 3)))
	assert.Equal(t, 9, m.FrameCount(im))
	m.OffsetStart = -6
	require.NoError(t, checkImage([]string{"b....."}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"b....."}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"b....."}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{"......"}, PaintWidget(m, im, 3)))
	require.NoError(t, checkImage([]string{".....r"}, PaintWidget(m, im, 4)))
	assert.Equal(t, 10, m.FrameCount(im))
}

func TestMarqueeVerticalScroll(t *testing.T) {
	child := Column{
		Children: []Widget{
			Box{Width: 1, Height: 1, Color: color.RGBA{0xff, 0, 0, 0xff}},
			Box{Width: 1, Height: 2, Color: color.RGBA{0, 0xff, 0, 0xff}},
			Box{Width: 1, Height: 4, Color: color.RGBA{0, 0, 0xff, 0xff}},
		},
	}
	m := Marquee{
		Height:          6,
		Child:           child,
		ScrollDirection: "vertical",
	}
	im := image.Rect(0, 0, 100, 100)

	// OffsetEnd affects the final position of the child
	m.OffsetStart = 2
	require.NoError(t, checkImage([]string{
		".",
		".",
		"r",
		"g",
		"g",
		"b",
	}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{
		".",
		"r",
		"g",
		"g",
		"b",
		"b",
	}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{
		"r",
		"g",
		"g",
		"b",
		"b",
		"b",
	}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{
		"g",
		"g",
		"b",
		"b",
		"b",
		"b",
	}, PaintWidget(m, im, 3)))
	require.NoError(t, checkImage([]string{
		"g",
		"b",
		"b",
		"b",
		"b",
		".",
	}, PaintWidget(m, im, 4)))
	require.NoError(t, checkImage([]string{
		"b",
		"b",
		"b",
		"b",
		".",
		".",
	}, PaintWidget(m, im, 5)))
	require.NoError(t, checkImage([]string{
		"b",
		"b",
		"b",
		".",
		".",
		".",
	}, PaintWidget(m, im, 6)))
	require.NoError(t, checkImage([]string{
		"b",
		"b",
		".",
		".",
		".",
		".",
	}, PaintWidget(m, im, 7)))
	require.NoError(t, checkImage([]string{
		"b",
		".",
		".",
		".",
		".",
		".",
	}, PaintWidget(m, im, 8)))
	require.NoError(t, checkImage([]string{
		".",
		".",
		".",
		".",
		".",
		".",
	}, PaintWidget(m, im, 9)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "r"}, PaintWidget(m, im, 10)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", "r", "g"}, PaintWidget(m, im, 11)))
	require.NoError(t, checkImage([]string{".", ".", ".", "r", "g", "g"}, PaintWidget(m, im, 12)))
	require.NoError(t, checkImage([]string{".", ".", "r", "g", "g", "b"}, PaintWidget(m, im, 13)))
	require.NoError(t, checkImage([]string{".", "r", "g", "g", "b", "b"}, PaintWidget(m, im, 14)))
	require.NoError(t, checkImage([]string{"r", "g", "g", "b", "b", "b"}, PaintWidget(m, im, 15)))
	assert.Equal(t, 16, m.FrameCount(im))

	// Negative OffsetStart
	m.OffsetStart = -2
	require.NoError(t, checkImage([]string{"g", "b", "b", "b", "b", "."}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"b", "b", "b", "b", ".", "."}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"b", "b", "b", ".", ".", "."}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{"b", "b", ".", ".", ".", "."}, PaintWidget(m, im, 3)))
	require.NoError(t, checkImage([]string{"b", ".", ".", ".", ".", "."}, PaintWidget(m, im, 4)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "."}, PaintWidget(m, im, 5)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "r"}, PaintWidget(m, im, 6)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", "r", "g"}, PaintWidget(m, im, 7)))
	require.NoError(t, checkImage([]string{".", ".", ".", "r", "g", "g"}, PaintWidget(m, im, 8)))
	require.NoError(t, checkImage([]string{".", ".", "r", "g", "g", "b"}, PaintWidget(m, im, 9)))
	require.NoError(t, checkImage([]string{".", "r", "g", "g", "b", "b"}, PaintWidget(m, im, 10)))
	require.NoError(t, checkImage([]string{"r", "g", "g", "b", "b", "b"}, PaintWidget(m, im, 11)))
	assert.Equal(t, 12, m.FrameCount(im))

	// Overly negative OffsetStart is truncated to child width
	m.OffsetStart = -1000
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "."}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "r"}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", "r", "g"}, PaintWidget(m, im, 2)))
	assert.Equal(t, 7, m.FrameCount(im))
	m.OffsetStart = -7
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "."}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "r"}, PaintWidget(m, im, 1)))
	assert.Equal(t, 7, m.FrameCount(im))
	m.OffsetStart = -8
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "."}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "r"}, PaintWidget(m, im, 1)))
	assert.Equal(t, 7, m.FrameCount(im))
	m.OffsetStart = -6
	require.NoError(t, checkImage([]string{"b", ".", ".", ".", ".", "."}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "."}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "r"}, PaintWidget(m, im, 2)))
	assert.Equal(t, 8, m.FrameCount(im))

	// OffsetEnd affects the final position of the child
	m.OffsetStart = 0
	m.OffsetEnd = 2
	require.NoError(t, checkImage([]string{"r", "g", "g", "b", "b", "b"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"g", "g", "b", "b", "b", "b"}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"g", "b", "b", "b", "b", "."}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{"b", "b", "b", "b", ".", "."}, PaintWidget(m, im, 3)))
	require.NoError(t, checkImage([]string{"b", "b", "b", ".", ".", "."}, PaintWidget(m, im, 4)))
	require.NoError(t, checkImage([]string{"b", "b", ".", ".", ".", "."}, PaintWidget(m, im, 5)))
	require.NoError(t, checkImage([]string{"b", ".", ".", ".", ".", "."}, PaintWidget(m, im, 6)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "."}, PaintWidget(m, im, 7)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "r"}, PaintWidget(m, im, 8)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", "r", "g"}, PaintWidget(m, im, 9)))
	require.NoError(t, checkImage([]string{".", ".", ".", "r", "g", "g"}, PaintWidget(m, im, 10)))
	require.NoError(t, checkImage([]string{".", ".", "r", "g", "g", "b"}, PaintWidget(m, im, 11)))
	assert.Equal(t, 12, m.FrameCount(im))
	require.NoError(t, checkImage([]string{".", ".", "r", "g", "g", "b"}, PaintWidget(m, im, 12)))
	require.NoError(t, checkImage([]string{".", ".", "r", "g", "g", "b"}, PaintWidget(m, im, 13)))
	require.NoError(t, checkImage([]string{".", ".", "r", "g", "g", "b"}, PaintWidget(m, im, 1024)))

	// Negative offset places child outside of marquee
	m.OffsetEnd = -4
	require.NoError(t, checkImage([]string{"r", "g", "g", "b", "b", "b"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"g", "g", "b", "b", "b", "b"}, PaintWidget(m, im, 1)))
	require.NoError(t, checkImage([]string{"g", "b", "b", "b", "b", "."}, PaintWidget(m, im, 2)))
	require.NoError(t, checkImage([]string{".", ".", ".", "r", "g", "g"}, PaintWidget(m, im, 10)))
	require.NoError(t, checkImage([]string{".", ".", "r", "g", "g", "b"}, PaintWidget(m, im, 11)))
	require.NoError(t, checkImage([]string{".", "r", "g", "g", "b", "b"}, PaintWidget(m, im, 12)))
	require.NoError(t, checkImage([]string{"r", "g", "g", "b", "b", "b"}, PaintWidget(m, im, 13)))
	require.NoError(t, checkImage([]string{"g", "g", "b", "b", "b", "b"}, PaintWidget(m, im, 14)))
	require.NoError(t, checkImage([]string{"g", "b", "b", "b", "b", "."}, PaintWidget(m, im, 15)))
	require.NoError(t, checkImage([]string{"b", "b", "b", "b", ".", "."}, PaintWidget(m, im, 16)))
	require.NoError(t, checkImage([]string{"b", "b", "b", ".", ".", "."}, PaintWidget(m, im, 17)))
	assert.Equal(t, 18, m.FrameCount(im))
	require.NoError(t, checkImage([]string{"b", "b", "b", ".", ".", "."}, PaintWidget(m, im, 18)))
	require.NoError(t, checkImage([]string{"b", "b", "b", ".", ".", "."}, PaintWidget(m, im, 19)))
	require.NoError(t, checkImage([]string{"b", "b", "b", ".", ".", "."}, PaintWidget(m, im, 1024)))

	// OffsetEnd >= width means it doesn't scroll back
	m.OffsetEnd = 6
	require.NoError(t, checkImage([]string{"r", "g", "g", "b", "b", "b"}, PaintWidget(m, im, 0)))
	require.NoError(t, checkImage([]string{"b", ".", ".", ".", ".", "."}, PaintWidget(m, im, 6)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "."}, PaintWidget(m, im, 7)))
	assert.Equal(t, 8, m.FrameCount(im))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "."}, PaintWidget(m, im, 8)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "."}, PaintWidget(m, im, 9)))
	require.NoError(t, checkImage([]string{".", ".", ".", ".", ".", "."}, PaintWidget(m, im, 1024)))
}
