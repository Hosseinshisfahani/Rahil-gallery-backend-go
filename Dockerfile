FROM golang:1.22-alpine AS builder

# Override when proxy.golang.org is blocked (common with VPN):
#   GOPROXY=https://goproxy.io,direct docker compose build api
#   GOPROXY=https://goproxy.cn,direct docker compose build api
# Or vendor locally: make docker-vendor && docker compose -f docker-compose.yml -f docker-compose.vendor.yml build api
ARG GOPROXY=https://goproxy.io,https://proxy.golang.org,direct
ARG GOSUMDB=sum.golang.org
ARG GIT_SHA=dev
ENV GOPROXY=${GOPROXY}
ENV GOSUMDB=${GOSUMDB}

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.20

ARG GIT_SHA=dev
LABEL org.opencontainers.image.revision="${GIT_SHA}"

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app
COPY --from=builder /out/api .
COPY data/catalog-images ./data/catalog-images

ENV CATALOG_ASSETS_DIR=/app/data/catalog-images

EXPOSE 8080

USER nobody

ENTRYPOINT ["/app/api"]
