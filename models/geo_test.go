package models

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
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

	Setup("../test/GeoIP2-City-Test.mmdb", "../test/GeoLite2-ASN-Test.mmdb")
	defer CloseDBs()

	assert.NotNil(t, db.asn)
	assert.NotNil(t, db.city)

	cityRecord := &GeoRecord{}
	assert.Nil(t, cityRecord.LookUp(net.ParseIP("81.2.69.192")))
	assert.Equal(t, expectedCity, cityRecord)
	assert.Error(t, cityRecord.LookUp(net.ParseIP("error")))

	asnRecord := &ASNRecord{}
	assert.Nil(t, asnRecord.LookUp(net.ParseIP("82.99.17.64")))
	assert.Equal(t, expectedASN, asnRecord)
	assert.Error(t, asnRecord.LookUp(net.ParseIP("error")))
}
