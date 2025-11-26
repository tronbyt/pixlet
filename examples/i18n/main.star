load("i18n.star", "tr")
load("render.star", "render")

def main(config):
    user = config.get("user", "Tronbyt")
    return render.Root(
        child = render.Text(
            content = tr("hello_user", user),
        ),
    )
