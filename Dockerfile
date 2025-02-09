FROM alpine:3.21 AS builder

RUN apk --no-cache add go npm libwebp-dev libwebp-static git make clang musl-dev
COPY . /pixlet
WORKDIR /pixlet
RUN npm install && npm run build && STATIC=1 make build

FROM scratch

COPY --from=builder /pixlet/pixlet /bin/pixlet

ENTRYPOINT ["/bin/pixlet"]
