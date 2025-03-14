//go:build !js && !wasm

package browser

import (
	"log"
	"net/http"
)

func (b *Browser) serveHTTP() error {
	log.Printf("listening at http://%s%s\n", b.addr, b.path)
	return http.ListenAndServe(b.addr, b.r)
}
