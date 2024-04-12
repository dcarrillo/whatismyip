# What is my IP address

[![CI](https://github.com/dcarrillo/whatismyip/workflows/CI/badge.svg)](https://github.com/dcarrillo/whatismyip/actions)
[![CodeQL](https://github.com/dcarrillo/whatismyip/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/dcarrillo/whatismyip/actions/workflows/codeql-analysis.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dcarrillo/whatismyip)](https://goreportcard.com/report/github.com/dcarrillo/whatismyip)
[![GitHub release](https://img.shields.io/github/release/dcarrillo/whatismyip.svg)](https://github.com/dcarrillo/whatismyip/releases/)
[![License Apache 2.0](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](./LICENSE)

- [What is my IP address](#what-is-my-ip-address)
  - [Features](#features)
  - [Endpoints](#endpoints)
  - [DNS discovery](#dns-discovery)
  - [Build](#build)
  - [Usage](#usage)
  - [Examples](#examples)
    - [Run a default TCP server](#run-a-default-tcp-server)
    - [Run a TLS (HTTP/2) and enable What is my DNS](#run-a-tls-http2-and-enable-what-is-my-dns)
    - [Run an HTTP/3 server](#run-an-http3-server)
    - [Run a default TCP server with a custom template and trust a pair of custom headers set by an upstream proxy](#run-a-default-tcp-server-with-a-custom-template-and-trust-a-pair-of-custom-headers-set-by-an-upstream-proxy)
  - [Download](#download)
  - [Docker](#docker)
    - [Run a container locally using test databases](#run-a-container-locally-using-test-databases)
    - [From Docker Hub](#from-docker-hub)

> [!NOTE]  
> Since version 2.3.0, the application includes an optional client [DNS discovery](#dns-discovery)

Just another "what is my IP address" service, including geolocation, TCP open port checking, and headers information. Written in go with high performance in mind,
it uses [gin](https://github.com/gin-gonic/gin) which uses [httprouter](https://github.com/julienschmidt/httprouter) a lightweight high performance HTTP multiplexer.

Take a look at [ifconfig.es](https://ifconfig.es) a live site using `whatismyip` and the `DNS discovery` enabled.

Get your public IP easily from the command line:

```text
curl ifconfig.es
127.0.0.1

curl -6 ifconfig.es
::1
```

Get the IP of your DNS provider:

```text
curl -L dns.ifconfig.es
2a04:e4c0:47::67 (Spain / OPENDNS)
```

## Features

- TLS and HTTP/2.
- Experimental HTTP/3 support. HTTP/3 requires a TLS server running (`-tls-bind`), as HTTP/3 starts as a TLS connection that then gets upgraded to UDP. The UDP port is the same as the one used for the TLS server.
- Beta DNS discovery: A best-effort approach to discovering the DNS server that is resolving the client's requests.
- Can run behind a proxy by trusting a custom header (usually `X-Real-IP`) to figure out the source IP address. It also supports a custom header to resolve the client port, if the proxy can only add a header for the IP (for example a fixed header from CDNs) the client port is shown as unknown.
- IPv4 and IPv6.
- Geolocation info including ASN. This feature is possible thanks to [maxmind](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data?lang=en) GeoLite2 databases. In order to use these databases, a license key is needed. Please visit Maxmind site for further instructions and get a free license.
- Checking TCP open ports.
- High performance.
- Self-contained server that can reload GeoLite2 databases and/or SSL certificates without stop/start. The `hup` signal is honored.
- HTML templates for the landing page.
- Text plain and JSON output.

## Endpoints

- https://ifconfig.es/
- https://ifconfig.es/client-port
- https://ifconfig.es/json (this is the same as `curl -H "Accept: application/json" https://ifconfig.es/`)
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
- https://ifconfig.es/scan/tcp/<port_number>
- https://dns.ifconfig.es

## DNS discovery

The DNS discovery works by forcing the client to make a request to `<uuid>.dns.ifconfig.es` this DNS request is handled by a microdns server
included in the `whatismyip` binary. In order to run the discovery server, a configuration file in the following form has to be created:

```yaml
---
domain: dns.example.com
redirect_port: ":8000"
resource_records:
  - "1800 IN SOA xns.example.com. hostmaster.example.com. 1 10000 2400 604800 1800"
  - "3600 IN NS xns.example.com."
ipv4:
  - "127.0.0.2"
ipv6:
  - "aaa:aaa:aaa:aaaa::1"
```

The DNS authority for example.com has delegated the subdomain zone `dns.example.com` to the server running the `whatismyip` service.

The client can request the URL `dns.example.com` by following the redirection `curl -L dns.example.com`.

To avoid the redirection, you can provide a valid URL, for example, for the real [ifconfig.es](https://ifconfig.es):

```bash
curl $(uuidgen).dns.ifconfig.es

curl $(cat /proc/sys/kernel/random/uuid).dns.ifconfig.es
```


## Build

Golang >= 1.19 is required.

`make build`

## Usage

```text
Usage of whatismyip:
  -bind string
        Listening address (see https://pkg.go.dev/net?#Listen) (default ":8080")
  -enable-http3
        Enable HTTP/3 protocol. HTTP/3 requires --tls-bind set, as HTTP/3 starts as a TLS connection that then gets upgraded to UDP. The UDP port is the same as the one used for the TLS server.
  -enable-secure-headers
        Add sane security-related headers to every response
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
        Trusted request header for remote IP (e.g. X-Real-IP). When using this feature if -trusted-port-header is not set the client port is shown as 'unknown'
  -trusted-port-header string
        Trusted request header for remote client port (e.g. X-Real-Port). When this parameter is set -trusted-header becomes mandatory
  -version
        Output version information and exit
```

## Examples

### Run a default TCP server

```bash
./whatismyip -geoip2-city ./test/GeoIP2-City-Test.mmdb -geoip2-asn ./test/GeoLite2-ASN-Test.mmdb
```

### Run a TLS (HTTP/2) and enable What is my DNS

```bash
./whatismyip -geoip2-city ./test/GeoIP2-City-Test.mmdb -geoip2-asn ./test/GeoLite2-ASN-Test.mmdb \
             -bind "" -tls-bind :8081 -tls-crt ./test/server.pem -tls-key ./test/server.key \
             -resolver ./test/resolver.yml
```

### Run an HTTP/3 server

```bash
./whatismyip -geoip2-city ./test/GeoIP2-City-Test.mmdb -geoip2-asn ./test/GeoLite2-ASN-Test.mmdb \
             -bind "" -tls-bind :8081 -tls-crt ./test/server.pem -tls-key ./test/server.key -enable-http3
```

### Run a default TCP server with a custom template and trust a pair of custom headers set by an upstream proxy

```bash
./whatismyip -geoip2-city ./test/GeoIP2-City-Test.mmdb -geoip2-asn ./test/GeoLite2-ASN-Test.mmdb \
             -trusted-header X-Real-IP -trusted-port-header X-Real-Port -template mytemplate.tmpl
```

## Download

Download the latest version from [github](https://github.com/dcarrillo/whatismyip/releases)

## Docker

An ultra-light (~4MB) image is available on [docker hub](https://hub.docker.com/r/dcarrillo/whatismyip). Since version `2.1.2`, the binary is compressed using [upx](https://github.com/upx/upx).

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
