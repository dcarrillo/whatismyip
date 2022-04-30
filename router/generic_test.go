package router

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootContentType(t *testing.T) {
	tests := []struct {
		name     string
		accepted string
		expected string
	}{
		{
			name:     "Accept wildcard",
			accepted: "*/*",
			expected: contentType.text,
		},
		{
			name:     "Bogus accept",
			accepted: "bogus/plain",
			expected: contentType.text,
		},
		{
			name:     "Accept plain text",
			accepted: "text/plain",
			expected: contentType.text,
		},
		{
			name:     "Accept json",
			accepted: "application/json",
			expected: contentType.json,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set(trustedHeader, testIP.ipv4)
			req.Header.Set("Accept", tt.accepted)

			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			assert.Equal(t, tt.expected, w.Header().Get("Content-Type"))
		})
	}
}

func TestGetIP(t *testing.T) {
	expected := testIP.ipv4 + "\n"
	tests := []struct {
		name     string
		accepted string
	}{
		{
			name:     "No browser",
			accepted: "*/*",
		},
		{
			name:     "Bogus accept",
			accepted: "bogus/plain",
		},
		{
			name:     "Plain accept",
			accepted: "text/plain",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set(trustedHeader, testIP.ipv4)
			req.Header.Set("Accept", tt.accepted)

			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			assert.Equal(t, expected, w.Body.String())
			assert.Equal(t, contentType.text, w.Header().Get("Content-Type"))
		})
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
	req.RemoteAddr = net.JoinHostPort(testIP.ipv4, "1000")
	req.Header.Set(trustedHeader, testIP.ipv4)

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
	type args struct {
		ip string
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "IPv4",
			args: args{
				ip: testIP.ipv4,
			},
			expected: jsonIPv4,
		},
		{
			name: "IPv6",
			args: args{
				ip: testIP.ipv6,
			},
			expected: jsonIPv6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/json", nil)
			req.RemoteAddr = net.JoinHostPort(tt.args.ip, "1000")
			req.Host = "test"
			req.Header.Set(trustedHeader, tt.args.ip)

			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			assert.Equal(t, contentType.json, w.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.expected, w.Body.String())
		})
	}
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
	req.RemoteAddr = net.JoinHostPort(testIP.ipv4, "1000")
	req.Host = "test"
	req.Header.Set(trustedHeader, testIP.ipv4)
	req.Header.Set("Header1", "one")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, contentType.text, w.Header().Get("Content-Type"))
	assert.Equal(t, expected, w.Body.String())
}
