# ColorTransform Widget Example

This example demonstrates the `render.ColorTransform` widget, which allows you to apply various color and appearance transformations to any widget.

## Transformations

The ColorTransform widget supports the following transformations:

### Brightness
Controls the brightness of the widget.
- `0.0` = completely black (useful for creating silhouettes)
- `1.0` = normal brightness (default)
- `>1.0` = brighter

```starlark
render.ColorTransform(
    child = render.Image(src = icon_data),
    brightness = 0.0,  # Black silhouette
)
```

### Saturation
Controls the color saturation.
- `0.0` = grayscale
- `1.0` = normal saturation (default)
- `>1.0` = more saturated

```starlark
render.ColorTransform(
    child = render.Image(src = icon_data),
    saturation = 0.0,  # Grayscale
)
```

### Hue Rotation
Rotates the hue of all colors by the specified degrees (0-360).

```starlark
render.ColorTransform(
    child = render.Image(src = icon_data),
    hue_rotate = 180,  # Shift hue by 180 degrees
)
```

### Opacity
Controls the transparency of the widget.
- `0.0` = completely transparent
- `1.0` = completely opaque (default)

```starlark
render.ColorTransform(
    child = render.Image(src = icon_data),
    opacity = 0.5,  # Semi-transparent
)
```

### Invert
Inverts all colors.

```starlark
render.ColorTransform(
    child = render.Image(src = icon_data),
    invert = True,
)
```

### Tint
Applies a color overlay (multiply blend mode).

```starlark
render.ColorTransform(
    child = render.Image(src = icon_data),
    tint = "#ff0000",  # Red tint
)
```

## Combining Transformations

You can combine multiple transformations:

```starlark
render.ColorTransform(
    child = render.Image(src = icon_data),
    brightness = 0.8,
    saturation = 0.5,
    opacity = 0.7,
    tint = "#0088ff",
)
```

## Use Cases

- **Silhouettes**: Set `brightness = 0.0` to create black silhouettes of icons
- **Disabled state**: Use `saturation = 0.0` and `opacity = 0.5` for a "disabled" appearance
- **Theming**: Use `tint` to apply brand colors to monochrome icons
- **Highlighting**: Use `brightness > 1.0` to make elements stand out
- **Color correction**: Use `hue_rotate` to adjust colors without changing the image
