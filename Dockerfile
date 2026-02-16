# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.9.0 AS xx

FROM --platform=$BUILDPLATFORM node:25-alpine AS frontend
WORKDIR /app

COPY frontend/package*.json .
RUN npm ci

COPY frontend .
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.26.0-alpine AS builder
WORKDIR /pixlet

RUN --mount=type=cache,target=/var/cache/apk <<EOT
  set -eux
  apk add --no-cache \
    ca-certificates \
    clang \
    git \
    make \
    tzdata
EOT

COPY go.mod go.sum ./
RUN go mod download

COPY --from=xx / /

ARG TARGETPLATFORM
RUN --mount=type=cache,target=/var/cache/apk,sharing=private <<EOT
  set -eux
  xx-apk add --no-cache gcc g++ libwebp-dev libwebp-static
EOT

COPY . .
COPY --from=frontend /app/dist frontend/dist
ARG PIXLET_VERSION
RUN STATIC=1 CC=xx-clang CGO_ENABLED=1 GO_CMD=xx-go make build

FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /pixlet/pixlet /bin/pixlet

ENTRYPOINT ["/bin/pixlet"]
