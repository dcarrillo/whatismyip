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
	html string
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
		html: "content-type: text/html; charset=utf-8",
		text: "text/plain; charset=utf-8",
		json: "application/json; charset=utf-8",
	}
	jsonIPv4     = `{"client_port":"1001","ip":"81.2.69.192","ip_version":4,"country":"United Kingdom","country_code":"GB","city":"London","latitude":51.5142,"longitude":-0.0931,"postal_code":"","time_zone":"Europe/London","asn":0,"asn_organization":"","host":"test", "headers": {}}`
	jsonIPv6     = `{"asn":3352, "asn_organization":"TELEFONICA DE ESPANA", "city":"", "client_port":"1001", "country":"", "country_code":"", "host":"test", "ip":"2a02:9000::1", "ip_version":6, "latitude":0, "longitude":0, "postal_code":"", "time_zone":"", "headers": {}}`
	jsonDNSIPv4  = `{"dns":{"ip":"81.2.69.192","country":"United Kingdom","provider":""}}`
	plainDNSIPv4 = "81.2.69.192 (United Kingdom / )\n"
)

const (
	trustedHeader     = "X-Real-IP"
	trustedPortHeader = "X-Real-Port"
	domain            = "dns.example.com"
)

func TestMain(m *testing.M) {
	app = gin.Default()
	app.TrustedPlatform = trustedHeader
	models.Setup("../test/GeoIP2-City-Test.mmdb", "../test/GeoLite2-ASN-Test.mmdb")
	Setup(app)
	defer models.CloseDBs()

	os.Exit(m.Run())
}
