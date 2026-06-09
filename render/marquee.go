package render

import (
	"image"

	"github.com/tronbyt/gg"
)

// MarqueeMaxContentScale bounds how much larger than the Marquee's own width
// (horizontal) or height (vertical) the scrolled child may be. It caps the
// drawing context handed to the child so a runaway app can't allocate an
// enormous canvas, while still letting long content (e.g. a WrappedText)
// scroll fully.
//
// Raised from 10 to 20 (2026) now that target devices tolerate larger WebP
// payloads. This DOUBLES the maximum scrollable length — but only the *spatial*
// ceiling. Two further limits decide whether that extra content is actually
// reached on screen, and you must raise them too to benefit:
//
//   - Temporal (the usual blocker): the WebP/GIF encoder stops adding frames
//     once cumulative frame time reaches maxDuration (default 15s). A Marquee
//     scrolls 1px/frame, so a child longer than
//     dwell_seconds * 1000 / root_delay_ms pixels is truncated mid-scroll
//     regardless of this cap — the rest is never encoded into the file, so the
//     device just loops the partial scroll. Raise maxDuration via the
//     `--max-duration` / `-d` flag on `pixlet render` / `pixlet serve`, or via
//     loader.WithMaxDuration(...) when embedding pixlet as a library. In
//     tronbyt-server this value is the per-app DisplayTime (falling back to the
//     device DefaultInterval), set in the App Config screen — no code change
//     needed there.
//   - Frame count: render.DefaultMaxFrameCount (2000) is a hard upper bound on
//     total frames; not normally reached at these sizes.
//
// Bottom line: to use the larger ceiling, pair it with a longer dwell time
// (more seconds of scroll baked into the file) and/or a smaller Root delay
// (more pixels scrolled per second).
const MarqueeMaxContentScale = 20

// Marquee scrolls its child horizontally or vertically.
//
// The `scroll_direction` will be 'horizontal' and will scroll from right
// to left if left empty, if specified as 'vertical' the Marquee will
// scroll from bottom to top.
//
// In horizontal mode the height of the Marquee will be that of its child,
// but its `width` must be specified explicitly. In vertical mode the width
// will be that of its child but the `height` must be specified explicitly.
//
// If the child's width fits fully, it will not scroll.
//
// The `offset_start` and `offset_end` parameters control the position
// of the child in the beginning and the end of the animation.
//
// Alignment for a child that fits fully along the horizontal/vertical axis is controlled by passing
// one of the following `align` values:
// - `"start"`: place child at the left/top
// - `"end"`: place child at the right/bottom
// - `"center"`: place child at the center
//
// Example:
//
//	render.Marquee(
//	    width=64,
//	    child=render.Text("this won't fit in 64 pixels"),
//	    offset_start=5,
//	    offset_end=32,
//	)
type Marquee struct {
	// Widget to potentially scroll
	Child Widget `starlark:"child,required"`
	// Width of the Marquee, required for horizontal
	Width int `starlark:"width"`
	// Height of the Marquee, required for vertical
	Height int `starlark:"height"`
	// Position of child at beginning of animation
	OffsetStart int `starlark:"offset_start"`
	// Position of child at end of animation
	OffsetEnd int `starlark:"offset_end"`
	// Direction to scroll, 'vertical' or 'horizontal', default is horizontal
	ScrollDirection string `starlark:"scroll_direction"`
	// Alignment when contents fit on screen, 'start', 'center' or 'end', default is start
	Align string `starlark:"align"`
	// Delay the scroll of the animation by a certain number of frames, default is 0
	Delay int `starlark:"delay"`
}

