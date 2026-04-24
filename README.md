# Pixlet
[![Docs](https://img.shields.io/badge/docs-tronbyt.com-blue?style=flat-square)](https://tronbyt.com)
[![Build & test](https://img.shields.io/github/actions/workflow/status/tronbyt/pixlet/main.yml?branch=main&style=flat-square)](https://github.com/tronbyt/pixlet/actions)
[![Discord Server](https://img.shields.io/discord/1343943766976888914?style=flat-square)](https://tronbyt.com/discord/)
[![GoDoc](https://godoc.org/github.com/tronbyt/pixlet/runtime?status.svg)](https://godoc.org/github.com/tronbyt/pixlet/runtime)

Pixlet is an app runtime and UX toolkit for highly-constrained displays.
We use Pixlet to develop applets for [Tronbyt](https://github.com/tronbyt/server), which has
a 64x32 RGB LED matrix display:

[![Example of a Tronbyt](docs/img/tidbyt_1.png)](https://tronbyt.com)

Apps developed with Pixlet can be served in a browser, rendered as WebP or
GIF animations, or pushed to a physical Tronbyt device.

## Documentation

> Hey! We have a new docs site! Check it out at [tronbyt.com](https://tronbyt.com).

- [Getting started](#getting-started)
- [How it works](#how-it-works)
- [In-depth tutorial](docs/tutorial.md)
- [Widget reference](docs/widgets.md)
- [Animation reference](docs/animation.md)
- [Modules reference](docs/modules.md)
- [Schema reference](docs/schema/schema.md)
- [Our thoughts on authoring apps](docs/authoring_apps.md)
- [Notes on the available fonts](docs/fonts.md)

## Getting started

### Install on macOS

```
brew install pixlet
```

### Install on Linux

Download the `pixlet` binary from [the latest release][1].

Alternatively you can [build from source](docs/BUILD.md).

[1]: https://github.com/tronbyt/pixlet/releases/latest

### Hello, World!

Pixlet applets are written in a simple, Python-like language called
Starlark. Here's the venerable Hello World program:

```starlark
load("render.star", "render")
def main():
    return render.Root(
        child = render.Text("Hello, World!")
    )
```

Render and serve it with:

```console
curl https://raw.githubusercontent.com/tronbyt/pixlet/main/examples/hello_world/hello_world.star | \
  pixlet serve /dev/stdin
```

You can view the result by navigating to [http://localhost:8080][3]:

![](docs/img/tutorial_1.gif)

[3]: http://localhost:8080

## How it works

Pixlet scripts are written in a simple, Python-like language called
[Starlark](https://github.com/google/starlark-go/). The scripts can
retrieve data over HTTP, transform it and use a collection of
_Widgets_ to describe how the data should be presented visually.

The Pixlet CLI runs these scripts and renders the result as a WebP
or GIF animation. You can view the animation in your browser, save
it, or even push it to a Tronbyt device with `pixlet push`.

### Example: A Clock App

This applet accepts a `timezone` parameter and produces a two frame
animation displaying the current time with a blinking ':' separator
between the hour and minute components.

```starlark
load("render.star", "render")
load("time.star", "time")

def main(config):
    timezone = config.get("timezone") or "America/New_York"
    now = time.now().in_location(timezone)

    return render.Root(
        delay = 500,
        child = render.Box(
            child = render.Animation(
                children = [
                    render.Text(
                        content = now.format("3:04 PM"),
                        font = "6x13",
                    ),
                    render.Text(
                        content = now.format("3 04 PM"),
                        font = "6x13",
                    ),
                ],
            ),
        ),
    )
```

Here's the resulting image:

![](docs/img/clock.gif)

### Example: A Bitcoin Tracker

Applets can get information from external data sources. For example,
here is a Bitcoin price tracker:

![](docs/img/tutorial_4.gif)

Read the [in-depth tutorial](docs/tutorial.md) to learn how to
make an applet like this.

## Push to a Tronbyt

If you have a Tronbyt, `pixlet` can push apps directly to it. For example,
to show the Bitcoin tracker on your Tronbyt:

```console
# render the bitcoin example
pixlet render examples/bitcoin/bitcoin.star

# login to your Tronbyt account
pixlet login

# list available Tronbyt devices
pixlet devices

# push to your favorite Tronbyt
pixlet push <YOUR DEVICE ID> examples/bitcoin/bitcoin.webp
```

To get the ID for a device, run `pixlet devices`. Alternatively, you can
open the settings for the device in the Tronbyt app on your phone, and tap **Get API key**.

If all goes well, you should see the Bitcoin tracker appear on your Tronbyt:

![](docs/img/tidbyt_2.jpg)

## Push as an Installation
Pushing an applet to your Tronbyt without an installation ID simply displays your applet one time. If you would like your applet to continously display as part of the rotation, add an installation ID to the push command:

```console
pixlet render examples/bitcoin/bitcoin.star
pixlet push --installation-id <INSTALLATION ID> <YOUR DEVICE ID> examples/bitcoin/bitcoin.webp
```

For example, if we set the `installationID` to "Bitcoin", it would appear in the mobile app as follows:

![](docs/img/mobile_1.jpg)

**Note:** `pixlet render` executes your Starlark code and generates a WebP image. `pixlet push` deploys the generated WebP image to your device. You'll need to repeat this process if you want to keep the app updated. You can also create [Community Apps](https://github.com/tronbyt/apps) that are downloaded to all Tronbyt instances, or create a personal apps repository.
