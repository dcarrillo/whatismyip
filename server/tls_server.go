package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/dcarrillo/whatismyip/internal/setting"
)

type TLSServer struct {
	server  *http.Server
	handler *http.Handler
	ctx     context.Context
}

func NewTLSServer(ctx context.Context, handler *http.Handler) *TLSServer {
	return &TLSServer{
		handler: handler,
		ctx:     ctx,
	}
}

func (t *TLSServer) Start() {
	t.server = &http.Server{
		Addr:         setting.App.TLSAddress,
		Handler:      *t.handler,
		ReadTimeout:  setting.App.Server.ReadTimeout,
		WriteTimeout: setting.App.Server.WriteTimeout,
	}

	log.Printf("Starting TLS server listening on %s", setting.App.TLSAddress)
	go func() {
		if err := t.server.ListenAndServeTLS(setting.App.TLSCrtPath, setting.App.TLSKeyPath); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	log.Printf("Stopping TLS server...")
}

func (t *TLSServer) Stop() {
	if err := t.server.Shutdown(t.ctx); err != nil {
		log.Printf("TLS server forced to shutdown: %s", err)
	}
}
