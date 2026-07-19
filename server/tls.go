package server

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net/http"

	"github.com/dcarrillo/whatismyip/internal/setting"
)

type TLS struct {
	server  *http.Server
	handler http.Handler
	ctx     context.Context
}

func NewTLSServer(ctx context.Context, handler http.Handler) *TLS {
	return &TLS{
		handler: handler,
		ctx:     ctx,
	}
}

func (t *TLS) Start() {
	t.server = &http.Server{
		Addr:         setting.App.TLSAddress,
		Handler:      t.handler,
		ReadTimeout:  setting.App.Server.ReadTimeout,
		WriteTimeout: setting.App.Server.WriteTimeout,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	log.Printf("Starting TLS server listening on %s", setting.App.TLSAddress)
	go func() {
		if err := t.server.ListenAndServeTLS(setting.App.TLSCrtPath, setting.App.TLSKeyPath); err != nil &&
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
