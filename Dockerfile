FROM golang:1.24.0-alpine3.21 AS builder

RUN apk --no-cache add npm libwebp-dev libwebp-static git make clang musl-dev tzdata ca-certificates
COPY . /pixlet
WORKDIR /pixlet
RUN npm install && npm run build && STATIC=1 CC=clang make build

FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /pixlet/pixlet /bin/pixlet

ENTRYPOINT ["/bin/pixlet"]
