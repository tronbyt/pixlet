load("http.star", "http")
load("render.star", "render")
load("schema.star", "schema")

def main(config):
    token = config.get("auth")
    if not token:
        return render.Root(
            child = render.WrappedText("Sign in to GitHub in the app settings"),
        )

    res = http.get(
        url = "https://api.github.com/user",
        headers = {
            "Accept": "application/vnd.github+json",
            "Authorization": "Bearer " + token,
        },
        ttl_seconds = 300,
    )
    if res.status_code != 200:
        return render.Root(
            child = render.WrappedText("GitHub error %d" % res.status_code),
        )

    return render.Root(
        child = render.Marquee(
            width = 64,
            child = render.Text(res.json()["login"]),
        ),
    )

def get_schema():
    return schema.Schema(
        version = "1",
        fields = [
            schema.OAuth2(
                id = "auth",
                name = "GitHub",
                desc = "Connect your GitHub account.",
                icon = "github",
                authorization_endpoint = "https://github.com/login/oauth/authorize",
                token_endpoint = "https://github.com/login/oauth/access_token",
                scopes = [
                    "read:user",
                ],
            ),
        ],
    )
