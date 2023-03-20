package server

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/dcarrillo/whatismyip/models"
	"golang.org/x/net/context"
)

type Server interface {
	Start()
	Stop()
}

type Factory struct {
	tcpServer  *TCPServer
	tlsServer  *TLSServer
	quicServer *QuicServer
}

func Setup(ctx context.Context, handler http.Handler) *Factory {
	var tcpServer *TCPServer
	var tlsServer *TLSServer
	var quicServer *QuicServer

	if setting.App.BindAddress != "" {
		tcpServer = NewTCPServer(ctx, &handler)
	}

	if setting.App.TLSAddress != "" {
		tlsServer = NewTLSServer(ctx, &handler)
		if setting.App.EnableHTTP3 {
			quicServer = NewQuicServer(ctx, tlsServer)
		}
	}

	return &Factory{
		tcpServer:  tcpServer,
		tlsServer:  tlsServer,
		quicServer: quicServer,
	}
}

func (w *Factory) Run() {
	w.start()
	log.Printf("Starting server handler...")
	w.Watcher()
}

func (w *Factory) Watcher() {
	signalChan := make(chan os.Signal, 3)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	var s os.Signal

	for {
		s = <-signalChan

		if s == syscall.SIGHUP {
			w.stop()
			models.CloseDBs()
			models.Setup(setting.App.GeodbPath.City, setting.App.GeodbPath.ASN)
			w.start()
		} else {
			log.Printf("Shutting down...")
			w.stop()
			models.CloseDBs()
			break
		}
	}
}

func (w *Factory) start() {
	if w.tcpServer != nil {
		w.tcpServer.Start()
	}

	if w.tlsServer != nil {
		w.tlsServer.Start()
		if w.quicServer != nil {
			w.quicServer.Start()
		}
	}
}

func (w *Factory) stop() {
	if w.tcpServer != nil {
		w.tcpServer.Stop()
	}

	if w.tlsServer != nil {
		if w.quicServer != nil {
			w.quicServer.Stop()
		}
		w.tlsServer.Stop()
	}
}
