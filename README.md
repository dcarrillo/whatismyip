# What is my IP address

- [What is my IP address](#what-is-my-ip-address)
  - [Features](#features)
  - [Endpoints](#endpoints)
  - [Build](#build)
  - [Usage](#usage)
  - [Docker](#docker)
    - [Running a container locally using test databases](#running-a-container-locally-using-test-databases)
    - [From Docker Hub](#from-docker-hub)

Just another "what is my IP address" service, including geolocation and headers information, written in go with high performance in mind, it uses [gin](https://github.com/gin-gonic/gin) which uses [httprouter](https://github.com/julienschmidt/httprouter) a lightweight high performance HTTP multiplexer.

Take a look at [ifconfig.es](https://ifconfig.es) a live site using `whatismyip`

Get your public IP easily from the command line:

```bash
curl ifconfig.es
127.0.0.1

curl -6 ifconfig.es
::1
```

## Features

- TLS available
- Can run behind a proxy by trusting a custom header (usually `X-Real-IP`) to figure out the source IP address.
- IPv4 and IPv6.
- Geolocation info including ASN. This feature is possible thanks to [maxmind](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data?lang=en) GeoLite2 databases. In order to use these databases, a license key is needed. Please visit Maxmind site for further instructions and get a free license.
- High performance
- Although a docker image is provided the executable can reload databases and/or SSL certificates by itself, `hup` signal is honored.
- HTML with templates, text plain and JSON output.

## Endpoints

- https://ifconfig.es/
- https://ifconfig.es/json
- https://ifconfig.es/geo
  - https://ifconfig.es/geo/city
  - https://ifconfig.es/geo/country
  - https://ifconfig.es/geo/country_code
  - https://ifconfig.es/geo/latitude
  - https://ifconfig.es/geo/longitude
  - https://ifconfig.es/geo/postal_code
  - https://ifconfig.es/geo/time_zone
- https://ifconfig.es/asn
  - https://ifconfig.es/asn/number
  - https://ifconfig.es/asn/organization
- https://ifconfig.es/all
- https://ifconfig.es/headers
  - https://ifconfig.es/<header_name>

## Build

Golang >= 1.17 is required. Previous versions may work.

`make build`

## Usage

```text
Usage of ./whatismyip:
  -bind string
        Listening address (see https://pkg.go.dev/net?#Listen) (default ":8080")
  -geoip2-asn string
        Path to GeoIP2 ASN database
  -geoip2-city string
        Path to GeoIP2 city database
  -template string
        Path to template file
  -tls-bind string
        Listening address for TLS (see https://pkg.go.dev/net?#Listen)
  -tls-crt string
        When using TLS, path to certificate file
  -tls-key string
        When using TLS, path to private key file
  -trusted-header string
        Trusted request header for remote IP (e.g. X-Real-IP)
  -version
        Output version information and exit
```

## Docker

An ultra-light (13MB) image is available.

### Running a container locally using test databases

`make docker-run`

### From Docker Hub

```bash
docker run --tty --interactive --rm \
    -v $PWD/<path to city database>:/tmp/GeoIP2-City-Test.mmdb:ro \
    -v $PWD/<path to ASN database>:/tmp/GeoLite2-ASN-Test.mmdb:ro -p 8080:8080 \
    dcarrillo/whatismyip:latest \
    -geoip2-city /tmp/GeoIP2-City-Test.mmdb \
    -geoip2-asn /tmp/GeoLite2-ASN-Test.mmdb
```
