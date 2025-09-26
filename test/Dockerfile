FROM golang:1.25-alpine AS builder

ARG ARG_VERSION
ENV VERSION=$ARG_VERSION

WORKDIR /app
COPY . .

FROM builder AS build-test-app
RUN CGO_ENABLED=0 \
    go build -ldflags="-s -w" -o whatismyip ./cmd

FROM scratch AS test
COPY --from=build-test-app /app/whatismyip /usr/bin/
ENTRYPOINT ["whatismyip"]
