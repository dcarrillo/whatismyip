package router

import (
	"net"
	"net/http"
	"path/filepath"

	"github.com/dcarrillo/whatismyip/internal/httputils"
	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/dcarrillo/whatismyip/service"
	"github.com/gin-gonic/gin"
)

// JSONResponse maps data as json
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
	switch ctx.NegotiateFormat(gin.MIMEPlain, gin.MIMEHTML, gin.MIMEJSON) {
	case gin.MIMEHTML:
		name := "home"
		if setting.App.TemplatePath != "" {
			name = filepath.Base(setting.App.TemplatePath)
		}
		ctx.HTML(http.StatusOK, name, jsonOutput(ctx))
	case gin.MIMEJSON:
		getJSON(ctx)
	default:
		ctx.String(http.StatusOK, ctx.ClientIP()+"\n")
	}
}

func getClientPort(ctx *gin.Context) string {
	var port string
	if setting.App.TrustedPortHeader == "" {
		if setting.App.TrustedHeader != "" {
			port = "unknown"
		} else {
			_, port, _ = net.SplitHostPort(ctx.Request.RemoteAddr)
		}
	} else {
		port = ctx.GetHeader(setting.App.TrustedPortHeader)
		if port == "" {
			port = "unknown"
		}
	}

	return port
}

func getClientPortAsString(ctx *gin.Context) {
	ctx.String(http.StatusOK, getClientPort(ctx)+"\n")
}

func getAllAsString(ctx *gin.Context) {
	output := "IP: " + ctx.ClientIP() + "\n"
	output += "Client Port: " + getClientPort(ctx) + "\n"

	r := service.Geo{IP: net.ParseIP(ctx.ClientIP())}
	if record := r.LookUpCity(); record != nil {
		output += geoCityRecordToString(record) + "\n"
	}

	if record := r.LookUpASN(); record != nil {
		output += geoASNRecordToString(record) + "\n"
	}

	h := httputils.GetHeadersWithoutTrustedHeaders(ctx)
	h.Set("Host", ctx.Request.Host)
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

	return JSONResponse{
		IP:              ctx.ClientIP(),
		IPVersion:       version,
		ClientPort:      getClientPort(ctx),
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
		Headers:         httputils.GetHeadersWithoutTrustedHeaders(ctx),
	}
}
