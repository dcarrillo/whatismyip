package setting

import (
	"flag"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseMandatoryFlags(t *testing.T) {
	var mandatoryFlags = []struct {
		args []string
		conf settings
	}{
		{
			[]string{},
			settings{},
		},
		{
			[]string{"-geoip2-city", "/city-path"},
			settings{},
		},
		{
			[]string{"-geoip2-asn", "/asn-path"},
			settings{},
		},
		{
			[]string{
				"-geoip2-city", "/city-path", "-geoip2-asn", "/asn-path", "-tls-bind", ":9000",
			},
			settings{},
		},
		{
			[]string{
				"-geoip2-city", "/city-path", "-geoip2-asn", "/asn-path", "-tls-bind", ":9000",
				"-tls-crt", "/crt-path",
			},
			settings{},
		},
		{
			[]string{
				"-geoip2-city", "/city-path", "-geoip2-asn", "/asn-path", "-tls-bind", ":9000",
				"-tls-key", "/key-path",
			},
			settings{},
		},
	}

	for _, tt := range mandatoryFlags {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			_, err := Setup(tt.args)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "mandatory")
		})
	}
}

func TestParseFlags(t *testing.T) {
	var flags = []struct {
		args []string
		conf settings
	}{
		{
			[]string{"-geoip2-city", "/city-path", "-geoip2-asn", "/asn-path"},
			settings{
				GeodbPath: geodbPath{
					City: "/city-path",
					ASN:  "/asn-path",
				},
				TemplatePath:  "",
				BindAddress:   ":8080",
				TLSAddress:    "",
				TLSCrtPath:    "",
				TLSKeyPath:    "",
				TrustedHeader: "",
				Server: serverSettings{
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 10 * time.Second,
				},
			},
		},
		{
			[]string{"-bind", ":8001", "-geoip2-city", "/city-path", "-geoip2-asn", "/asn-path"},
			settings{
				GeodbPath: geodbPath{
					City: "/city-path",
					ASN:  "/asn-path",
				},
				TemplatePath:  "",
				BindAddress:   ":8001",
				TLSAddress:    "",
				TLSCrtPath:    "",
				TLSKeyPath:    "",
				TrustedHeader: "",
				Server: serverSettings{
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 10 * time.Second,
				},
			},
		},
		{
			[]string{
				"-geoip2-city", "/city-path", "-geoip2-asn", "/asn-path", "-tls-bind", ":9000",
				"-tls-crt", "/crt-path", "-tls-key", "/key-path",
			},
			settings{
				GeodbPath: geodbPath{
					City: "/city-path",
					ASN:  "/asn-path",
				},
				TemplatePath:  "",
				BindAddress:   ":8080",
				TLSAddress:    ":9000",
				TLSCrtPath:    "/crt-path",
				TLSKeyPath:    "/key-path",
				TrustedHeader: "",
				Server: serverSettings{
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 10 * time.Second,
				},
			},
		},
		{
			[]string{
				"-geoip2-city", "/city-path", "-geoip2-asn", "/asn-path",
				"-trusted-header", "header",
			},
			settings{
				GeodbPath: geodbPath{
					City: "/city-path",
					ASN:  "/asn-path",
				},
				TemplatePath:  "",
				BindAddress:   ":8080",
				TLSAddress:    "",
				TLSCrtPath:    "",
				TLSKeyPath:    "",
				TrustedHeader: "header",
				Server: serverSettings{
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 10 * time.Second,
				},
			},
		},
	}

	for _, tt := range flags {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			_, err := Setup(tt.args)
			assert.Nil(t, err)
			assert.True(t, reflect.DeepEqual(App, tt.conf))
		})
	}
}

func TestParseFlagsUsage(t *testing.T) {
	var usageArgs = []string{"-help", "-h", "--help"}

	for _, arg := range usageArgs {
		t.Run(arg, func(t *testing.T) {
			output, err := Setup([]string{arg})
			assert.ErrorIs(t, err, flag.ErrHelp)
			assert.Contains(t, output, "Usage of")
		})
	}
}

func TestParseFlagVersion(t *testing.T) {
	output, err := Setup([]string{"-version"})
	assert.ErrorIs(t, err, ErrVersion)
	assert.Contains(t, output, "whatismyip version")
}

func TestParseFlagTemplate(t *testing.T) {
	flags := []string{
		"-geoip2-city", "/city-path", "-geoip2-asn", "/asn-path",
		"-template", "/template-path",
	}
	_, err := Setup(flags)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")

	flags = []string{
		"-geoip2-city", "/city-path", "-geoip2-asn", "/asn-path",
		"-template", "/",
	}
	_, err = Setup(flags)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a file")
}
