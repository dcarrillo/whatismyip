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
	dirname := filepath.Dir(filename)

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
		},
		ExposedPorts: []string{"8000:8000", "8001:8001"},
		WaitingFor:   wait.ForLog("Starting TLS server listening on :8001"),
		BindMounts: map[string]string{
			filepath.Join(dirname, "/../test/GeoIP2-City-Test.mmdb"):  "/tmp/GeoIP2-City-Test.mmdb",
			filepath.Join(dirname, "/../test/GeoLite2-ASN-Test.mmdb"): "/tmp/GeoLite2-ASN-Test.mmdb",
			filepath.Join(dirname, "/../test/server.pem"):             "/tmp/server.pem",
			filepath.Join(dirname, "/../test/server.key"):             "/tmp/server.key",
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
	for _, url := range []string{"http://localhost:8000/json", "https://localhost:8001/json"} {
		resp, _ := http.Get(url)
		assert.Equal(t, 200, resp.StatusCode)

		var dat router.JSONResponse
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		assert.NoError(t, json.Unmarshal(body, &dat))
	}
}
