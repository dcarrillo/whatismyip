package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct {
	server   *http.Server
	ctx      context.Context
	addr     string
	timeouts Timeouts
}

func NewPrometheusServer(ctx context.Context, addr string, timeouts Timeouts) *Prometheus {
	return &Prometheus{
		ctx:      ctx,
		addr:     addr,
		timeouts: timeouts,
	}
}

func (p *Prometheus) Start() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	p.server = &http.Server{
		Addr:         p.addr,
		Handler:      mux,
		ReadTimeout:  p.timeouts.ReadTimeout,
		WriteTimeout: p.timeouts.WriteTimeout,
	}

	log.Printf("Starting Prometheus server listening on %s", p.addr)
	go func() {
		if err := p.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
}

func (p *Prometheus) Stop() {
	log.Print("Stopping Prometheus server...")
	ctx, cancel := context.WithTimeout(p.ctx, shutdownTimeout)
	defer cancel()
	if err := p.server.Shutdown(ctx); err != nil {
		log.Printf("Prometheus server forced to shutdown: %s", err)
	}
}
