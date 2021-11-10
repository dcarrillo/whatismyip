package service

import (
	"log"
	"net"

	"github.com/dcarrillo/whatismyip/models"
)

type Geo struct {
	IP net.IP
}

func (g *Geo) LookUpCity() *models.GeoRecord {
	record := &models.GeoRecord{}
	err := record.LookUp(g.IP)
	if err != nil {
		log.Println(err)
		return nil
	}

	return record
}

func (g *Geo) LookUpASN() *models.ASNRecord {
	record := &models.ASNRecord{}
	err := record.LookUp(g.IP)
	if err != nil {
		log.Println(err)
		return nil
	}

	return record
}
