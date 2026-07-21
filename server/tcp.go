package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"
)

type ServerTimeouts struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type TCP struct {
	server   *http.Server
	handler  http.Handler
	ctx      context.Context
	addr     string
	timeouts ServerTimeouts
}

func NewTCPServer(ctx context.Context, handler http.Handler, addr string, timeouts ServerTimeouts) *TCP {
	return &TCP{
		handler:  handler,
		ctx:      ctx,
		addr:     addr,
		timeouts: timeouts,
	}
}

func (t *TCP) Start() {
	t.server = &http.Server{
		Addr:         t.addr,
		Handler:      t.handler,
		ReadTimeout:  t.timeouts.ReadTimeout,
		WriteTimeout: t.timeouts.WriteTimeout,
	}

	log.Printf("Starting TCP server listening on %s", t.addr)
	go func() {
		if err := t.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
}

func (t *TCP) Stop() {
	log.Print("Stopping TCP server...")
	ctx, cancel := context.WithTimeout(t.ctx, shutdownTimeout)
	defer cancel()
	if err := t.server.Shutdown(ctx); err != nil {
		log.Printf("TCP server forced to shutdown: %s", err)
	}
}
