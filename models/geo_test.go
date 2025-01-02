package models

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModels(t *testing.T) {
	expectedCity := &GeoRecord{
		Country: struct {
			ISOCode string            "maxminddb:\"iso_code\""
			Names   map[string]string "maxminddb:\"names\""
		}{
			ISOCode: "GB",
			Names: map[string]string{
				"de":    "Vereinigtes Königreich",
				"en":    "United Kingdom",
				"es":    "Reino Unido",
				"fr":    "Royaume-Uni",
				"ja":    "イギリス",
				"pt-BR": "Reino Unido",
				"ru":    "Великобритания",
				"zh-CN": "英国",
			},
		},
		City: struct {
			Names map[string]string "maxminddb:\"names\""
		}{
			Names: map[string]string{
				"de":    "London",
				"en":    "London",
				"es":    "Londres",
				"fr":    "Londres",
				"ja":    "ロンドン",
				"pt-BR": "Londres",
				"ru":    "Лондон",
			},
		},
		Location: struct {
			Latitude  float64 "maxminddb:\"latitude\""
			Longitude float64 "maxminddb:\"longitude\""
			TimeZone  string  "maxminddb:\"time_zone\""
		}{
			Latitude:  51.5142,
			Longitude: -0.0931,
			TimeZone:  "Europe/London",
		},
		Postal: struct {
			Code string "maxminddb:\"code\""
		}{
			Code: "",
		},
	}

	expectedASN := &ASNRecord{
		AutonomousSystemNumber:       12552,
		AutonomousSystemOrganization: "IP-Only",
	}

	db, err := Setup("../test/GeoIP2-City-Test.mmdb", "../test/GeoLite2-ASN-Test.mmdb")
	require.NoError(t, err, fmt.Sprintf("Error setting up db: %s", err))
	defer db.CloseDBs()
	assert.NotNil(t, db.ASN)
	assert.NotNil(t, db.City)

	cityRecord, err := db.LookupCity(net.ParseIP("81.2.69.192"))
	require.NoError(t, err, fmt.Sprintf("Error looking up city: %s", err))
	assert.Equal(t, expectedCity, cityRecord)
	_, err = db.LookupCity(net.ParseIP("error"))
	assert.Error(t, err)

	asnRecord, err := db.LookupASN(net.ParseIP("82.99.17.64"))
	require.NoError(t, err, fmt.Sprintf("Error looking up asn: %s", err))
	assert.Equal(t, expectedASN, asnRecord)
	_, err = db.LookupASN(net.ParseIP("error"))
	assert.Error(t, err)
}
