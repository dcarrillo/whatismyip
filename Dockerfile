FROM golang:1.17-alpine as builder

ARG ARG_VERSION
ENV VERSION $ARG_VERSION

WORKDIR /app

COPY . .

RUN apk add make && make build VERSION=$VERSION

# Build final image
FROM scratch

WORKDIR /app

COPY --from=builder /app/whatismyip /usr/bin/

EXPOSE 8080

ENTRYPOINT ["whatismyip"]
