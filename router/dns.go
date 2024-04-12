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
type dnsData struct {
	IP              string `json:"ip"`
	Country         string `json:"country"`
	AsnOrganization string `json:"provider"`
}

// TODO
// Implement a proper vhost manager instead of using a middleware
func GetDNSDiscoveryHandler(store *cache.Cache, domain string, redirectPort string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !strings.HasSuffix(ctx.Request.Host, domain) {
			ctx.Next()
			return
		}

		if ctx.Request.Host == domain {
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

	geo := service.Geo{IP: ip}
	j := DNSJSONResponse{
		DNS: dnsData{
			IP:              ipStr,
			Country:         geo.LookUpCity().Country.Names["en"],
			AsnOrganization: geo.LookUpASN().AutonomousSystemOrganization,
		},
	}

	switch ctx.NegotiateFormat(gin.MIMEPlain, gin.MIMEHTML, gin.MIMEJSON) {
	case gin.MIMEJSON:
		ctx.JSON(http.StatusOK, j)
	default:
		ctx.String(http.StatusOK, fmt.Sprintf("%s (%s / %s)\n", j.DNS.IP, j.DNS.Country, j.DNS.AsnOrganization))
	}
}