func (m Marquee) PaintBounds(bounds image.Rectangle, frameIdx int) image.Rectangle {
	var cb image.Rectangle

	if m.isVertical() {
		cb = m.Child.PaintBounds(image.Rect(0, 0, bounds.Dx(), m.Height*MarqueeMaxContentScale), 0)
	} else {
		cb = m.Child.PaintBounds(image.Rect(0, 0, m.Width*MarqueeMaxContentScale, bounds.Dy()), 0)
	}

	if m.isVertical() {
		return image.Rect(0, 0, cb.Dx(), m.Height)
	} else {
		return image.Rect(0, 0, m.Width, cb.Dy())
	}
}

func (m Marquee) FrameCount(bounds image.Rectangle) int {
	var cb image.Rectangle
	var cw int
	var size int
	if m.isVertical() {
		cb = m.Child.PaintBounds(image.Rect(0, 0, bounds.Dx(), m.Height*MarqueeMaxContentScale), 0)
		cw = cb.Dy()
		size = m.Height
	} else {
		cb = m.Child.PaintBounds(image.Rect(0, 0, m.Width*MarqueeMaxContentScale, bounds.Dy()), 0)
		cw = cb.Dx()
		size = m.Width
	}

	if cw <= size {
		return 1
	}

	offstart := max(m.OffsetStart, -cw)

	offend := max(m.OffsetEnd, -cw)

	delay := m.Delay
	// If start and end offsets are identical, do not
	// repeat these identical frames after another.
	if offstart == offend {
		return cw + offstart + size - offend + delay
	} else {
		return cw + offstart + size - offend + 1 + delay
	}
}

func (m Marquee) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	var cb image.Rectangle
	var cw int
	var size int
	if m.isVertical() {
		// We'll only scroll frame 0 of the child. Scrolling an
		// animation would be madness.
		cb = m.Child.PaintBounds(image.Rect(0, 0, bounds.Dx(), m.Height*MarqueeMaxContentScale), 0)
		cw = cb.Dy()
		size = m.Height
	} else {
		cb = m.Child.PaintBounds(image.Rect(0, 0, m.Width*MarqueeMaxContentScale, bounds.Dy()), 0)
		cw = cb.Dx()
		size = m.Width
	}

	offstart := max(m.OffsetStart, -cw)

	offend := max(m.OffsetEnd, -cw)

	delay := m.Delay
	loopIdx := cw + offstart + delay
	endIdx := cw + offstart + size - offend + delay

	align := 0.0 //default is align="start"
	var offset int
	if cw <= size {
		// child fits entirely and we don't want to scroll it anyway
		offset = 0

		//modify alignment
		switch m.Align {
		case "center":
			align = 0.5
			offset = size / 2
		case "end":
			align = 1.0
			offset = size
		}
	} else if frameIdx <= delay {
		// delay the scrolling for the number of frames specified by delay
		offset = offstart
	} else if frameIdx <= loopIdx {
		// first scroll child out of view
		offset = offstart - frameIdx + delay
	} else if frameIdx <= endIdx {
		// then, scroll back into view
		offset = offend + (endIdx - frameIdx)
	} else {
		// if more than FrameCount frames are requested,
		// freeze marquee at final frame
		offset = offend
	}

	pb := m.PaintBounds(bounds, frameIdx)

	if m.isVertical() {
		offset -= int(align * float64(cb.Dy()))
		dc.Push()
		dc.DrawRectangle(0, 0, float64(pb.Dx()), float64(pb.Dy()))
		dc.Clip()
		dc.Translate(0, float64(offset))
		m.Child.Paint(dc, image.Rect(0, 0, bounds.Dx(), m.Height*MarqueeMaxContentScale), 0)
		dc.Pop()
	} else {
		offset -= int(align * float64(cb.Dx()))
		dc.Push()
		dc.DrawRectangle(0, 0, float64(pb.Dx()), float64(pb.Dy()))
		dc.Clip()
		dc.Translate(float64(offset), 0)
		m.Child.Paint(dc, image.Rect(0, 0, m.Width*MarqueeMaxContentScale, bounds.Dy()), 0)
		dc.Pop()
	}
}

func (m Marquee) isVertical() bool {
	return m.ScrollDirection == "vertical"
}
