# Can't use Alpine because of
# - https://github.com/golang/go/issues/54805: libpixlet.so can't be loaded dynamically
# - https://github.com/python/cpython/issues/109332: CPython doesn't support musl
FROM debian:trixie-slim AS builder

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ca-certificates \
        curl && \
    curl -fsSL https://deb.nodesource.com/setup_23.x | bash - && \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        clang \
        git \
        golang-go \
        libwebp-dev \
        make \
        npm \
        tzdata && \
    rm -rf /var/lib/apt/lists/*
COPY . /pixlet
WORKDIR /pixlet
RUN npm install && npm run build && STATIC=1 CC=clang make build

FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /pixlet/pixlet /bin/pixlet
COPY --from=builder /pixlet/libpixlet.so /lib/libpixlet.so
COPY --from=builder /pixlet/libpixlet.h /usr/include/libpixlet/libpixlet.h

ENTRYPOINT ["/bin/pixlet"]
