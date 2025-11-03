package integrationtests

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	validator "github.com/dcarrillo/whatismyip/internal/validator/uuid"
	"github.com/dcarrillo/whatismyip/router"
	"github.com/docker/docker/api/types"
	"github.com/quic-go/quic-go/http3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func customDialContext() func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &net.Dialer{
			Resolver: &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
					d := net.Dialer{}
					return d.DialContext(ctx, "udp", "127.0.0.1:53531")
				},
			},
		}

		return dialer.DialContext(ctx, network, addr)
	}
}

func testWhatIsMyDNS(t *testing.T) {
	t.Run("RequestDNSDiscovery", func(t *testing.T) {
		http.DefaultTransport.(*http.Transport).DialContext = customDialContext()
		req, err := http.NewRequest("GET", "http://localhost:8000", nil)
		assert.NoError(t, err)
		req.Host = "dns.example.com"
		client := &http.Client{
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		u, err := resp.Location()
		assert.NoError(t, err)
		assert.True(t, validator.IsValid(strings.Split(u.Hostname(), ".")[0]))

		for _, accept := range []string{"application/json", "*/*", "text/html"} {
			req, err = http.NewRequest("GET", u.String(), nil)
			req.Host = u.Hostname()
			req.Header.Set("Accept", accept)
			assert.NoError(t, err)
			resp, err = client.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			if accept == "application/json" {
				assert.NoError(t, json.Unmarshal(body, &router.DNSJSONResponse{}))
			} else {
				ip := strings.Split(string(body), " ")[0]
				assert.True(t, net.ParseIP(ip) != nil)
			}
		}
	})
}

func TestContainerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skiping integration tests")
	}

	ctx := context.Background()
	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			FromDockerfile: tc.FromDockerfile{
				Context:       "../",
				Dockerfile:    "./test/Dockerfile",
				PrintBuildLog: true,
				KeepImage:     false,
				BuildOptionsModifier: func(buildOptions *types.ImageBuildOptions) {
					buildOptions.Target = "test"
				},
			},
			ExposedPorts: []string{
				"8000:8000",
				"8001:8001",
				"8001:8001/udp",
				"9100:9100",
				"53531:53/udp",
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
				"-metrics-bind", ":9100",
				"-resolver", "/resolver.yml",
			},
			Files: []tc.ContainerFile{
				{
					HostFilePath:      "./../test/GeoIP2-City-Test.mmdb",
					ContainerFilePath: "/GeoIP2-City-Test.mmdb",
				},
				{
					HostFilePath:      "./../test/GeoLite2-ASN-Test.mmdb",
					ContainerFilePath: "/GeoLite2-ASN-Test.mmdb",
				},
				{
					HostFilePath:      "./../test/server.pem",
					ContainerFilePath: "/server.pem",
				},
				{
					HostFilePath:      "./../test/server.key",
					ContainerFilePath: "/server.key",
				},
				{
					HostFilePath:      "./../test/resolver.yml",
					ContainerFilePath: "/resolver.yml",
				},
			},
			WaitingFor: wait.ForLog("Starting QUIC server"),
			AutoRemove: true,
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { c.Terminate(ctx) })

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
				resp, err = client.Do(req)
				assert.NoError(t, err)
				body, err = io.ReadAll(resp.Body)
				assert.NoError(t, err)
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

	t.Run("RequestMetricsEndpoint", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://localhost:9100/metrics", nil)
		assert.NoError(t, err)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		bodyStr := string(body)

		assert.Contains(t, bodyStr, "whatismyip_http_requests_total")
		assert.Contains(t, bodyStr, "whatismyip_http_request_duration_seconds")
		assert.Contains(t, bodyStr, "# HELP")
		assert.Contains(t, bodyStr, "# TYPE")
	})

	testWhatIsMyDNS(t)
}

// TODO If other flags like this one are implemented we should think of a better approach
func TestContainerIntegrationDisableScan(t *testing.T) {
	if testing.Short() {
		t.Skip("Skiping integration tests")
	}

	ctx := context.Background()
	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			FromDockerfile: tc.FromDockerfile{
				Context:       "../",
				Dockerfile:    "./test/Dockerfile",
				PrintBuildLog: true,
				KeepImage:     false,
				BuildOptionsModifier: func(buildOptions *types.ImageBuildOptions) {
					buildOptions.Target = "test"
				},
			},
			ExposedPorts: []string{
				"8000:8000",
			},
			Cmd: []string{
				"-geoip2-city", "/GeoIP2-City-Test.mmdb",
				"-geoip2-asn", "/GeoLite2-ASN-Test.mmdb",
				"-bind", ":8000",
				"-trusted-header", "X-Real-IP",
				"-enable-secure-headers",
				"-disable-scan",
			},
			Files: []tc.ContainerFile{
				{
					HostFilePath:      "./../test/GeoIP2-City-Test.mmdb",
					ContainerFilePath: "/GeoIP2-City-Test.mmdb",
				},
				{
					HostFilePath:      "./../test/GeoLite2-ASN-Test.mmdb",
					ContainerFilePath: "/GeoLite2-ASN-Test.mmdb",
				},
			},
			WaitingFor: wait.ForLog("Starting TCP server"),
			AutoRemove: true,
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { c.Terminate(ctx) })

	t.Run("RequestScanEndpointWithDisabledScan", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://localhost:8000/scan/tcp/8000", nil)
		assert.NoError(t, err)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("X-Real-IP", "127.0.0.1")

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NotEmpty(t, body)
	})
}

func doQuicRequest(req *http.Request) (*http.Response, []byte, error) {
	roundTripper := &http3.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
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
