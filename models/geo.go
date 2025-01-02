package models

import (
	"fmt"
	"log"
	"net"

	"github.com/oschwald/maxminddb-golang"
)

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

type GeoDB struct {
	cityPath string
	asnPath  string
	City     *maxminddb.Reader
	ASN      *maxminddb.Reader
}

func Setup(cityPath string, asnPath string) (*GeoDB, error) {
	city, asn, err := openDatabases(cityPath, asnPath)
	if err != nil {
		return nil, err
	}

	return &GeoDB{
		cityPath: cityPath,
		asnPath:  asnPath,
		City:     city,
		ASN:      asn,
	}, nil
}

func (db *GeoDB) CloseDBs() error {
	var errs []error

	if db.City != nil {
		if err := db.City.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing city db: %w", err))
		}
	}

	if db.ASN != nil {
		if err := db.ASN.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing ASN db: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing databases: %s", errs)
	}
	return nil
}

func (db *GeoDB) Reload() error {
	if err := db.CloseDBs(); err != nil {
		return fmt.Errorf("closing existing connections: %w", err)
	}

	city, asn, err := openDatabases(db.cityPath, db.asnPath)
	if err != nil {
		return fmt.Errorf("opening new connections: %w", err)
	}

	db.City = city
	db.ASN = asn
	return nil
}

func (db *GeoDB) LookupCity(ip net.IP) (*GeoRecord, error) {
	record := &GeoRecord{}
	err := db.City.Lookup(ip, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (db *GeoDB) LookupASN(ip net.IP) (*ASNRecord, error) {
	record := &ASNRecord{}
	err := db.ASN.Lookup(ip, record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func openDatabases(cityPath, asnPath string) (*maxminddb.Reader, *maxminddb.Reader, error) {
	city, err := openMMDB(cityPath)
	if err != nil {
		return nil, nil, err
	}

	asn, err := openMMDB(asnPath)
	if err != nil {
		return nil, nil, err
	}

	return city, asn, nil
}

func openMMDB(path string) (*maxminddb.Reader, error) {
	db, err := maxminddb.Open(path)
	if err != nil {
		return nil, err
	}
	log.Printf("Database %s has been loaded\n", path)

	return db, nil
}
