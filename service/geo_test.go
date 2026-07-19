package service

import (
	"context"
	"net"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestReloadIsSafeForConcurrentLookups(t *testing.T) {
	svc, err := NewGeo(context.Background(), "../test/GeoIP2-City-Test.mmdb", "../test/GeoLite2-ASN-Test.mmdb")
	require.NoError(t, err)
	defer svc.Shutdown()

	ip := net.ParseIP("81.2.69.192")

	var wg sync.WaitGroup
	stop := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				svc.LookUpCity(ip)
				svc.LookUpASN(ip)
			}
		}
	}()

	for range 100 {
		svc.Reload()
	}
	close(stop)
	wg.Wait()

	assert.NotNil(t, svc.LookUpCity(ip))
}
