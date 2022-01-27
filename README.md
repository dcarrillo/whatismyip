# What is my IP address

[![CI](https://github.com/dcarrillo/whatismyip/workflows/CI/badge.svg)](https://github.com/dcarrillo/whatismyip/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/dcarrillo/whatismyip)](https://goreportcard.com/report/github.com/dcarrillo/whatismyip)
[![GitHub release](https://img.shields.io/github/release/dcarrillo/whatismyip.svg)](https://github.com/dcarrillo/whatismyip/releases/)
[![License Apache 2.0](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](./LICENSE)

- [What is my IP address](#what-is-my-ip-address)
  - [Features](#features)
  - [Endpoints](#endpoints)
  - [Build](#build)
  - [Usage](#usage)
  - [Examples](#examples)
    - [Run a default TCP server](#run-a-default-tcp-server)
    - [Run a TLS (HTTP/2) server only](#run-a-tls-http2-server-only)
    - [Run a default TCP server with a custom template and trust a custom header set by an upstream proxy](#run-a-default-tcp-server-with-a-custom-template-and-trust-a-custom-header-set-by-an-upstream-proxy)
  - [Download](#download)
  - [Docker](#docker)
    - [Run a container locally using test databases](#run-a-container-locally-using-test-databases)
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

- TLS and HTTP/2.
- Can run behind a proxy by trusting a custom header (usually `X-Real-IP`) to figure out the source IP address.
- IPv4 and IPv6.
- Geolocation info including ASN. This feature is possible thanks to [maxmind](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data?lang=en) GeoLite2 databases. In order to use these databases, a license key is needed. Please visit Maxmind site for further instructions and get a free license.
- High performance.
- Self-contained server what can reload GeoLite2 databases and/or SSL certificates without stop/start. The `hup` signal is honored.
- HTML templates for the landing page.
- Text plain and JSON output.

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

## Examples

### Run a default TCP server

```bash
./whatismyip -geoip2-city ./test/GeoIP2-City-Test.mmdb -geoip2-asn ./test/GeoLite2-ASN-Test.mmdb
```

### Run a TLS (HTTP/2) server only

```bash
./whatismyip -geoip2-city ./test/GeoIP2-City-Test.mmdb -geoip2-asn ./test/GeoLite2-ASN-Test.mmdb \
             -bind "" -tls-bind :8081 -tls-crt ./test/server.pem -tls-key ./test/server.key
```

### Run a default TCP server with a custom template and trust a custom header set by an upstream proxy

```bash
./whatismyip -geoip2-city ./test/GeoIP2-City-Test.mmdb -geoip2-asn ./test/GeoLite2-ASN-Test.mmdb \
             -trusted-header X-Real-IP -template mytemplate.tmpl
```

## Download

Download latest version from https://github.com/dcarrillo/whatismyip/releases

## Docker

An ultra-light (~9MB) image is available.

### Run a container locally using test databases

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
