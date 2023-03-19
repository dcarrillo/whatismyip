package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/dcarrillo/whatismyip/internal/setting"
)

type TCPServer struct {
	server  *http.Server
	handler *http.Handler
	ctx     context.Context
}

func NewTCPServer(ctx context.Context, handler *http.Handler) TCPServer {
	return TCPServer{
		handler: handler,
		ctx:     ctx,
	}
}

func (t *TCPServer) Start() {
	t.server = &http.Server{
		Addr:         setting.App.BindAddress,
		Handler:      *t.handler,
		ReadTimeout:  setting.App.Server.ReadTimeout,
		WriteTimeout: setting.App.Server.WriteTimeout,
	}

	log.Printf("Starting TCP server listening on %s", setting.App.BindAddress)
	go func() {
		if err := t.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	log.Printf("Stopping TCP server...")
}

func (t *TCPServer) Stop() {
	if err := t.server.Shutdown(t.ctx); err != nil {
		log.Printf("TCP server forced to shutdown: %s", err)
	}
}
