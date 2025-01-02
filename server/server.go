package server

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dcarrillo/whatismyip/service"
)

type Server interface {
	Start()
	Stop()
}

type Manager struct {
	servers []Server
	geoSvc  *service.Geo
}

func Setup(servers []Server, geoSvc *service.Geo) *Manager {
	return &Manager{
		servers: servers,
		geoSvc:  geoSvc,
	}
}

func (m *Manager) Run() {
	m.start()

	signalChan := make(chan os.Signal, len(m.servers))
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	var s os.Signal
	for {
		s = <-signalChan

		if s == syscall.SIGHUP {
			m.stop()
			if m.geoSvc != nil {
				m.geoSvc.Reload()
			}
			m.start()
		} else {
			log.Print("Shutting down...")
			if m.geoSvc != nil {
				m.geoSvc.Shutdown()
			}
			m.stop()
			break
		}
	}
}

func (m *Manager) start() {
	for _, s := range m.servers {
		s.Start()
	}
}

func (m *Manager) stop() {
	for _, s := range m.servers {
		s.Stop()
	}
}
