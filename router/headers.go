package router

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/dcarrillo/whatismyip/internal/httputils"
	"github.com/gin-gonic/gin"
)

func (rt *Router) getHeadersAsSortedString(ctx *gin.Context) {
	h := httputils.GetHeadersWithoutTrustedHeaders(ctx, rt.trustedHeader, rt.trustedPortHeader)
	h.Set("Host", ctx.Request.Host)
	ctx.String(http.StatusOK, httputils.HeadersToSortedString(h))
}

func (rt *Router) getHeaderAsString(ctx *gin.Context) {
	headers := httputils.GetHeadersWithoutTrustedHeaders(ctx, rt.trustedHeader, rt.trustedPortHeader)

	h := ctx.Params.ByName("header")
	if v := headers.Get(ctx.Params.ByName("header")); v != "" {
		ctx.String(http.StatusOK, template.HTMLEscapeString(v))
	} else if strings.ToLower(h) == "host" {
		ctx.String(http.StatusOK, template.HTMLEscapeString(ctx.Request.Host))
	} else {
		ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}
}
