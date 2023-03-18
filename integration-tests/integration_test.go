package integrationtests

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dcarrillo/whatismyip/router"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func buildContainer() testcontainers.ContainerRequest {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../",
			Dockerfile: "Dockerfile",
		},
		Cmd: []string{
			"-geoip2-city", "/tmp/GeoIP2-City-Test.mmdb",
			"-geoip2-asn", "/tmp/GeoLite2-ASN-Test.mmdb",
			"-bind", ":8000",
			"-tls-bind", ":8001",
			"-tls-crt", "/tmp/server.pem",
			"-tls-key", "/tmp/server.key",
			"-trusted-header", "X-Real-IP",
			"-enable-secure-headers",
			"-enable-http3",
		},
		ExposedPorts: []string{"8000:8000", "8001:8001", "8001:8001/udp"},
		WaitingFor: wait.ForHTTP("/geo").
			WithTLS(true, &tls.Config{InsecureSkipVerify: true}).
			WithPort("8001"),
		Mounts: testcontainers.Mounts(
			testcontainers.BindMount(
				filepath.Join(dir, "/../test/GeoIP2-City-Test.mmdb"),
				"/tmp/GeoIP2-City-Test.mmdb",
			),
			testcontainers.BindMount(
				filepath.Join(dir, "/../test/GeoLite2-ASN-Test.mmdb"),
				"/tmp/GeoLite2-ASN-Test.mmdb",
			),
			testcontainers.BindMount(filepath.Join(dir, "/../test/server.pem"), "/tmp/server.pem"),
			testcontainers.BindMount(filepath.Join(dir, "/../test/server.key"), "/tmp/server.key"),
		),
	}

	return req
}

func TestContainerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skiping integration tests")
	}

	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: buildContainer(),
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = container.Terminate(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	tests := []struct {
		name string
		url  string
		quic bool
	}{
		{
			name: "RequestOverHTTP",
			url:  "http://localhost:8000",
			quic: false,
		},
		{
			name: "RequestOverHTTPs",
			url:  "https://localhost:8001",
			quic: false,
		},
		{
			name: "RequestOverUDPWithQuic",
			url:  "https://localhost:8001",
			quic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			req.Header.Set("Accept", "application/json")

			var resp *http.Response
			var body []byte
			if tt.quic {
				resp, body, err = doQuicRequest(req)
			} else {
				client := &http.Client{}
				resp, _ = client.Do(req)
				body, err = io.ReadAll(resp.Body)
				if strings.Contains(tt.url, "https://") {
					assert.Equal(t, `h3=":8001"; ma=2592000,h3-29=":8001"; ma=2592000`, resp.Header.Get("Alt-Svc"))
				}
			}
			assert.NoError(t, err)

			assert.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)

			assert.NoError(t, json.Unmarshal(body, &router.JSONResponse{}))
			assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
			assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
			assert.Equal(t, "1; mode=block", resp.Header.Get("X-Xss-Protection"))
		})
	}
}

func doQuicRequest(req *http.Request) (*http.Response, []byte, error) {
	roundTripper := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		QuicConfig: &quic.Config{},
	}
	defer roundTripper.Close()

	client := &http.Client{
		Transport: roundTripper,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	return resp, body, nil
}
