package setting

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dcarrillo/whatismyip/internal/core"
	"gopkg.in/yaml.v3"
)

type geodbConf struct {
	City string
	ASN  string
}
type serverSettings struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type resolver struct {
	Domain          string   `yaml:"domain"`
	ResourceRecords []string `yaml:"resource_records"`
	RedirectPort    string   `yaml:"redirect_port,omitempty"`
	Ipv4            []string `yaml:"ipv4,omitempty"`
	Ipv6            []string `yaml:"ipv6,omitempty"`
}

type Settings struct {
	GeodbPath           geodbConf
	TemplatePath        string
	BindAddress         string
	TLSAddress          string
	TLSCrtPath          string
	TLSKeyPath          string
	PrometheusAddress   string
	TrustedHeader       string
	TrustedPortHeader   string
	EnableSecureHeaders bool
	EnableHTTP3         bool
	DisableTCPScan      bool
	Server              serverSettings
	Resolver            resolver
	version             bool
}

const defaultAddress = ":8080"

var ErrVersion = errors.New("setting: version requested")

func Setup(args []string) (cfg Settings, output string, err error) {
	flags := flag.NewFlagSet("whatismyip", flag.ContinueOnError)
	var buf bytes.Buffer
	var resolverConf string
	flags.SetOutput(&buf)

	cfg.Server = serverSettings{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	flags.StringVar(&cfg.GeodbPath.City, "geoip2-city", "", "Path to GeoIP2 city database. Enables geo information (--geoip2-asn becomes mandatory)")
	flags.StringVar(&cfg.GeodbPath.ASN, "geoip2-asn", "", "Path to GeoIP2 ASN database. Enables ASN information. (--geoip2-city becomes mandatory)")
	flags.StringVar(&cfg.TemplatePath, "template", "", "Path to the template file")
	flags.StringVar(
		&resolverConf,
		"resolver",
		"",
		"Path to the resolver configuration. It actually enables the resolver for DNS client discovery.")
	flags.StringVar(
		&cfg.BindAddress,
		"bind",
		defaultAddress,
		"Listening address (see https://pkg.go.dev/net?#Listen)",
	)
	flags.StringVar(
		&cfg.TLSAddress,
		"tls-bind",
		"",
		"Listening address for TLS (see https://pkg.go.dev/net?#Listen)",
	)
	flags.StringVar(&cfg.TLSCrtPath, "tls-crt", "", "When using TLS, path to certificate file")
	flags.StringVar(&cfg.TLSKeyPath, "tls-key", "", "When using TLS, path to private key file")
	flags.StringVar(
		&cfg.PrometheusAddress,
		"metrics-bind",
		"",
		"Listening address for Prometheus metrics endpoint (see https://pkg.go.dev/net?#Listen). It enables the metrics available at the given address/port via the /metrics endpoint.",
	)
	flags.StringVar(
		&cfg.TrustedHeader,
		"trusted-header",
		"",
		"Trusted request header for remote IP (e.g. X-Real-IP). When using this feature if -trusted-port-header is not set the client port is shown as 'unknown'",
	)
	flags.StringVar(
		&cfg.TrustedPortHeader,
		"trusted-port-header",
		"",
		"Trusted request header for remote client port (e.g. X-Real-Port). When this parameter is set -trusted-header becomes mandatory",
	)
	flags.BoolVar(&cfg.version, "version", false, "Output version information and exit")
	flags.BoolVar(
		&cfg.EnableSecureHeaders,
		"enable-secure-headers",
		false,
		"Add sane security-related headers to every response",
	)
	flags.BoolVar(
		&cfg.EnableHTTP3,
		"enable-http3",
		false,
		"Enable HTTP/3 protocol. HTTP/3 requires --tls-bind set, as HTTP/3 starts as a TLS connection that then gets upgraded to UDP. The UDP port is the same as the one used for the TLS server.",
	)
	flags.BoolVar(
		&cfg.DisableTCPScan,
		"disable-scan",
		false,
		"Disable TCP port scanning functionality",
	)

	err = flags.Parse(args)
	if err != nil {
		return cfg, buf.String(), err
	}

	if cfg.version {
		return cfg, fmt.Sprintf("whatismyip version %s", core.Version), ErrVersion
	}

	if (cfg.GeodbPath.City != "" && cfg.GeodbPath.ASN == "") || (cfg.GeodbPath.City == "" && cfg.GeodbPath.ASN != "") {
		return cfg, "", errors.New("both --geoip2-city and --geoip2-asn are mandatory to enable geo information")
	}

	if cfg.TrustedPortHeader != "" && cfg.TrustedHeader == "" {
		return cfg, "", errors.New("trusted-header is mandatory when trusted-port-header is set")
	}

	if (cfg.TLSAddress != "") && (cfg.TLSCrtPath == "" || cfg.TLSKeyPath == "") {
		return cfg, "", errors.New("in order to use TLS, the -tls-crt and -tls-key flags are mandatory")
	}

	if cfg.EnableHTTP3 && cfg.TLSAddress == "" {
		return cfg, "", errors.New("in order to use HTTP3, the -tls-bind is mandatory")
	}

	if cfg.TemplatePath != "" {
		info, err := os.Stat(cfg.TemplatePath)
		if err != nil {
			return cfg, "", fmt.Errorf("template path: %w", err)
		}
		if info.IsDir() {
			return cfg, "", fmt.Errorf("%s must be a file", cfg.TemplatePath)
		}
	}

	if resolverConf != "" {
		var err error
		cfg.Resolver, err = readYAML(resolverConf)
		if err != nil {
			return cfg, "", fmt.Errorf("reading resolver configuration: %w", err)
		}
	}

	return cfg, buf.String(), nil
}

func readYAML(path string) (resolver resolver, err error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return resolver, err
	}
	return resolver, yaml.Unmarshal(yamlFile, &resolver)
}
