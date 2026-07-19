package service

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/dcarrillo/whatismyip/internal/metrics"
	"github.com/dcarrillo/whatismyip/models"
)

type Geo struct {
	ctx    context.Context
	cancel context.CancelFunc
	db     *models.GeoDB
	mu     sync.RWMutex
}

func NewGeo(ctx context.Context, cityPath string, asnPath string) (*Geo, error) {
	ctx, cancel := context.WithCancel(ctx)

	db, err := models.Setup(cityPath, asnPath)
	if err != nil {
		cancel()
		return nil, err
	}

	geo := &Geo{
		ctx:    ctx,
		cancel: cancel,
		db:     db,
	}

	return geo, nil
}

func (g *Geo) LookUpCity(ip net.IP) *models.GeoRecord {
	g.mu.RLock()
	defer g.mu.RUnlock()

	record, err := g.db.LookupCity(ip)
	if err != nil {
		log.Print(err)
		return nil
	}

	metrics.RecordGeoLookup("city")
	return record
}

func (g *Geo) LookUpASN(ip net.IP) *models.ASNRecord {
	g.mu.RLock()
	defer g.mu.RUnlock()

	record, err := g.db.LookupASN(ip)
	if err != nil {
		log.Print(err)
		return nil
	}

	metrics.RecordGeoLookup("asn")
	return record
}

func (g *Geo) Shutdown() {
	g.cancel()

	g.mu.Lock()
	defer g.mu.Unlock()

	if err := g.db.CloseDBs(); err != nil {
		log.Printf("Error closing geo databases: %s", err)
	}
}

func (g *Geo) Reload() {
	if err := g.ctx.Err(); err != nil {
		log.Printf("Skipping reload, service is shutting down: %v", err)
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if err := g.db.Reload(); err != nil {
		log.Printf("Geo database reload failed: %s", err)
		return
	}
	log.Print("Geo database reloaded")
}
