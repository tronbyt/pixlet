# Pixlet Fonts

Pixlet offers a couple of different fonts for rendering text. Which
one to use depends on the information being presented and the amount
of space available when rendering. Pixlet's `Text` Widget attempts to
draw the font on the baseline, i.e. offset upwards by the font's
descent. And if that sentence made no sense to you, read then next
section. =)

## Typography and Unicode refresher

Our most important Unicode block is Basic Latin (a.k.a. ASCII). We
also need at least the Latin-1 Supplement block (a.k.a. ISO 8859,
"Latin-1") to be able to render symbols like Ñ, Ä and Ö. These matter
a lot when displaying names of people, locations, etc.

The definitions in this graphic are helpful:

![Typography line terms](img/Typography_Line_Terms.png)

In addition to these, a glyph's _advance_ is the distance between its
"origin point", i.e. the leftmost point of the baseline, and the
following glyph.

A _cap height_ of 6 and a _descent_ of 1 is sufficient to fit both
upper and lower case A-Z, as well as the most common special
characters. From fiddling with fonts a bit, this seems to be the
minimum.

To cover Latin-1 Supplement, we need more space. At least one
additional pixel in the _ascent_ for characters with diacritics to be
legible.

## Unicode Emoji Support

The Pixlet Text widget automatically detects and renders Unicode emojis alongside regular text. You can include emojis directly in your text strings with any font:

```python
# Single emoji
render.Text("😀", font="6x10")

# Multiple emojis
render.Text("😀😂😍😉😎", font="5x8")

# Mixed text and emojis
render.Text("Hello 😀 World!", font="tb-8")

# Flag emojis
render.Text("🇺🇸🇬🇧🇫🇷", font="6x10")
```

### How Emoji Rendering Works

- **Automatic Detection**: The Text widget automatically detects Unicode emoji sequences
- **Built-in Support**: Uses Pixlet's built-in emoji sprite sheet with 1753+ emojis
- **Seamless Integration**: Emojis render alongside text using any font
- **Standard Size**: All emojis render at 10x10 pixels regardless of text font

### Scalable Emojis

For large emojis or standalone emoji graphics, use the dedicated `render.Emoji` widget:

```python
# Create emojis at any size
render.Emoji(emoji="🚀", height=32)   # 32px tall rocket
render.Emoji(emoji="⚡", height=64)   # 64px tall lightning
```

The Emoji widget provides:
- **Custom sizing**: Any height from 1 to hundreds of pixels
- **High quality scaling**: Smooth interpolation for large sizes
- **Performance**: Optimized rendering at any scale

For a complete list of supported emojis and detailed usage, see [emoji.md](emoji.md).

## The fonts

Note that all of these are free or public domain fonts created by
others. Attribution is given for each below.

### tb-8

![tb-8 example](img/tb-8.gif)

A modified version of 5x8. This font has variable advance (i.e. it's
not mono-spaced) and slightly tweaked glyphs for improved
legibility. All digits are monospace with advance 5.

Like 5x8, this covers Basic Latin and Latin-1 Supplement, plus some
other latin glyphs.

All digits in tb-8 are monospaced 5x6, so rendering them with a
baseline offset of -1 and a height of 6 is perfectly fine and renders
the full glyphs without cropping.

Common numerical symbols ('+', '-', '/', '*', '=', '%', '.') also fit
in height 6 with offset -1. These are however all variable-width. Note
that ',' has descent 1 and won't fit in this case.

Currency symbols require full height in the general case, but euro
sign '€' and dollar sign '$' have no descent and fit in 7 pixels.

- Advance: 2-6
- Height: 8
- Cap height: 6
- Ascent: 7
- Descent: 1

