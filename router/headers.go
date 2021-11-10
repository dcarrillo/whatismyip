package router

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/dcarrillo/whatismyip/internal/httputils"
	"github.com/gin-gonic/gin"
)

func getHeadersAsSortedString(ctx *gin.Context) {
	h := ctx.Request.Header
	h["Host"] = []string{ctx.Request.Host}

	ctx.String(http.StatusOK, httputils.HeadersToSortedString(h))
}

func getHeaderAsString(ctx *gin.Context) {
	h := ctx.Params.ByName("header")
	if v := ctx.GetHeader(h); v != "" {
		ctx.String(http.StatusOK, template.HTMLEscapeString(v))
	} else if strings.ToLower(h) == "host" {
		ctx.String(http.StatusOK, template.HTMLEscapeString(ctx.Request.Host))
	} else {
		ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}
}
