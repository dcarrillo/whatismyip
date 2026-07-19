package server

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dcarrillo/whatismyip/service"
)

const shutdownTimeout = 10 * time.Second

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

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		s := <-signalChan

		if s == syscall.SIGHUP {
			m.stop()
			if m.geoSvc != nil {
				m.geoSvc.Reload()
			}
			m.start()
		} else {
			log.Print("Shutting down...")
			m.stop()
			if m.geoSvc != nil {
				m.geoSvc.Shutdown()
			}
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
