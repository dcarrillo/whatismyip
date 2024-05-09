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

type geodbPath struct {
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

type settings struct {
	GeodbPath           geodbPath
	TemplatePath        string
	BindAddress         string
	TLSAddress          string
	TLSCrtPath          string
	TLSKeyPath          string
	TrustedHeader       string
	TrustedPortHeader   string
	EnableSecureHeaders bool
	EnableHTTP3         bool
	Server              serverSettings
	Resolver            resolver
	version             bool
}

const defaultAddress = ":8080"

var ErrVersion = errors.New("setting: version requested")

var App = settings{
	// hard-coded for the time being
	Server: serverSettings{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	},
}

func Setup(args []string) (output string, err error) {
	flags := flag.NewFlagSet("whatismyip", flag.ContinueOnError)
	var buf bytes.Buffer
	var resolverConf string
	flags.SetOutput(&buf)

	flags.StringVar(&App.GeodbPath.City, "geoip2-city", "", "Path to GeoIP2 city database")
	flags.StringVar(&App.GeodbPath.ASN, "geoip2-asn", "", "Path to GeoIP2 ASN database")
	flags.StringVar(&App.TemplatePath, "template", "", "Path to the template file")
	flags.StringVar(
		&resolverConf,
		"resolver",
		"",
		"Path to the resolver configuration. It actually enables the resolver for DNS client discovery.")
	flags.StringVar(
		&App.BindAddress,
		"bind",
		defaultAddress,
		"Listening address (see https://pkg.go.dev/net?#Listen)",
	)
	flags.StringVar(
		&App.TLSAddress,
		"tls-bind",
		"",
		"Listening address for TLS (see https://pkg.go.dev/net?#Listen)",
	)
	flags.StringVar(&App.TLSCrtPath, "tls-crt", "", "When using TLS, path to certificate file")
	flags.StringVar(&App.TLSKeyPath, "tls-key", "", "When using TLS, path to private key file")
	flags.StringVar(
		&App.TrustedHeader,
		"trusted-header",
		"",
		"Trusted request header for remote IP (e.g. X-Real-IP). When using this feature if -trusted-port-header is not set the client port is shown as 'unknown'",
	)
	flags.StringVar(
		&App.TrustedPortHeader,
		"trusted-port-header",
		"",
		"Trusted request header for remote client port (e.g. X-Real-Port). When this parameter is set -trusted-header becomes mandatory",
	)
	flags.BoolVar(&App.version, "version", false, "Output version information and exit")
	flags.BoolVar(
		&App.EnableSecureHeaders,
		"enable-secure-headers",
		false,
		"Add sane security-related headers to every response",
	)
	flags.BoolVar(
		&App.EnableHTTP3,
		"enable-http3",
		false,
		"Enable HTTP/3 protocol. HTTP/3 requires --tls-bind set, as HTTP/3 starts as a TLS connection that then gets upgraded to UDP. The UDP port is the same as the one used for the TLS server.",
	)

	err = flags.Parse(args)
	if err != nil {
		return buf.String(), err
	}

	if App.version {
		return fmt.Sprintf("whatismyip version %s", core.Version), ErrVersion
	}

	if App.TrustedPortHeader != "" && App.TrustedHeader == "" {
		return "", fmt.Errorf("truster-header is mandatory when truster-port-header is set")
	}

	if App.GeodbPath.City == "" || App.GeodbPath.ASN == "" {
		return "", fmt.Errorf("geoip2-city and geoip2-asn parameters are mandatory")
	}

	if (App.TLSAddress != "") && (App.TLSCrtPath == "" || App.TLSKeyPath == "") {
		return "", fmt.Errorf("in order to use TLS, the -tls-crt and -tls-key flags are mandatory")
	}

	if App.EnableHTTP3 && App.TLSAddress == "" {
		return "", fmt.Errorf("in order to use HTTP3, the -tls-bind is mandatory")
	}

	if App.TemplatePath != "" {
		info, err := os.Stat(App.TemplatePath)
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%s no such file or directory", App.TemplatePath)
		}
		if info.IsDir() {
			return "", fmt.Errorf("%s must be a file", App.TemplatePath)
		}
	}

	if resolverConf != "" {
		var err error
		App.Resolver, err = readYAML(resolverConf)
		if err != nil {
			return "", fmt.Errorf("error reading resolver configuration %s", err)
		}
	}

	return buf.String(), nil
}

func readYAML(path string) (resolver resolver, err error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return resolver, err
	}
	return resolver, yaml.Unmarshal(yamlFile, &resolver)
}
