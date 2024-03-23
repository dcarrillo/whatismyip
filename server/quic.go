package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/quic-go/quic-go/http3"
)

type Quic struct {
	server    *http3.Server
	tlsServer *TLS
	ctx       context.Context
}

func NewQuicServer(ctx context.Context, tlsServer *TLS) *Quic {
	return &Quic{
		tlsServer: tlsServer,
		ctx:       ctx,
	}
}

func (q *Quic) Start() {
	q.server = &http3.Server{
		Addr:    setting.App.TLSAddress,
		Handler: q.tlsServer.server.Handler,
	}

	parentHandler := q.tlsServer.server.Handler
	q.tlsServer.server.Handler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if err := q.server.SetQuicHeaders(rw.Header()); err != nil {
			log.Fatal(err)
		}

		parentHandler.ServeHTTP(rw, req)
	})

	log.Printf("Starting QUIC server listening on %s (udp)", setting.App.TLSAddress)
	go func() {
		if err := q.server.ListenAndServeTLS(setting.App.TLSCrtPath, setting.App.TLSKeyPath); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
}

func (q *Quic) Stop() {
	log.Printf("Stopping QUIC server...")
	if err := q.server.Close(); err != nil {
		log.Printf("QUIC server forced to shutdown")
	}
}
