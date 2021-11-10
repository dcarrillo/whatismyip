package router

import (
	"os"
	"testing"

	"github.com/dcarrillo/whatismyip/models"
	"github.com/gin-gonic/gin"
)

type testIPs struct {
	ipv4    string
	ipv4ASN string
	ipv6    string
	ipv6ASN string
}

type contentTypes struct {
	text string
	json string
}

var (
	app    *gin.Engine
	testIP = testIPs{
		ipv4:    "81.2.69.192",
		ipv4ASN: "82.99.17.64",
		ipv6:    "2a02:9000::1",
		ipv6ASN: "2a02:a800::1",
	}
	contentType = contentTypes{
		text: "text/plain; charset=utf-8",
		json: "application/json; charset=utf-8",
	}
)

const trustedHeader = "X-Real-IP"

func TestMain(m *testing.M) {
	app = gin.Default()
	app.TrustedPlatform = trustedHeader
	models.Setup("../test/GeoIP2-City-Test.mmdb", "../test/GeoLite2-ASN-Test.mmdb")
	Setup(app)
	defer models.CloseDBs()

	os.Exit(m.Run())
}
