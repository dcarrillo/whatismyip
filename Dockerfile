FROM golang:1.20-alpine as builder

ARG ARG_VERSION
ENV VERSION $ARG_VERSION

WORKDIR /app

COPY . .

RUN apk add make git upx && make build VERSION=$VERSION \
    && upx --best --lzma whatismyip

# Build final image
FROM scratch

WORKDIR /app

COPY --from=builder /app/whatismyip /usr/bin/

EXPOSE 8080

ENTRYPOINT ["whatismyip"]
