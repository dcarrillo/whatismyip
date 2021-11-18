package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIP4RootFromCli(t *testing.T) {
	uas := []string{
		"",
		"curl",
		"wget",
		"libwww-perl",
		"python",
		"ansible-httpget",
		"HTTPie",
		"WindowsPowerShell",
		"http_request",
		"Go-http-client",
	}

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", testIP.ipv4)

	for _, ua := range uas {
		req.Header.Set("User-Agent", ua)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, testIP.ipv4, w.Body.String())
	}
}

func TestHost(t *testing.T) {
	req, _ := http.NewRequest("GET", "/host", nil)
	req.Host = "test"
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "test", w.Body.String())
}

func TestClientPort(t *testing.T) {
	req, _ := http.NewRequest("GET", "/client-port", nil)
	req.RemoteAddr = testIP.ipv4 + ":" + "1000"
	req.Header.Set("X-Real-IP", testIP.ipv4)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType.text, w.Header().Get("Content-Type"))
	assert.Equal(t, "1000\n", w.Body.String())
}

func TestNotFound(t *testing.T) {
	req, _ := http.NewRequest("GET", "/not-found", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
	assert.Equal(t, contentType.text, w.Header().Get("Content-Type"))
	assert.Equal(t, "Not Found", w.Body.String())
}

func TestJSON(t *testing.T) {
	expectedIPv4 := `{"client_port":"1000","ip":"81.2.69.192","ip_version":4,"country":"United Kingdom","country_code":"GB","city":"London","latitude":51.5142,"longitude":-0.0931,"postal_code":"","time_zone":"Europe/London","asn":0,"asn_organization":"","host":"test","headers":{"X-Real-Ip":["81.2.69.192"]}}`
	expectedIPv6 := `{"asn":3352, "asn_organization":"TELEFONICA DE ESPANA", "city":"", "client_port":"9000", "country":"", "country_code":"", "headers":{"X-Real-Ip":["2a02:9000::1"]}, "host":"test", "ip":"2a02:9000::1", "ip_version":6, "latitude":0, "longitude":0, "postal_code":"", "time_zone":""}`

	req, _ := http.NewRequest("GET", "/json", nil)
	req.RemoteAddr = testIP.ipv4 + ":" + "1000"
	req.Host = "test"
	req.Header.Set("X-Real-IP", testIP.ipv4)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType.json, w.Header().Get("Content-Type"))
	assert.JSONEq(t, expectedIPv4, w.Body.String())

	req.RemoteAddr = testIP.ipv6 + ":" + "1000"
	req.Host = "test"
	req.Header.Set("X-Real-IP", testIP.ipv6)

	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType.json, w.Header().Get("Content-Type"))
	assert.JSONEq(t, expectedIPv6, w.Body.String())
}

func TestAll(t *testing.T) {
	expected := `IP: 81.2.69.192
Client Port: 1000
City: London
Country: United Kingdom
Country Code: GB
Latitude: 51.514200
Longitude: -0.093100
Postal Code: 
Time Zone: Europe/London

ASN Number: 0
ASN Organization: 

Header1: one
Host: test
X-Real-Ip: 81.2.69.192
`

	req, _ := http.NewRequest("GET", "/all", nil)
	req.RemoteAddr = testIP.ipv4 + ":" + "1000"
	req.Host = "test"
	req.Header.Set("X-Real-IP", testIP.ipv4)
	req.Header.Set("Header1", "one")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType.text, w.Header().Get("Content-Type"))
	assert.Equal(t, expected, w.Body.String())
}
