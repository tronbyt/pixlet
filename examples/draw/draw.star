"""
An example app that demonstrates the drawing primitives.
"""

load("math.star", "math")
load("render.star", "render")

def main():
    return render.Root(
        child = render.Stack(
            children = [
                render.Box(
                    width = 64,
                    height = 32,
                    color = "#111",
                ),
                render.Line(
                    x1 = 0,
                    y1 = 0,
                    x2 = 63,
                    y2 = 31,
                    width = 1,
                    color = "#fff",
                ),
                render.Padding(
                    pad = (10, 5, 0, 0),
                    child = render.Polygon(
                        vertices = [(0, 0), (44, 0), (44, 10), (0, 10)],
                        color = "#f0f",
                    ),
                ),
                render.Padding(
                    pad = (22, 6, 0, 0),  # Position arc roughly at center
                    child = render.Arc(
                        x = 10,  # Center relative to widget
                        y = 10,
                        radius = 10,
                        start_angle = 0,
                        end_angle = math.pi * 1.5,
                        width = 3,
                        color = "#0ff",
                    ),
                ),
            ],
        ),
    )
