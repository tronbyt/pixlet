# Can't use Alpine because of
# - https://github.com/golang/go/issues/54805: libpixlet.so can't be loaded dynamically
# - https://github.com/python/cpython/issues/109332: CPython doesn't support musl
FROM golang:1.24.0-bookworm AS builder

RUN echo deb http://deb.debian.org/debian bookworm-backports main > /etc/apt/sources.list.d/bookworm-backports.list && \
    apt-get update && apt install -y --no-install-recommends npm libwebp-dev/bookworm-backports git make clang tzdata ca-certificates && \
    rm -rf /var/lib/apt/lists/*
COPY . /pixlet
WORKDIR /pixlet
RUN npm install && npm run build && STATIC=1 CC=clang make build

FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /pixlet/pixlet /bin/pixlet
COPY --from=builder /pixlet/libpixlet.so /lib/libpixlet.so

ENTRYPOINT ["/bin/pixlet"]
