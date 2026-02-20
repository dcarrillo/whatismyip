package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/dcarrillo/whatismyip/internal/logger"
	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct {
	server *http.Server
	ctx    context.Context
}

func NewPrometheusServer(ctx context.Context) *Prometheus {
	return &Prometheus{
		ctx: ctx,
	}
}

func (p *Prometheus) Start() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	p.server = &http.Server{
		Addr:         setting.App.PrometheusAddress,
		Handler:      mux,
		ReadTimeout:  setting.App.Server.ReadTimeout,
		WriteTimeout: setting.App.Server.WriteTimeout,
	}

	logger.Info("Starting Prometheus server", "address", setting.App.PrometheusAddress)
	go func() {
		if err := p.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Prometheus server error", "error", err)
		}
	}()
}

func (p *Prometheus) Stop() {
	logger.Info("Stopping Prometheus server")
	if err := p.server.Shutdown(p.ctx); err != nil {
		logger.Error("Prometheus server forced to shutdown", "error", err)
	}
}
