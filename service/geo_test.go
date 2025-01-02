package service

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var geoSvc *Geo

func TestMain(m *testing.M) {
	geoSvc, _ = NewGeo(context.Background(), "../test/GeoIP2-City-Test.mmdb", "../test/GeoLite2-ASN-Test.mmdb")
	os.Exit(m.Run())
}

func TestCityLookup(t *testing.T) {
	c := geoSvc.LookUpCity(net.ParseIP("error"))
	assert.Nil(t, c)

	c = geoSvc.LookUpCity(net.ParseIP("1.1.1.1"))
	assert.NotNil(t, c)
}

func TestASNLookup(t *testing.T) {
	a := geoSvc.LookUpASN(net.ParseIP("error"))
	assert.Nil(t, a)

	a = geoSvc.LookUpASN(net.ParseIP("1.1.1.1"))
	assert.NotNil(t, a)
}
