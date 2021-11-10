package models

import (
	"log"
	"net"

	"github.com/oschwald/maxminddb-golang"
)

type Record interface {
	LookUp(ip net.IP)
}

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

func Setup(cityPath string, asnPath string) {
	db.city = openMMDB(cityPath)
	db.asn = openMMDB(asnPath)
}

func CloseDBs() {
	log.Printf("Closing dbs...")
	if err := db.city.Close(); err != nil {
		log.Printf("Error closing city db: %s", err)
	}
	if err := db.asn.Close(); err != nil {
		log.Printf("Error closing ASN db: %s", err)
	}
}

func (record *GeoRecord) LookUp(ip net.IP) error {
	err := db.city.Lookup(ip, record)
	if err != nil {
		return err
	}

	return nil
}

func (record *ASNRecord) LookUp(ip net.IP) error {
	err := db.asn.Lookup(ip, record)
	if err != nil {
		return err
	}

	return nil
}
