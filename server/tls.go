package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/dcarrillo/whatismyip/internal/logger"
	"github.com/dcarrillo/whatismyip/internal/setting"
)

type TLS struct {
	server  *http.Server
	handler *http.Handler
	ctx     context.Context
}

func NewTLSServer(ctx context.Context, handler *http.Handler) *TLS {
	return &TLS{
		handler: handler,
		ctx:     ctx,
	}
}

func (t *TLS) Start() {
	t.server = &http.Server{
		Addr:         setting.App.TLSAddress,
		Handler:      *t.handler,
		ReadTimeout:  setting.App.Server.ReadTimeout,
		WriteTimeout: setting.App.Server.WriteTimeout,
	}

	logger.Info("Starting TLS server", "address", setting.App.TLSAddress)
	go func() {
		if err := t.server.ListenAndServeTLS(setting.App.TLSCrtPath, setting.App.TLSKeyPath); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Error("TLS server error", "error", err)
		}
	}()
}

func (t *TLS) Stop() {
	logger.Info("Stopping TLS server")
	if err := t.server.Shutdown(t.ctx); err != nil {
		logger.Error("TLS server forced to shutdown", "error", err)
	}
}
