package router

import (
	"net"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/dcarrillo/whatismyip/internal/httputils"
	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/dcarrillo/whatismyip/service"
	"github.com/gin-gonic/gin"
)

const userAgentPattern = `curl|wget|libwww-perl|python|ansible-httpget|HTTPie|WindowsPowerShell|http_request|Go-http-client|^$`

type JSONResponse struct {
	IP              string      `json:"ip"`
	IPVersion       byte        `json:"ip_version"`
	ClientPort      string      `json:"client_port"`
	Country         string      `json:"country"`
	CountryCode     string      `json:"country_code"`
	City            string      `json:"city"`
	Latitude        float64     `json:"latitude"`
	Longitude       float64     `json:"longitude"`
	PostalCode      string      `json:"postal_code"`
	TimeZone        string      `json:"time_zone"`
	ASN             uint        `json:"asn"`
	ASNOrganization string      `json:"asn_organization"`
	Host            string      `json:"host"`
	Headers         http.Header `json:"headers"`
}

func getRoot(ctx *gin.Context) {
	reg := regexp.MustCompile(userAgentPattern)
	if reg.Match([]byte(ctx.Request.UserAgent())) {
		ctx.String(http.StatusOK, ctx.ClientIP())
	} else {
		name := "home"
		if setting.App.TemplatePath != "" {
			name = filepath.Base(setting.App.TemplatePath)
		}
		ctx.HTML(http.StatusOK, name, jsonOutput(ctx))
	}
}

func getClientPortAsString(ctx *gin.Context) {
	_, port, _ := net.SplitHostPort(ctx.Request.RemoteAddr)
	ctx.String(http.StatusOK, port+"\n")
}

func getAllAsString(ctx *gin.Context) {
	output := "IP: " + ctx.ClientIP() + "\n"
	_, port, _ := net.SplitHostPort(ctx.Request.RemoteAddr)
	output += "Client Port: " + port + "\n"

	r := service.Geo{IP: net.ParseIP(ctx.ClientIP())}
	if record := r.LookUpCity(); record != nil {
		output += geoCityRecordToString(record) + "\n"
	}

	if record := r.LookUpASN(); record != nil {
		output += geoASNRecordToString(record) + "\n"
	}

	h := ctx.Request.Header
	h["Host"] = []string{ctx.Request.Host}
	output += httputils.HeadersToSortedString(h)

	ctx.String(http.StatusOK, output)
}

func getJSON(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, jsonOutput(ctx))
}

func jsonOutput(ctx *gin.Context) JSONResponse {
	ip := service.Geo{IP: net.ParseIP(ctx.ClientIP())}
	asnRecord := ip.LookUpASN()
	cityRecord := ip.LookUpCity()
	var version byte = 4
	if p := net.ParseIP(ctx.ClientIP()).To4(); p == nil {
		version = 6
	}

	_, port, _ := net.SplitHostPort(ctx.Request.RemoteAddr)
	return JSONResponse{
		IP:              ctx.ClientIP(),
		IPVersion:       version,
		ClientPort:      port,
		Country:         cityRecord.Country.Names["en"],
		CountryCode:     cityRecord.Country.ISOCode,
		City:            cityRecord.City.Names["en"],
		Latitude:        cityRecord.Location.Latitude,
		Longitude:       cityRecord.Location.Longitude,
		PostalCode:      cityRecord.Postal.Code,
		TimeZone:        cityRecord.Location.TimeZone,
		ASN:             asnRecord.AutonomousSystemNumber,
		ASNOrganization: asnRecord.AutonomousSystemOrganization,
		Host:            ctx.Request.Host,
		Headers:         ctx.Request.Header,
	}
}
