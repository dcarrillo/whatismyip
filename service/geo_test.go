package service

import (
	"net"
	"os"
	"testing"

	"github.com/dcarrillo/whatismyip/models"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	models.Setup("../test/GeoIP2-City-Test.mmdb", "../test/GeoLite2-ASN-Test.mmdb")
	defer models.CloseDBs()
	os.Exit(m.Run())
}

func TestCityLookup(t *testing.T) {
	ip := Geo{IP: net.ParseIP("error")}
	c := ip.LookUpCity()
	assert.Nil(t, c)

	ip = Geo{IP: net.ParseIP("1.1.1.1")}
	c = ip.LookUpCity()
	assert.NotNil(t, c)
}

func TestASNLookup(t *testing.T) {
	ip := Geo{IP: net.ParseIP("error")}
	a := ip.LookUpASN()
	assert.Nil(t, a)

	ip = Geo{IP: net.ParseIP("1.1.1.1")}
	a = ip.LookUpASN()
	assert.NotNil(t, a)
}
