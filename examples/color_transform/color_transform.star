load("render.star", "render")

def main():
    return render.Root(
        child = render.Column(
            children = [
                # Row 1: Original, Black silhouette, Grayscale
                render.Row(
                    children = [
                        render.Box(width = 8, height = 8, color = "#f00"),
                        render.Box(width = 2, height = 1),
                        render.ColorTransform(
                            child = render.Box(width = 8, height = 8, color = "#f00"),
                            brightness = 0.0,
                        ),
                        render.Box(width = 2, height = 1),
                        render.ColorTransform(
                            child = render.Box(width = 8, height = 8, color = "#f00"),
                            saturation = 0.0,
                        ),
                    ],
                ),
                render.Box(width = 1, height = 2),

                # Row 2: Semi-transparent, Inverted, Blue tint (on white)
                render.Row(
                    children = [
                        render.ColorTransform(
                            child = render.Box(width = 8, height = 8, color = "#f00"),
                            opacity = 0.5,
                        ),
                        render.Box(width = 2, height = 1),
                        render.ColorTransform(
                            child = render.Box(width = 8, height = 8, color = "#f00"),
                            invert = True,
                        ),
                        render.Box(width = 2, height = 1),
                        render.ColorTransform(
                            child = render.Box(width = 8, height = 8, color = "#fff"),
                            tint = "#00f",
                        ),
                    ],
                ),
                render.Box(width = 1, height = 2),

                # Row 3: Hue rotation
                render.Row(
                    children = [
                        render.ColorTransform(
                            child = render.Box(width = 8, height = 8, color = "#f00"),
                            hue_rotate = 90,
                        ),
                    ],
                ),
            ],
        ),
    )
