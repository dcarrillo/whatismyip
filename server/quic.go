package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/dcarrillo/whatismyip/internal/logger"
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
		if err := q.server.SetQUICHeaders(rw.Header()); err != nil {
			logger.Error("Failed to set QUIC headers", "error", err)
		}

		parentHandler.ServeHTTP(rw, req)
	})

	logger.Info("Starting QUIC server", "address", setting.App.TLSAddress, "protocol", "udp")
	go func() {
		if err := q.server.ListenAndServeTLS(setting.App.TLSCrtPath, setting.App.TLSKeyPath); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Error("QUIC server error", "error", err)
		}
	}()
}

func (q *Quic) Stop() {
	logger.Info("Stopping QUIC server")
	if err := q.server.Close(); err != nil {
		logger.Error("QUIC server forced to shutdown", "error", err)
	}
}
