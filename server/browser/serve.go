package browser

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

func (b *Browser) serveHTTP(ctx context.Context) error {
	server := &http.Server{
		Addr:    net.JoinHostPort(b.host, strconv.Itoa(b.port)),
		Handler: b.r,
	}
	go func() {
		<-ctx.Done()
		_ = server.Close()
	}()

	u := &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(b.host, strconv.Itoa(b.port)),
		Path:   b.path,
	}

	slog.Info("Starting HTTP server", "address", u.String())
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
