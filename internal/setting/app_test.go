package setting

import (
	"flag"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMandatoryFlags(t *testing.T) {
	mandatoryFlags := []struct {
		args []string
	}{
		{
			[]string{
				"-geoip2-city", "my-city-path",
			},
		},
		{
			[]string{
				"-geoip2-asn", "my-asn-path",
			},
		},

		{
			[]string{
				"-tls-bind", ":9000",
			},
		},
		{
			[]string{
				"-tls-bind", ":9000", "-tls-crt", "/crt-path",
			},
		},
		{
			[]string{
				"-tls-bind", ":9000", "-tls-key", "/key-path",
			},
		},
		{
			[]string{
				"-enable-http3",
			},
		},
		{
			[]string{
				"-bind", ":8000", "-trusted-port-header", "port-header",
			},
		},
	}

	for _, tt := range mandatoryFlags {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			_, err := Setup(tt.args)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "mandatory")
		})
	}
}

func TestParseFlags(t *testing.T) {
	flags := []struct {
		args []string
		conf settings
	}{
		{
			[]string{},
			settings{
				BindAddress: ":8080",
				Server: serverSettings{
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 10 * time.Second,
				},
			},
		},
		{
			[]string{"-disable-scan"},
			settings{
				BindAddress: ":8080",
				Server: serverSettings{
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 10 * time.Second,
				},
				DisableTCPScan: true,
			},
		},
		{
			[]string{"-bind", ":8001", "-geoip2-city", "/city-path", "-geoip2-asn", "/asn-path"},
			settings{
				GeodbPath: geodbConf{
					City: "/city-path",
					ASN:  "/asn-path",
				},
				BindAddress: ":8001",
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
				GeodbPath: geodbConf{
					City: "/city-path",
					ASN:  "/asn-path",
				},
				BindAddress: ":8080",
				TLSAddress:  ":9000",
				TLSCrtPath:  "/crt-path",
				TLSKeyPath:  "/key-path",
				Server: serverSettings{
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 10 * time.Second,
				},
			},
		},
		{
			[]string{
				"-geoip2-city", "/city-path", "-geoip2-asn", "/asn-path",
				"-trusted-header", "header", "-trusted-port-header", "port-header",
			},
			settings{
				GeodbPath: geodbConf{
					City: "/city-path",
					ASN:  "/asn-path",
				},
				BindAddress:       ":8080",
				TrustedHeader:     "header",
				TrustedPortHeader: "port-header",
				Server: serverSettings{
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 10 * time.Second,
				},
			},
		},
		{
			[]string{
				"-geoip2-city", "/city-path", "-geoip2-asn", "/asn-path",
				"-trusted-header", "header", "-enable-secure-headers",
			},
			settings{
				GeodbPath: geodbConf{
					City: "/city-path",
					ASN:  "/asn-path",
				},
				BindAddress:         ":8080",
				TrustedHeader:       "header",
				EnableSecureHeaders: true,
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
			require.Nil(t, err)
			assert.True(t, reflect.DeepEqual(App, tt.conf))
		})
	}
}

func TestParseFlagsUsage(t *testing.T) {
	usageArgs := []string{"-help", "-h", "--help"}

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
	testCases := []struct {
		name   string
		flags  []string
		errMsg string
	}{
		{
			name:   "Invalid template path",
			flags:  []string{"-template", "/template-path"},
			errMsg: "no such file or directory",
		},
		{
			name:   "Template path is a directory",
			flags:  []string{"-template", "/"},
			errMsg: "must be a file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Setup(tc.flags)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}
