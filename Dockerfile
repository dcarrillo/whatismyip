FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG ARG_VERSION
ARG TARGETOS
ARG TARGETARCH
ENV VERSION=$ARG_VERSION

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/ go mod download -x
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/ apk --no-cache add make ca-certificates \
    && update-ca-certificates \
    && GOOS=$TARGETOS GOARCH=$TARGETARCH make build

FROM scratch
COPY --from=builder /app/whatismyip /usr/bin/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["whatismyip"]
