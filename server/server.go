package server

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/dcarrillo/whatismyip/models"
)

type Server interface {
	Start()
	Stop()
}

type Manager struct {
	servers []Server
}

func Setup(servers []Server) *Manager {
	return &Manager{
		servers: servers,
	}
}

func (m *Manager) Run() {
	m.start()

	models.Setup(setting.App.GeodbPath.City, setting.App.GeodbPath.ASN)
	signalChan := make(chan os.Signal, len(m.servers))
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	var s os.Signal
	for {
		s = <-signalChan

		if s == syscall.SIGHUP {
			m.stop()
			models.CloseDBs()
			models.Setup(setting.App.GeodbPath.City, setting.App.GeodbPath.ASN)
			m.start()
		} else {
			log.Printf("Shutting down...")
			m.stop()
			models.CloseDBs()
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
