GOPATH ?= $(shell go env GOPATH)
VERSION ?= devel-$(shell git rev-parse --short HEAD)
DOCKER_URL ?= dcarrillo/whatismyip

.PHONY: test
test: unit-test integration-test

.PHONY: unit-test
unit-test:
	go test -race -short -cover ./...

.PHONY: integration-test
integration-test:
	go test ./integration-tests -v

.PHONY: install-tools
install-tools:
	@command golangci-lint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.43.0; \
	fi

	@command $(GOPATH)/shadow > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@v0.1.7; \
	fi
.PHONY: lint
lint: install-tools
	golangci-lint run
	shadow ./...

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-X 'github.com/dcarrillo/whatismyip/internal/core.Version=${VERSION}'" -o whatismyip ./cmd

.PHONY: docker-build
docker-build:
	docker build --tag ${DOCKER_URL}:${VERSION} .

.PHONY: docker-run
docker-run: docker-build
	docker run --tty --interactive --rm \
	-v $$PWD/test/GeoIP2-City-Test.mmdb:/tmp/GeoIP2-City-Test.mmdb:ro \
	-v $$PWD/test/GeoLite2-ASN-Test.mmdb:/tmp/GeoLite2-ASN-Test.mmdb:ro -p 8080:8080 \
	${DOCKER_URL}:${VERSION} \
		-geoip2-city /tmp/GeoIP2-City-Test.mmdb \
		-geoip2-asn /tmp/GeoLite2-ASN-Test.mmdb \
		-trusted-header X-Real-IP
