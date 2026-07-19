GOPATH ?= $(shell go env GOPATH)
VERSION ?= devel-$(shell git rev-parse --short HEAD)
DOCKER_URL ?= dcarrillo/whatismyip

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
	CGO_ENABLED=0 go build -ldflags="-s -w -X 'github.com/dcarrillo/whatismyip/internal/core.Version=${VERSION}'" -o whatismyip ./cmd

docker-build-dev:
	docker build --target=dev --build-arg=ARG_VERSION="${VERSION}" --tag ${DOCKER_URL}:${VERSION} .

docker-build-prod:
	docker build --target=prod --build-arg=ARG_VERSION="${VERSION}" --tag ${DOCKER_URL}:${VERSION} .

docker-push: docker-build-prod
ifneq (,$(findstring devel-,$(VERSION)))
	@echo "VERSION is set to ${VERSION}, I can't push devel builds"
	exit 1
else
	docker push ${DOCKER_URL}:${VERSION}
	docker tag ${DOCKER_URL}:${VERSION} ${DOCKER_URL}:latest
	docker push ${DOCKER_URL}:latest
endif

docker-run: docker-build-dev
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
