package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeo(t *testing.T) {
	expected := `City: London
Country: United Kingdom
Country Code: GB
Latitude: 51.514200
Longitude: -0.093100
Postal Code: 
Time Zone: Europe/London
`

	req, _ := http.NewRequest("GET", "/geo", nil)
	req.Header.Set("X-Real-IP", testIP.ipv4)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType.text, w.Header().Get("Content-Type"))
	assert.Equal(t, expected, w.Body.String())
}

func TestGeoField(t *testing.T) {
	req, _ := http.NewRequest("GET", "/geo/latitude", nil)
	req.Header.Set("X-Real-IP", testIP.ipv4)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType.text, w.Header().Get("Content-Type"))
	assert.Equal(t, "51.514200", w.Body.String())
}

func TestGeoField404(t *testing.T) {
	req, _ := http.NewRequest("GET", "/geo/not-found", nil)
	req.Header.Set("X-Real-IP", testIP.ipv4)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestASN(t *testing.T) {
	expected := `ASN Number: 12552
ASN Organization: IP-Only
`

	req, _ := http.NewRequest("GET", "/asn", nil)
	req.Header.Set("X-Real-IP", testIP.ipv4ASN)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType.text, w.Header().Get("Content-Type"))
	assert.Equal(t, expected, w.Body.String())
}

func TestASNField(t *testing.T) {
	req, _ := http.NewRequest("GET", "/asn/organization", nil)
	req.Header.Set("X-Real-IP", testIP.ipv4ASN)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType.text, w.Header().Get("Content-Type"))
	assert.Equal(t, "IP-Only", w.Body.String())
}

func TestASNField404(t *testing.T) {
	req, _ := http.NewRequest("GET", "/asn/not-found", nil)
	req.Header.Set("X-Real-IP", testIP.ipv4ASN)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestASN_IPv6(t *testing.T) {
	expected := `ASN Number: 6739
ASN Organization: Cableuropa - ONO
`

	req, _ := http.NewRequest("GET", "/asn", nil)
	req.Header.Set("X-Real-IP", testIP.ipv6ASN)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType.text, w.Header().Get("Content-Type"))
	assert.Equal(t, expected, w.Body.String())
}
