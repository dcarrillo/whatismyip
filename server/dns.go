package server

import (
	"context"
	"log"
	"strconv"

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

	log.Printf("Starting DNS server listening on :%d (udp)", port)
	go func() {
		if err := d.server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
}

func (d *DNS) Stop() {
	log.Print("Stopping DNS server...")
	if err := d.server.Shutdown(); err != nil {
		log.Printf("DNS server forced to shutdown: %s", err)
	}
}
