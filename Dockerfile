FROM golang:1.17-alpine as builder

WORKDIR /app

COPY . .

RUN apk add make && make build

# Build final image
FROM scratch

WORKDIR /app

COPY --from=builder /app/whatismyip /usr/bin/

EXPOSE 8080

ENTRYPOINT ["whatismyip"]
