load("http.star", "http")
load("icon.png", icon = "file")
load("render.star", "render")

COINDESK_PRICE_URL = "https://min-api.cryptocompare.com/data/price?fsym=BTC&tsyms=USD"

BTC_ICON = icon.readall()

def main():
    rep = http.get(COINDESK_PRICE_URL, ttl_seconds = 240)
    if rep.status_code != 200:
        fail("Coindesk request failed with status %d", rep.status_code)
    rate = rep.json()["USD"]

    return render.Root(
        child = render.Box(
            render.Row(
                expanded = True,
                main_align = "space_evenly",
                cross_align = "center",
                children = [
                    render.Image(src = BTC_ICON),
                    render.Text("$%d" % rate),
                ],
            ),
        ),
    )
