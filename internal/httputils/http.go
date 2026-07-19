package httputils

import (
	"fmt"
	"net/http"
	"net/textproto"
	"sort"
	"strings"

	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/gin-gonic/gin"
)

// HeadersToSortedString sorts and dumps http.Header to a string separated by \n
func HeadersToSortedString(headers http.Header) string {
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var output strings.Builder
	for _, k := range keys {
		for _, h := range headers[k] {
			output.WriteString(k + ": " + h + "\n")
		}
	}

	return output.String()
}

// GetHeadersWithoutTrustedHeaders returns a copy of the request headers with the trusted headers removed
func GetHeadersWithoutTrustedHeaders(ctx *gin.Context) http.Header {
	h := ctx.Request.Header.Clone()

	for _, k := range []string{setting.App.TrustedHeader, setting.App.TrustedPortHeader} {
		delete(h, textproto.CanonicalMIMEHeaderKey(k))
	}

	return h
}

// GetLogFormatter returns our custom log format
func GetLogFormatter(param gin.LogFormatterParams) string {
	return fmt.Sprintf("%s - [%s] \"%s %s %s\" %d %d %d %s \"%s\" \"%s\" \"%s\"\n",
		param.ClientIP,
		param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
		param.Method,
		param.Path,
		param.Request.Proto,
		param.StatusCode,
		param.BodySize,
		param.Latency.Nanoseconds(),
		normalizeLog(param.Request.Referer()),
		normalizeLog(param.Request.UserAgent()),
		normalizeLog(param.Request.Header["X-Forwarded-For"]),
		normalizeLog(param.ErrorMessage),
	)
}

func normalizeLog(log any) any {
	switch v := log.(type) {
	case string:
		if v == "" {
			return "-"
		}
	case []string:
		if len(v) == 0 {
			return "-"
		}
		return strings.Join(v, ", ")
	}

	return log
}
