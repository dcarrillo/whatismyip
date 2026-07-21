package router

import (
	"net"
	"net/http"
	"path/filepath"

	"github.com/dcarrillo/whatismyip/internal/httputils"
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

func (rt *Router) getRoot(ctx *gin.Context) {
	switch ctx.NegotiateFormat(gin.MIMEPlain, gin.MIMEHTML, gin.MIMEJSON) {
	case gin.MIMEHTML:
		name := "home"
		if rt.templatePath != "" {
			name = filepath.Base(rt.templatePath)
		}
		ctx.HTML(http.StatusOK, name, rt.jsonOutput(ctx))
	case gin.MIMEJSON:
		rt.getJSON(ctx)
	default:
		ctx.String(http.StatusOK, ctx.ClientIP()+"\n")
	}
}

func (rt *Router) getClientPort(ctx *gin.Context) string {
	var port string
	if rt.trustedPortHeader == "" {
		if rt.trustedHeader != "" {
			port = "unknown"
		} else {
			_, port, _ = net.SplitHostPort(ctx.Request.RemoteAddr)
		}
	} else {
		port = ctx.GetHeader(rt.trustedPortHeader)
		if port == "" {
			port = "unknown"
		}
	}

	return port
}

func (rt *Router) getClientPortAsString(ctx *gin.Context) {
	ctx.String(http.StatusOK, rt.getClientPort(ctx)+"\n")
}

func (rt *Router) getAllAsString(ctx *gin.Context) {
	ip := net.ParseIP(ctx.ClientIP())

	output := "IP: " + ip.String() + "\n"
	output += "Client Port: " + rt.getClientPort(ctx) + "\n"

	if rt.geo != nil {
		if cityRecord := rt.geo.LookUpCity(ip); cityRecord != nil {
			output += geoCityRecordToString(cityRecord) + "\n"
		}
		if asnRecord := rt.geo.LookUpASN(ip); asnRecord != nil {
			output += geoASNRecordToString(asnRecord) + "\n"
		}
	}

	h := httputils.GetHeadersWithoutTrustedHeaders(ctx, rt.trustedHeader, rt.trustedPortHeader)
	h.Set("Host", ctx.Request.Host)
	output += httputils.HeadersToSortedString(h)

	ctx.String(http.StatusOK, output)
}

func (rt *Router) getJSON(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, rt.jsonOutput(ctx))
}

func (rt *Router) jsonOutput(ctx *gin.Context) JSONResponse {
	ip := net.ParseIP(ctx.ClientIP())

	var version byte = 4
	if p := ip.To4(); p == nil {
		version = 6
	}

	geoResp := GeoResponse{}
	if rt.geo != nil {
		if cityRecord := rt.geo.LookUpCity(ip); cityRecord != nil {
			geoResp.Country = cityRecord.Country.Names["en"]
			geoResp.CountryCode = cityRecord.Country.ISOCode
			geoResp.City = cityRecord.City.Names["en"]
			geoResp.Latitude = cityRecord.Location.Latitude
			geoResp.Longitude = cityRecord.Location.Longitude
			geoResp.PostalCode = cityRecord.Postal.Code
			geoResp.TimeZone = cityRecord.Location.TimeZone
		}
		if asnRecord := rt.geo.LookUpASN(ip); asnRecord != nil {
			geoResp.ASN = asnRecord.AutonomousSystemNumber
			geoResp.ASNOrganization = asnRecord.AutonomousSystemOrganization
		}
	}

	return JSONResponse{
		IP:          ip.String(),
		IPVersion:   version,
		ClientPort:  rt.getClientPort(ctx),
		Host:        ctx.Request.Host,
		Headers:     httputils.GetHeadersWithoutTrustedHeaders(ctx, rt.trustedHeader, rt.trustedPortHeader),
		GeoResponse: geoResp,
	}
}
