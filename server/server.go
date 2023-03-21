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

func (f *Factory) Run() {
	f.start()

	signalChan := make(chan os.Signal, 3)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	var s os.Signal
	for {
		s = <-signalChan

		if s == syscall.SIGHUP {
			f.stop()
			models.CloseDBs()
			models.Setup(setting.App.GeodbPath.City, setting.App.GeodbPath.ASN)
			f.start()
		} else {
			log.Printf("Shutting down...")
			f.stop()
			models.CloseDBs()
			break
		}
	}
}

func (f *Factory) start() {
	if f.tcpServer != nil {
		f.tcpServer.Start()
	}

	if f.tlsServer != nil {
		f.tlsServer.Start()
		if f.quicServer != nil {
			f.quicServer.Start()
		}
	}
}

func (f *Factory) stop() {
	if f.tcpServer != nil {
		f.tcpServer.Stop()
	}

	if f.tlsServer != nil {
		if f.quicServer != nil {
			f.quicServer.Stop()
		}
		f.tlsServer.Stop()
	}
}
