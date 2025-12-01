package runtime

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/tink/go/hybrid"
	"github.com/google/tink/go/keyset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyAEAD struct{}

func (a *dummyAEAD) Encrypt(plaintext []byte, additionalData []byte) ([]byte, error) {
	return plaintext, nil
}

func (a *dummyAEAD) Decrypt(ciphertext []byte, additionalData []byte) ([]byte, error) {
	return ciphertext, nil
}

func TestSecretDecrypt(t *testing.T) {
	plaintext := "h4x0rrszZ!!"

	// make a test decryption key
	dummyKEK := &dummyAEAD{}
	khPriv, err := keyset.NewHandle(hybrid.ECIESHKDFAES128CTRHMACSHA256KeyTemplate())
	require.NoError(t, err)

	privJSON := &bytes.Buffer{}
	err = khPriv.Write(keyset.NewJSONWriter(privJSON), dummyKEK)
	require.NoError(t, err)

	decryptionKey := &SecretDecryptionKey{
		EncryptedKeysetJSON: privJSON.Bytes(),
		KeyEncryptionKey:    dummyKEK,
	}

	// get the corresponding public key and serialize it
	khPub, err := khPriv.Public()
	require.NoError(t, err)

	pubJSON := &bytes.Buffer{}
	err = khPub.WriteWithNoSecrets(keyset.NewJSONWriter(pubJSON))
	require.NoError(t, err)

	// encrypt the secret using app ID
	encrypted, err := (&SecretEncryptionKey{
		PublicKeysetJSON: pubJSON.Bytes(),
	}).Encrypt("testid", plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, encrypted, "")

	src := fmt.Sprintf(`
load("render.star", "render")
load("schema.star", "schema")
load("secret.star", "secret")
load("assert.star", "assert")

EXPECTED_PLAINTEXT = "%s"
ENCRYPTED = "%s"
DECRYPTED = secret.decrypt(ENCRYPTED)

def main():
	assert.eq(DECRYPTED, EXPECTED_PLAINTEXT)
	return render.Root(child=render.Box())
`, plaintext, encrypted)

	app, err := NewApplet("testid", []byte(src), WithSecretDecryptionKey(decryptionKey), WithTests(t))
	require.NoError(t, err)

	roots, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(roots))
}

func TestSecretDoesntDecryptWithoutKey(t *testing.T) {
	plaintext := "h4x0rrszZ!!"

	// make a test decryption key
	dummyKEK := &dummyAEAD{}
	khPriv, err := keyset.NewHandle(hybrid.ECIESHKDFAES128CTRHMACSHA256KeyTemplate())
	require.NoError(t, err)

	privJSON := &bytes.Buffer{}
	err = khPriv.Write(keyset.NewJSONWriter(privJSON), dummyKEK)
	require.NoError(t, err)

	// get the corresponding public key and serialize it
	khPub, err := khPriv.Public()
	require.NoError(t, err)

	pubJSON := &bytes.Buffer{}
	err = khPub.WriteWithNoSecrets(keyset.NewJSONWriter(pubJSON))
	require.NoError(t, err)

	// encrypt the secret
	encrypted, err := (&SecretEncryptionKey{
		PublicKeysetJSON: pubJSON.Bytes(),
	}).Encrypt("test", plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, encrypted, "")

	src := fmt.Sprintf(`
load("render.star", "render")
load("schema.star", "schema")
load("secret.star", "secret")
load("assert.star", "assert")

EXPECTED_PLAINTEXT = "%s"
ENCRYPTED = "%s"
DECRYPTED = secret.decrypt(ENCRYPTED)

def main():
	assert.eq(DECRYPTED, None)
	return render.Root(child=render.Box())
`, plaintext, encrypted)

	app, err := NewApplet("test.star", []byte(src), WithTests(t))
	require.NoError(t, err)

	roots, err := app.Run(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(roots))
}
