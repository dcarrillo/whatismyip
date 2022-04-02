package setting

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dcarrillo/whatismyip/internal/core"
)

type geodbPath struct {
	City string
	ASN  string
}
type serverSettings struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}
type settings struct {
	GeodbPath           geodbPath
	TemplatePath        string
	BindAddress         string
	TLSAddress          string
	TLSCrtPath          string
	TLSKeyPath          string
	TrustedHeader       string
	EnableSecureHeaders bool
	Server              serverSettings
	version             bool
}

const defaultAddress = ":8080"

// ErrVersion is the custom error triggered when -version flag is passed
var ErrVersion = errors.New("setting: version requested")

// App is the var with the parsed settings
var App = settings{
	// hard-coded for the time being
	Server: serverSettings{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	},
}

// Setup initializes the App object parsing the flags
func Setup(args []string) (output string, err error) {
	flags := flag.NewFlagSet("whatismyip", flag.ContinueOnError)
	var buf bytes.Buffer
	flags.SetOutput(&buf)

	flags.StringVar(&App.GeodbPath.City, "geoip2-city", "", "Path to GeoIP2 city database")
	flags.StringVar(&App.GeodbPath.ASN, "geoip2-asn", "", "Path to GeoIP2 ASN database")
	flags.StringVar(&App.TemplatePath, "template", "", "Path to template file")
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
	flags.StringVar(&App.TrustedHeader,
		"trusted-header",
		"",
		"Trusted request header for remote IP (e.g. X-Real-IP)",
	)
	flags.BoolVar(&App.version, "version", false, "Output version information and exit")
	flags.BoolVar(
		&App.EnableSecureHeaders,
		"enable-secure-headers",
		false,
		"Add sane security-related headers to every response",
	)

	err = flags.Parse(args)
	if err != nil {
		return buf.String(), err
	}

	if App.version {
		return fmt.Sprintf("whatismyip version %s", core.Version), ErrVersion
	}

	if App.GeodbPath.City == "" || App.GeodbPath.ASN == "" {
		return "", fmt.Errorf("geoip2-city and geoip2-asn parameters are mandatory")
	}

	if (App.TLSAddress != "") && (App.TLSCrtPath == "" || App.TLSKeyPath == "") {
		return "", fmt.Errorf("In order to use TLS -tls-crt and -tls-key flags are mandatory")
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

	return buf.String(), nil
}
