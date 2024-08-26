FROM golang:1.23-alpine AS builder

ARG ARG_VERSION
ENV VERSION=$ARG_VERSION

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN --mount=type=cache,target=/go/pkg/mod/ go mod download -x
COPY . .

FROM builder AS build-dev-app
# hadolint ignore=DL3018
RUN --mount=type=cache,target=/go/pkg/mod/ apk --no-cache add make && make build

FROM builder AS build-prod-app
# hadolint ignore=DL3018
RUN --mount=type=cache,target=/go/pkg/mod/ apk --no-cache add make upx \
    && make build \
    && upx --best --lzma whatismyip

FROM scratch AS dev
COPY --from=build-dev-app /app/whatismyip /usr/bin/
ENTRYPOINT ["whatismyip"]

FROM scratch AS prod
COPY --from=build-prod-app /app/whatismyip /usr/bin/
ENTRYPOINT ["whatismyip"]
