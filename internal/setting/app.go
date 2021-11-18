package setting

import (
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
	GeodbPath     geodbPath
	TemplatePath  string
	BindAddress   string
	TLSAddress    string
	TLSCrtPath    string
	TLSKeyPath    string
	TrustedHeader string
	Server        serverSettings
}

const defaultAddress = ":8080"

var App *settings

func Setup() {
	city := flag.String("geoip2-city", "", "Path to GeoIP2 city database")
	asn := flag.String("geoip2-asn", "", "Path to GeoIP2 ASN database")
	template := flag.String("template", "", "Path to template file")
	address := flag.String(
		"bind",
		defaultAddress,
		"Listening address (see https://pkg.go.dev/net?#Listen)",
	)
	addressTLS := flag.String(
		"tls-bind",
		"",
		"Listening address for TLS (see https://pkg.go.dev/net?#Listen)",
	)
	tlsCrtPath := flag.String("tls-crt", "", "When using TLS, path to certificate file")
	tlsKeyPath := flag.String("tls-key", "", "When using TLS, path to private key file")
	trustedHeader := flag.String(
		"trusted-header",
		"",
		"Trusted request header for remote IP (e.g. X-Real-IP)",
	)
	ver := flag.Bool("version", false, "Output version information and exit")

	flag.Parse()

	if *ver {
		fmt.Printf("whatismyip version %s", core.Version)
		os.Exit(0)
	}

	if *city == "" || *asn == "" {
		exitWithError("geoip2-city and geoip2-asn parameters are mandatory")
	}

	if (*addressTLS != "") && (*tlsCrtPath == "" || *tlsKeyPath == "") {
		exitWithError("In order to use TLS -tls-crt and -tls-key flags are mandatory")
	}

	if *template != "" {
		info, err := os.Stat(*template)
		if os.IsNotExist(err) || info.IsDir() {
			exitWithError(*template + " doesn't exist or it's not a file")
		}
	}

	App = &settings{
		GeodbPath:     geodbPath{City: *city, ASN: *asn},
		TemplatePath:  *template,
		BindAddress:   *address,
		TLSAddress:    *addressTLS,
		TLSCrtPath:    *tlsCrtPath,
		TLSKeyPath:    *tlsKeyPath,
		TrustedHeader: *trustedHeader,
		Server: serverSettings{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func exitWithError(error string) {
	fmt.Printf("%s\n\n", error)
	flag.Usage()
	os.Exit(1)
}
