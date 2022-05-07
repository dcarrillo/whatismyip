package integrationtests

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dcarrillo/whatismyip/router"
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
		},
		ExposedPorts: []string{"8000:8000", "8001:8001"},
		WaitingFor:   wait.ForLog("Starting TLS server listening on :8001"),
		BindMounts: map[string]string{
			"/tmp/GeoIP2-City-Test.mmdb":  filepath.Join(dir, "/../test/GeoIP2-City-Test.mmdb"),
			"/tmp/GeoLite2-ASN-Test.mmdb": filepath.Join(dir, "/../test/GeoLite2-ASN-Test.mmdb"),
			"/tmp/server.pem":             filepath.Join(dir, "/../test/server.pem"),
			"/tmp/server.key":             filepath.Join(dir, "/../test/server.key"),
		},
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
		err := container.Terminate(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	for _, url := range []string{"http://localhost:8000", "https://localhost:8001"} {
		client := &http.Client{}
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Accept", "application/json")
		resp, _ := client.Do(req)
		assert.Equal(t, 200, resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		assert.NoError(t, json.Unmarshal(body, &router.JSONResponse{}))
		assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
		assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
		assert.Equal(t, "1; mode=block", resp.Header.Get("X-Xss-Protection"))
	}
}