### Dina_r400-6
By [Jørgen Ibsen](https://www.dcmembers.com/jibsen/download/61/)

![Dina_r400-6 example](img/Dina_r400-6.gif)

Covers Basic Latin and Latin-1 Supplement, but nothing beyond
that. 256 code points.

- Advance: 6
- Height: 10
- Cap height: 6
- Ascent: 8
- Descent: 2

### 5x8
By [Markus Kuhn](http://www.cl.cam.ac.uk/~mgk25/).

![5x8 example](img/5x8.gif)

Covers Basic Latin, Latin-1 Supplement, a ton of other Latin code
blocks. 1426 code points in total.

- Advance: 5
- Height: 8
- Cap height: 6
- Ascent: 7
- Descent: 1

### 6x13
By [Markus Kuhn](http://www.cl.cam.ac.uk/~mgk25/).

![6x16 example](img/6x13.gif)

Covers Basic Latin, Latin1- Supplement, and a bazillion other glyphs,
including the runic code block. 4121 code points in total.

- Advance: 6
- Height: 13
- Cap height: 9
- Ascent: 11
- Descent: 2

### 6x10
By [olikraus](https://github.com/olikraus/u8g2/wiki).

![6x16 example](img/6x10.gif)

A medium, monospace, bitmap font with a size of 6x10 pixels. Ideal for fixed-width applications.

- Advance: 6
- Height: 10
- Cap height: 7
- Ascent: 8
- Descent: 2

### 6x10-rounded
Original font By [olikraus](https://github.com/olikraus/u8g2/wiki) & modified by [SynthSolar](https://github.com/SynthSolar).

![6x16-rounded example](img/6x10-rounded.gif)

A medium, monospace, rounded, bitmap font with a size of 6x10 pixels. Ideal for fixed-width applications.

- Advance: 6
- Height: 10
- Cap height: 7
- Ascent: 8
- Descent: 2

### 10x20
By [Markus Kuhn](http://www.cl.cam.ac.uk/~mgk25/).

![10x20 example](img/10x20.gif)

Covers Basic Latin, Latin1- Supplement, and a bazillion other glyphs,
including the runic code block. 4121 code points in total.

- Advance: 10
- Height: 20
- Cap height: 13
- Ascent: 16
- Descent: 4

### tom-thumb
By [Robey Pointer](https://robey.lag.net/2010/01/23/tiny-monospace-font.html)

![tom-thumb example](img/tom-thumb.gif)

A very tiny, monospace, bitmap font. It's a 4x6 font with 3x5 usable pixels.
This font is great for a really tiny font that also supports upper and lower
cased characters.

- Advance: 4
- Height: 6
- Cap height: 4
- Ascent: 5
- Descent: 1

### CG-pixel-3x5-mono
By [Ilmari Karonen](https://vyznev.net/)

![CG-pixel-3x5-mono example](img/CG-pixel-3x5-mono.gif)

This font is a true 3x5 font which only occupies 5 pixels on a display. Check
out the 4x5 version if you have the width to spare.

- Advance: 4
- Height: 5
- Cap height: 4
- Ascent: 5
- Descent: 0

### CG-pixel-4x5-mono
By [Ilmari Karonen](https://vyznev.net/)

![CG-pixel-4x5-mono example](img/CG-pixel-4x5-mono.gif)

This font is a true 4x5 font which only occupies 5 pixels on a display. Check
out the 3x5 version if you are also constrained on width.

- Advance: 5
- Height: 5
- Cap height: 4
- Ascent: 5
- Descent: 0

### Terminus Variable
By [Dimitar Toshkov Zhekov](https://terminus-font.sourceforge.net/)

![Terminus Variable](img/terminus-variable.png)

Terminus Variable is a family of bitmap fonts that supports multiple pixel sizes. Each size ships in a normal weight, with matching light weight variants at 14px and larger. All styles include 1356 glyphs (ISO10646-1 coverage). Numeric glyphs, the colon, and the space character share the same width, so clock layouts and other tabular number displays stay aligned.

#### terminus-12
- Advance: 2-7
- Height: 12
- Cap height: 8
- Ascent: 10
- Descent: 2

#### terminus-14 / terminus-14-light
- Advance: 3-10
- Height: 14
- Cap height: 10
- Ascent: 12
- Descent: 2

#### terminus-16 / terminus-16-light
- Advance: 3-10
- Height: 16
- Cap height: 10
- Ascent: 12
- Descent: 4

#### terminus-18 / terminus-18-light
- Advance: 3-12
- Height: 18
- Cap height: 12
- Ascent: 15
- Descent: 3

#### terminus-20 / terminus-20-light
- Advance: 3-12
- Height: 20
- Cap height: 13
- Ascent: 16
- Descent: 4

#### terminus-22 / terminus-22-light
- Advance: 3-13
- Height: 22
- Cap height: 14
- Ascent: 17
- Descent: 5

#### terminus-24 / terminus-24-light
- Advance: 3-14
- Height: 24
- Cap height: 15
- Ascent: 19
- Descent: 5

#### terminus-28 / terminus-28-light
- Advance: 4-16
- Height: 28
- Cap height: 18
- Ascent: 22
- Descent: 6

#### terminus-32 / terminus-32-light
- Advance: 4-18
- Height: 32
- Cap height: 20
- Ascent: 26
- Descent: 6
