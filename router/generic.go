package router

import (
	"net"
	"net/http"
	"path/filepath"

	"github.com/dcarrillo/whatismyip/internal/httputils"
	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/gin-gonic/gin"
)

type GeoResponse struct {
	Country         string  `json:"country,omitempty"`
	CountryCode     string  `json:"country_code,omitempty"`
	City            string  `json:"city,omitempty"`
	Latitude        float64 `json:"latitude,omitempty"`
	Longitude       float64 `json:"longitude,omitempty"`
	PostalCode      string  `json:"postal_code,omitempty"`
	TimeZone        string  `json:"time_zone,omitempty"`
	ASN             uint    `json:"asn,omitempty"`
	ASNOrganization string  `json:"asn_organization,omitempty"`
}

type JSONResponse struct {
	IP         string      `json:"ip"`
	IPVersion  byte        `json:"ip_version"`
	ClientPort string      `json:"client_port"`
	Host       string      `json:"host"`
	Headers    http.Header `json:"headers"`
	GeoResponse
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
	ip := net.ParseIP(ctx.ClientIP())

	output := "IP: " + ip.String() + "\n"
	output += "Client Port: " + getClientPort(ctx) + "\n"

	if geoSvc != nil {
		output += geoCityRecordToString(geoSvc.LookUpCity(ip)) + "\n"
		output += geoASNRecordToString(geoSvc.LookUpASN(ip)) + "\n"
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
	ip := net.ParseIP(ctx.ClientIP())

	var version byte = 4
	if p := ip.To4(); p == nil {
		version = 6
	}

	geoResp := GeoResponse{}
	if geoSvc != nil {
		cityRecord := geoSvc.LookUpCity(ip)
		asnRecord := geoSvc.LookUpASN(ip)

		geoResp = GeoResponse{
			Country:         cityRecord.Country.Names["en"],
			CountryCode:     cityRecord.Country.ISOCode,
			City:            cityRecord.City.Names["en"],
			Latitude:        cityRecord.Location.Latitude,
			Longitude:       cityRecord.Location.Longitude,
			PostalCode:      cityRecord.Postal.Code,
			TimeZone:        cityRecord.Location.TimeZone,
			ASN:             asnRecord.AutonomousSystemNumber,
			ASNOrganization: asnRecord.AutonomousSystemOrganization,
		}
	}

	return JSONResponse{
		IP:          ip.String(),
		IPVersion:   version,
		ClientPort:  getClientPort(ctx),
		Host:        ctx.Request.Host,
		Headers:     httputils.GetHeadersWithoutTrustedHeaders(ctx),
		GeoResponse: geoResp,
	}
}
