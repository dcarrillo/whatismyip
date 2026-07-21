package router

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	validator "github.com/dcarrillo/whatismyip/internal/validator/uuid"
	"github.com/dcarrillo/whatismyip/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

type DNSJSONResponse struct {
	DNS dnsData `json:"dns"`
}
type dnsGeoData struct {
	Country         string `json:"country,omitempty"`
	AsnOrganization string `json:"provider,omitempty"`
}

type dnsData struct {
	IP string `json:"ip"`
	dnsGeoData
}

// TODO
// Implement a proper vhost manager instead of using a middleware
func GetDNSDiscoveryHandler(store *cache.Cache, geoSvc *service.Geo, domain string, redirectPort string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		host := hostWithoutPort(ctx.Request.Host)
		if host != domain && !strings.HasSuffix(host, "."+domain) {
			ctx.Next()
			return
		}

		if host == domain && ctx.Request.URL.Path == "/" {
			ctx.Redirect(http.StatusFound, fmt.Sprintf("http://%s.%s%s", uuid.New().String(), domain, redirectPort))
			ctx.Abort()
			return
		}

		handleDNS(ctx, store, geoSvc)
		ctx.Abort()
	}
}

func hostWithoutPort(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}

func handleDNS(ctx *gin.Context, store *cache.Cache, geoSvc *service.Geo) {
	d := strings.Split(ctx.Request.Host, ".")[0]
	if !validator.IsValid(d) {
		ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	v, found := store.Get(d)
	if !found {
		ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	ipStr, ok := v.(string)
	if !ok {
		ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	geoResp := dnsGeoData{}
	if geoSvc != nil {
		if cityRecord := geoSvc.LookUpCity(ip); cityRecord != nil {
			geoResp.Country = cityRecord.Country.Names["en"]
		}
		if asnRecord := geoSvc.LookUpASN(ip); asnRecord != nil {
			geoResp.AsnOrganization = asnRecord.AutonomousSystemOrganization
		}
	}

	j := DNSJSONResponse{
		DNS: dnsData{
			IP:         ipStr,
			dnsGeoData: geoResp,
		},
	}

	switch ctx.NegotiateFormat(gin.MIMEPlain, gin.MIMEHTML, gin.MIMEJSON) {
	case gin.MIMEJSON:
		ctx.JSON(http.StatusOK, j)
	default:
		ctx.String(http.StatusOK, fmt.Sprintf("%s (%s / %s)\n", j.DNS.IP, j.DNS.Country, j.DNS.AsnOrganization))
	}
}
