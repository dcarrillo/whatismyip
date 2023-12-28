package integrationtests

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
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
			"-geoip2-city", "/GeoIP2-City-Test.mmdb",
			"-geoip2-asn", "/GeoLite2-ASN-Test.mmdb",
			"-bind", ":8000",
			"-tls-bind", ":8001",
			"-tls-crt", "/server.pem",
			"-tls-key", "/server.key",
			"-trusted-header", "X-Real-IP",
			"-enable-secure-headers",
			"-enable-http3",
		},
		ExposedPorts: []string{"8000:8000", "8001:8001", "8001:8001/udp"},
		WaitingFor: wait.ForHTTP("/geo").
			WithTLS(true, &tls.Config{InsecureSkipVerify: true}).
			WithPort("8001"),
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(dir, "/../test/GeoIP2-City-Test.mmdb"),
				ContainerFilePath: "/GeoIP2-City-Test.mmdb",
				FileMode:          0644,
			},
			{
				HostFilePath:      filepath.Join(dir, "/../test/GeoLite2-ASN-Test.mmdb"),
				ContainerFilePath: "/GeoLite2-ASN-Test.mmdb",
				FileMode:          0644,
			},
			{
				HostFilePath:      filepath.Join(dir, "/../test/server.pem"),
				ContainerFilePath: "/server.pem",
				FileMode:          0644,
			},
			{
				HostFilePath:      filepath.Join(dir, "/../test/server.key"),
				ContainerFilePath: "/server.key",
				FileMode:          0644,
			},
		},
	}

	return req
}

func initContainer(t assert.TestingT, request testcontainers.ContainerRequest) func() {
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})
	assert.NoError(t, err)

	return func() {
		assert.NoError(t, container.Terminate(ctx))
	}
}

func TestContainerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skiping integration tests")
	}

	t.Cleanup(initContainer(t, buildContainer()))

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

	testsPortScan := []struct {
		name string
		port int
		want bool
	}{
		{
			name: "RequestOpenPortScan",
			port: 8000,
			want: true,
		},
		{
			name: "RequestClosedPortScan",
			port: 65533,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			assert.NoError(t, err)
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
					assert.Equal(t, `h3=":8001"; ma=2592000`, resp.Header.Get("Alt-Svc"))
				}
			}
			assert.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)

			assert.NoError(t, json.Unmarshal(body, &router.JSONResponse{}))
			assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
			assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
			assert.Equal(t, "1; mode=block", resp.Header.Get("X-Xss-Protection"))
		})
	}

	for _, tt := range testsPortScan {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8000/scan/tcp/%d", tt.port), nil)
			assert.NoError(t, err)
			req.Header.Set("Accept", "application/json")
			req.Header.Set("X-Real-IP", "127.0.0.1")

			client := &http.Client{}
			resp, err := client.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			j := router.JSONScanResponse{}
			assert.NoError(t, json.Unmarshal(body, &j))
			assert.Equal(t, tt.want, j.Reachable)
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
