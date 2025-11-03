package server

import (
	"context"
	"errors"
	"log"
	"net/http"

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

	log.Printf("Starting Prometheus server listening on %s", setting.App.PrometheusAddress)
	go func() {
		if err := p.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
}

func (p *Prometheus) Stop() {
	log.Print("Stopping Prometheus server...")
	if err := p.server.Shutdown(p.ctx); err != nil {
		log.Printf("Prometheus server forced to shutdown: %s", err)
	}
}
