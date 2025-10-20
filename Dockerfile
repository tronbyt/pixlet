# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.8.0 AS xx

FROM --platform=$BUILDPLATFORM node:24-alpine AS frontend
WORKDIR /app

COPY package.json package-lock.json .
RUN npm ci

COPY src src
COPY public public
COPY webpack.*.js .babelrc.json .
RUN npm run build

# Can't use Alpine because of
# - https://github.com/golang/go/issues/54805: libpixlet.so can't be loaded dynamically
# - https://github.com/python/cpython/issues/109332: CPython doesn't support musl
FROM --platform=$BUILDPLATFORM golang:1.25.3 AS builder
WORKDIR /pixlet

ARG DEBIAN_FRONTEND=noninteractive
RUN --mount=type=cache,target=/var/lib/apt/lists <<EOT
  set -eux
  apt-get update
  apt-get install -y --no-install-recommends \
    ca-certificates \
    clang \
    git \
    make \
    tzdata
EOT

COPY go.mod go.sum .
RUN go mod download

COPY --from=xx / /

ARG TARGETPLATFORM
RUN --mount=type=cache,target=/var/lib/apt/lists,sharing=private <<EOT
  set -eux
  apt-get update
  xx-apt-get install -y --no-install-recommends gcc g++ libwebp-dev
EOT

COPY . .
COPY --from=frontend /app/dist dist
RUN STATIC=1 CC=xx-clang CGO_ENABLED=1 make build GO_CMD=xx-go

FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /pixlet/pixlet /bin/pixlet
COPY --from=builder /pixlet/libpixlet.so /lib/libpixlet.so
COPY --from=builder /pixlet/libpixlet.h /usr/include/libpixlet/libpixlet.h

ENTRYPOINT ["/bin/pixlet"]
