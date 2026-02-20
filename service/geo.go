package service

import (
	"context"
	"net"
	"sync"

	"github.com/dcarrillo/whatismyip/internal/logger"
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
	record, err := g.db.LookupCity(ip)
	if err != nil {
		logger.Error("Failed to lookup city", "error", err)
		return nil
	}

	metrics.RecordGeoLookup("city")
	return record
}

func (g *Geo) LookUpASN(ip net.IP) *models.ASNRecord {
	record, err := g.db.LookupASN(ip)
	if err != nil {
		logger.Error("Failed to lookup ASN", "error", err)
		return nil
	}

	metrics.RecordGeoLookup("asn")
	return record
}

func (g *Geo) Shutdown() {
	g.cancel()
	g.db.CloseDBs()
}

func (g *Geo) Reload() {
	if err := g.ctx.Err(); err != nil {
		logger.Warn("Skipping reload, service is shutting down", "error", err)
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.db.Reload()
	logger.Info("Geo database reloaded")
}
