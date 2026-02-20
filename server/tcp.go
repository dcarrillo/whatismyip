package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/dcarrillo/whatismyip/internal/logger"
	"github.com/dcarrillo/whatismyip/internal/setting"
)

type TCP struct {
	server  *http.Server
	handler *http.Handler
	ctx     context.Context
}

func NewTCPServer(ctx context.Context, handler *http.Handler) *TCP {
	return &TCP{
		handler: handler,
		ctx:     ctx,
	}
}

func (t *TCP) Start() {
	t.server = &http.Server{
		Addr:         setting.App.BindAddress,
		Handler:      *t.handler,
		ReadTimeout:  setting.App.Server.ReadTimeout,
		WriteTimeout: setting.App.Server.WriteTimeout,
	}

	logger.Info("Starting TCP server", "address", setting.App.BindAddress)
	go func() {
		if err := t.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("TCP server error", "error", err)
		}
	}()
}

func (t *TCP) Stop() {
	logger.Info("Stopping TCP server")
	if err := t.server.Shutdown(t.ctx); err != nil {
		logger.Error("TCP server forced to shutdown", "error", err)
	}
}
