package service

import (
	"log"
	"net"

	"github.com/dcarrillo/whatismyip/models"
)

// Geo defines a base type for lookups
type Geo struct {
	IP net.IP
}

// LookUpCity queries the database for city data related to the given IP
func (g *Geo) LookUpCity() *models.GeoRecord {
	record := &models.GeoRecord{}
	err := record.LookUp(g.IP)
	if err != nil {
		log.Println(err)
		return nil
	}

	return record
}

// LookUpASN queries the database for ASN data related to the given IP
func (g *Geo) LookUpASN() *models.ASNRecord {
	record := &models.ASNRecord{}
	err := record.LookUp(g.IP)
	if err != nil {
		log.Println(err)
		return nil
	}

	return record
}
