package browser

import (
	"context"
	"errors"
	"log"
	"net/http"
)

func (b *Browser) serveHTTP(ctx context.Context) error {
	server := &http.Server{
		Addr:    b.addr,
		Handler: b.r,
	}
	go func() {
		<-ctx.Done()
		_ = server.Close()
	}()

	log.Printf("listening at http://%s%s\n", b.addr, b.path)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
