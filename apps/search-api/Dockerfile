# syntax=docker/dockerfile:1.7

FROM golang:1.26.5-trixie AS builder

WORKDIR /src

COPY apps/search-api/go.mod ./
RUN go mod download

COPY apps/search-api/ ./

ARG VERSION=dev

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -trimpath \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /out/search-api \
      ./cmd/server

FROM alpine:3.22 AS runtime

RUN addgroup -S groundedsearch \
    && adduser -S -G groundedsearch -H -s /sbin/nologin groundedsearch

COPY --from=builder /out/search-api /usr/local/bin/search-api

USER groundedsearch

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/search-api"]
