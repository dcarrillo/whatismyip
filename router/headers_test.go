package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "/user-agent", nil)
	req.Header.Set("User-Agent", "test")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType.text, w.Header().Get("Content-Type"))
	assert.Equal(t, "test", w.Body.String())
}

func TestHeaders(t *testing.T) {
	expected := `Header1: value1
Header2: value21
Header2: value22
Header3: value3
Host: 
`
	_, _ = setting.Setup([]string{
		"-geoip2-city", "city",
		"-geoip2-asn", "asn",
		"-trusted-header", trustedHeader,
		"-trusted-port-header", trustedPortHeader,
	})
	req, _ := http.NewRequest("GET", "/headers", nil)
	req.Header = map[string][]string{
		"Header1": {"value1"},
		"Header2": {"value21", "value22"},
		"Header3": {"value3"},
	}
	req.Header.Set(trustedHeader, "1.1.1.1")
	req.Header.Set(trustedPortHeader, "1025")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType.text, w.Header().Get("Content-Type"))
	assert.Equal(t, expected, w.Body.String())
}
