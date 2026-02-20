package server

import (
	"context"
	"strconv"

	"github.com/dcarrillo/whatismyip/internal/logger"
	"github.com/miekg/dns"
)

const port = 53

type DNS struct {
	server  *dns.Server
	handler *dns.Handler
	ctx     context.Context
}

func NewDNSServer(ctx context.Context, handler dns.Handler) *DNS {
	return &DNS{
		handler: &handler,
		ctx:     ctx,
	}
}

func (d *DNS) Start() {
	d.server = &dns.Server{
		Addr:    ":" + strconv.Itoa(port),
		Net:     "udp",
		Handler: *d.handler,
		// UDPSize:   65535,
		// ReusePort: true,
	}

	logger.Info("Starting DNS server", "protocol", "udp", "port", port)
	go func() {
		if err := d.server.ListenAndServe(); err != nil {
			logger.Error("DNS server error", "error", err)
		}
	}()
}

func (d *DNS) Stop() {
	logger.Info("Stopping DNS server")
	if err := d.server.Shutdown(); err != nil {
		logger.Error("DNS server forced to shutdown", "error", err)
	}
}
