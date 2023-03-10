package models

import (
	"log"
	"net"

	"github.com/oschwald/maxminddb-golang"
)

// GeoRecord is the model for City database
type GeoRecord struct {
	Country struct {
		ISOCode string            `maxminddb:"iso_code"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Location struct {
		Latitude  float64 `maxminddb:"latitude"`
		Longitude float64 `maxminddb:"longitude"`
		TimeZone  string  `maxminddb:"time_zone"`
	} `maxminddb:"location"`
	Postal struct {
		Code string `maxminddb:"code"`
	} `maxminddb:"postal"`
}

// ASNRecord is the model for ASN database
type ASNRecord struct {
	AutonomousSystemNumber       uint   `maxminddb:"autonomous_system_number"`
	AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
}

type geodb struct {
	city *maxminddb.Reader
	asn  *maxminddb.Reader
}

var db geodb

func openMMDB(path string) *maxminddb.Reader {
	db, err := maxminddb.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Database %s has been loaded\n", path)

	return db
}

// Setup opens all Geolite2 databases
func Setup(cityPath string, asnPath string) {
	db.city = openMMDB(cityPath)
	db.asn = openMMDB(asnPath)
}

// CloseDBs unmaps from memory and frees resources to the filesystem
func CloseDBs() {
	log.Printf("Closing dbs...")
	if err := db.city.Close(); err != nil {
		log.Printf("Error closing city db: %s", err)
	}
	if err := db.asn.Close(); err != nil {
		log.Printf("Error closing ASN db: %s", err)
	}
}

// LookUp an IP and get city data
func (record *GeoRecord) LookUp(ip net.IP) error {
	return db.city.Lookup(ip, record)
}

// LookUp an IP and get ASN data
func (record *ASNRecord) LookUp(ip net.IP) error {
	return db.asn.Lookup(ip, record)
}
