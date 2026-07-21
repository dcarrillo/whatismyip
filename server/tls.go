package server

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net/http"
)

type TLS struct {
	server   *http.Server
	handler  http.Handler
	ctx      context.Context
	addr     string
	crtPath  string
	keyPath  string
	timeouts ServerTimeouts
}

func NewTLSServer(ctx context.Context, handler http.Handler, addr, crt, key string, timeouts ServerTimeouts) *TLS {
	return &TLS{
		handler:  handler,
		ctx:      ctx,
		addr:     addr,
		crtPath:  crt,
		keyPath:  key,
		timeouts: timeouts,
	}
}

func (t *TLS) Start() {
	t.server = &http.Server{
		Addr:         t.addr,
		Handler:      t.handler,
		ReadTimeout:  t.timeouts.ReadTimeout,
		WriteTimeout: t.timeouts.WriteTimeout,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	log.Printf("Starting TLS server listening on %s", t.addr)
	go func() {
		if err := t.server.ListenAndServeTLS(t.crtPath, t.keyPath); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
}

func (t *TLS) Stop() {
	log.Print("Stopping TLS server...")
	ctx, cancel := context.WithTimeout(t.ctx, shutdownTimeout)
	defer cancel()
	if err := t.server.Shutdown(ctx); err != nil {
		log.Printf("TLS server forced to shutdown: %s", err)
	}
}
