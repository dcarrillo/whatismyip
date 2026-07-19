package models

import (
	"errors"
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
	return closeReaders(db.City, db.ASN)
}

func (db *GeoDB) Reload() error {
	city, asn, err := openDatabases(db.cityPath, db.asnPath)
	if err != nil {
		return fmt.Errorf("opening new databases: %w", err)
	}

	oldCity, oldASN := db.City, db.ASN
	db.City, db.ASN = city, asn

	if err := closeReaders(oldCity, oldASN); err != nil {
		return fmt.Errorf("closing previous databases: %w", err)
	}
	return nil
}

func closeReaders(city *maxminddb.Reader, asn *maxminddb.Reader) error {
	var errs []error

	if city != nil {
		if err := city.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing city db: %w", err))
		}
	}

	if asn != nil {
		if err := asn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing ASN db: %w", err))
		}
	}

	return errors.Join(errs...)
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
