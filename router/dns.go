package router

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	validator "github.com/dcarrillo/whatismyip/internal/validator/uuid"
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
func GetDNSDiscoveryHandler(store *cache.Cache, domain string, redirectPort string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !strings.HasSuffix(ctx.Request.Host, domain) {
			ctx.Next()
			return
		}

		if ctx.Request.Host == domain && ctx.Request.URL.Path == "/" {
			ctx.Redirect(http.StatusFound, fmt.Sprintf("http://%s.%s%s", uuid.New().String(), domain, redirectPort))
			ctx.Abort()
			return
		}

		handleDNS(ctx, store)
		ctx.Abort()
	}
}

func handleDNS(ctx *gin.Context, store *cache.Cache) {
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
		cityRecord := geoSvc.LookUpCity(ip)
		asnRecord := geoSvc.LookUpASN(ip)

		geoResp = dnsGeoData{
			Country:         cityRecord.Country.Names["en"],
			AsnOrganization: asnRecord.AutonomousSystemOrganization,
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
