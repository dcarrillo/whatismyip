package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/quic-go/quic-go/http3"
)

type Quic struct {
	server    *http3.Server
	tlsServer *TLS
	ctx       context.Context
	addr      string
	crtPath   string
	keyPath   string
}

func NewQuicServer(ctx context.Context, tlsServer *TLS, addr, crt, key string) *Quic {
	return &Quic{
		tlsServer: tlsServer,
		ctx:       ctx,
		addr:      addr,
		crtPath:   crt,
		keyPath:   key,
	}
}

func (q *Quic) Start() {
	q.server = &http3.Server{
		Addr:    q.addr,
		Handler: q.tlsServer.server.Handler,
	}

	parentHandler := q.tlsServer.server.Handler
	q.tlsServer.server.Handler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if err := q.server.SetQUICHeaders(rw.Header()); err != nil {
			log.Printf("Failed to set QUIC headers: %s", err)
		}

		parentHandler.ServeHTTP(rw, req)
	})

	log.Printf("Starting QUIC server listening on %s (udp)", q.addr)
	go func() {
		if err := q.server.ListenAndServeTLS(q.crtPath, q.keyPath); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
}

func (q *Quic) Stop() {
	log.Print("Stopping QUIC server...")
	ctx, cancel := context.WithTimeout(q.ctx, shutdownTimeout)
	defer cancel()
	if err := q.server.Shutdown(ctx); err != nil {
		log.Printf("QUIC server forced to shutdown: %s", err)
	}
}
