package server

import (
	"context"
	"net"
	"net/http"
	"time"
)

// RunHTTPServer runs standard HTTP server at given address,
// with graceful shutdown when context is cancelled
func RunHTTPServer(ctx context.Context, addr string, handler http.Handler) error {
	srv := http.Server{
		Addr: addr,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		Handler:           handler,
		ReadHeaderTimeout: time.Minute,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(ctx)
	}()

	if err := srv.ListenAndServe(); err != nil && ctx.Err() == nil {
		return err
	}
	return nil
}
