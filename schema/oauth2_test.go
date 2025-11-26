package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/pixlet/runtime"
)

func TestOAuth2(t *testing.T) {
	const source = `
load("encoding/json.star", "json")
load("schema.star", "schema")
load("assert.star", "assert")

def oauth_handler(params):
    params = json.decode(params)
    return "foobar123"

t = schema.OAuth2(
    id = "auth",
    name = "GitHub",
    desc = "Connect your GitHub account.",
    icon = "github",
    handler = oauth_handler,
    client_id = "the-oauth2-client-id",
    authorization_endpoint = "https://example.com/",
    scopes = [
        "read:user",
    ],
)

assert.eq(t.id, "auth")
assert.eq(t.name, "GitHub")
assert.eq(t.desc, "Connect your GitHub account.")
assert.eq(t.icon, "github")
assert.eq(t.handler("{}"), "foobar123")
assert.eq(t.client_id, "the-oauth2-client-id")
assert.eq(t.authorization_endpoint, "https://example.com/")
assert.eq(t.scopes, ["read:user"])

def main():
    return []

`

	app, err := runtime.NewApplet("oauth2.star", []byte(source), runtime.WithTests(t))
	assert.NoError(t, err)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.NotNil(t, screens)
}
