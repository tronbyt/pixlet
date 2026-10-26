package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tronbyt/pixlet/runtime"
)

// TestOAuth2 checks the attributes of a field in the original Tidbyt mode,
// with a client ID and a handler.
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

	app, err := runtime.NewApplet(t.Context(), "oauth2.star", []byte(source), runtime.WithTests(t))
	require.NoError(t, err)

	screens, err := app.Run(t.Context())
	require.NoError(t, err)
	assert.NotNil(t, screens)
}

// TestOAuth2ServerManaged checks the attributes of a server-managed field:
// the new arguments read back, and the unset handler and client ID read as
// None and "".
func TestOAuth2ServerManaged(t *testing.T) {
	const source = `
load("schema.star", "schema")
load("assert.star", "assert")

t = schema.OAuth2(
    id = "auth",
    name = "GitHub",
    desc = "Connect your GitHub account.",
    icon = "github",
    authorization_endpoint = "https://github.com/login/oauth/authorize",
    token_endpoint = "https://github.com/login/oauth/access_token",
    authorization_params = {"prompt": "consent"},
    scopes = [
        "read:user",
    ],
)

assert.eq(t.id, "auth")
assert.eq(t.token_endpoint, "https://github.com/login/oauth/access_token")
assert.eq(t.authorization_params, {"prompt": "consent"})
assert.eq(t.handler, None)
assert.eq(t.client_id, "")

def main():
    return []

`

	app, err := runtime.NewApplet(t.Context(), "oauth2.star", []byte(source), runtime.WithTests(t))
	require.NoError(t, err)

	screens, err := app.Run(t.Context())
	require.NoError(t, err)
	assert.NotNil(t, screens)
}

// TestOAuth2AuthorizationParamsMustBeStrings checks that a non-string value
// in authorization_params fails when the field is created, since it is sent
// as a URL query parameter.
func TestOAuth2AuthorizationParamsMustBeStrings(t *testing.T) {
	const source = `
load("schema.star", "schema")

t = schema.OAuth2(
    id = "auth",
    name = "GitHub",
    desc = "Connect your GitHub account.",
    icon = "github",
    authorization_endpoint = "https://github.com/login/oauth/authorize",
    token_endpoint = "https://github.com/login/oauth/access_token",
    authorization_params = {"max_age": 3600},
    scopes = ["read:user"],
)

def main():
    return []
`

	_, err := runtime.NewApplet(t.Context(), "oauth2.star", []byte(source), runtime.WithTests(t))
	require.ErrorContains(t, err, "authorization_params")
}
