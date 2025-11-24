# Authoring apps for 2x resolution

Tidbyt has a 64x32 pixel display. However, `pixlet` is able to render apps at higher resolutions. In particular, the `-2` flag will render apps at 2x, i.e. at a 128x64 resolution:

```
pixlet render -2 my_app.star
pixlet serve -2 my_app.star
```

Note that `pixlet render -2` will automatically add a `@2x` suffix to the filename if none is specified.

When authoring apps, it's highly recommended to make them look good at both 1x and 2x resolutions. This guide outlines a few things to keep in mind to make that happen.

## Enabling 2x support in the manifest

To enable 2x support for your app, you must set the `supports2x` flag to `true` in your app's manifest file. This flag is read by `pixlet` and indicates that the app is designed to handle 2x resolutions.

For example:
```yaml
id: my-app
name: My App
summary: An example app
desc: This is a longer description.
author: Your Name
supports2x: true
```

## Use `canvas` for dimensions

The `canvas` object provides information about the output resolution. Use it to fetch the canvas dimensions rather than hard coding `64` and `32` for width and height:

```starlark
load("render.star", "render", "canvas")

def main():
    return render.Root(
        child = render.Box(
            width = canvas.width(),
            height = canvas.height(),
            color = "#f00",
        ),
    )
```

## Scaling dimensions

A common pattern to scale dimensions is to use `canvas.is2x()` to determine if the app is being rendered at 2x resolution:

```starlark
load("render.star", "render", "canvas")

def main():
    scale = 2 if canvas.is2x() else 1
    return render.Root(
        child = render.Box(
            width = 12 * scale,
            height = 8 * scale,
            color = "#f00",
        ),
    )
```

## Font selection

When rendering at 2x, it may be appropriate to choose a larger font. In other cases, it may be better to show more content. When no font is specified for `render.Text`, the default font is `tb-8` on 1x displays and `terminus-16` on 2x.

Note that not all fonts have alternatives at exactly 2x the size. Sometimes, one needs to choose a font with double the height, in other cases, double the width is better. Use `pixlet community list-fonts` to see available fonts and their dimensions.

## Images

For images, you might want to load a higher-resolution asset when rendering at 2x. This can be done by dynamically choosing the image source based on `canvas.is2x()`.

```starlark
load("render.star", "render", "canvas")
load("images/myimage.png", MY_IMAGE_1X = "file")
load("images/myimage@2x.png", MY_IMAGE_2X = "file")

def main():
    image = MY_IMAGE_2X if canvas.is2x() else MY_IMAGE_1X

    return render.Root(
        child = render.Image(src = image.readall()),
    )
```

## Animation speed

Since `render.Marquee` moves its child by 1 pixel per frame, animations can appear slower on a 2x canvas if other elements are scaled up. To maintain a similar perceived speed, you may need to double the animation speed by halving the frame delay in `render.Root`. For example, you could set the delay like this:

```starlark
load("render.star", "render", "canvas")

def main():
    delay = 25 if canvas.is2x() else 50
    return render.Root(
        child = render.Text("Hello, World!"),
        delay = delay,
    )
```
