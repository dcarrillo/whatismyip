GOPATH ?= $(shell go env GOPATH)
VERSION ?= devel-$(shell git rev-parse --short HEAD)
DOCKER_URL ?= dcarrillo/whatismyip
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: test
test: unit-test integration-test

unit-test:
	go test -count=1 -race -short -cover ./...

integration-test:
	cd integration-tests && go test -count=1 -v ./...

install-tools:
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint"; \
		curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(GOPATH)/bin; \
	fi

	@if ! command -v revive &> /dev/null; then \
		echo "Installing revive..."; \
    	go install github.com/mgechev/revive@latest; \
	fi	

	@if ! command -v shadow &> /dev/null; then \
		go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest; \
	fi

lint: install-tools
	gofmt -l . && test -z $$(gofmt -l .)
	golangci-lint run
	shadow ./...

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w -X 'github.com/dcarrillo/whatismyip/internal/core.Version=${VERSION}'" -o whatismyip ./cmd

build-all:
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 $(MAKE) build && mv whatismyip dist/whatismyip-linux-amd64
	GOOS=linux GOARCH=arm64 $(MAKE) build && mv whatismyip dist/whatismyip-linux-arm64
	GOOS=darwin GOARCH=amd64 $(MAKE) build && mv whatismyip dist/whatismyip-darwin-amd64
	GOOS=darwin GOARCH=arm64 $(MAKE) build && mv whatismyip dist/whatismyip-darwin-arm64
	GOOS=windows GOARCH=amd64 $(MAKE) build && mv whatismyip dist/whatismyip-windows-amd64.exe

docker-build:
	docker build --build-arg=ARG_VERSION="${VERSION}" --tag ${DOCKER_URL}:${VERSION} .

docker-run: docker-build
	docker run --tty --interactive --rm \
		--publish 8080:8080/tcp \
		--publish 8081:8081/tcp \
		--publish 9100:9100/tcp \
		--publish 8081:8081/udp \
		--volume ${PWD}/test:/test \
		${DOCKER_URL}:${VERSION} \
		-geoip2-city /test/GeoIP2-City-Test.mmdb \
		-geoip2-asn /test/GeoLite2-ASN-Test.mmdb \
		-trusted-header X-Real-PortReal-IP \
		-tls-bind :8081 \
		-tls-crt /test/server.pem \
		-tls-key /test/server.key \
		-enable-http3 \
		-metrics-bind :9100
