package server

import (
	"context"
	"log"
	"net/http"

	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/quic-go/quic-go/http3"
)

type QuicServer struct {
	server    *http3.Server
	tlsServer *TLSServer
	ctx       context.Context
}

func NewQuicServer(ctx context.Context, tlsServer *TLSServer) QuicServer {
	return QuicServer{
		tlsServer: tlsServer,
		ctx:       ctx,
	}
}

func (q *QuicServer) Start() {
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
			err.Error() != "quic: Server closed" {
			log.Fatal(err)
		}
	}()
	log.Printf("Stopping QUIC server...")
}

func (q *QuicServer) Stop() {
	if err := q.server.Close(); err != nil {
		log.Printf("QUIC server forced to shutdown")
	}
}
