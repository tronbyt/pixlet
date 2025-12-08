PIXLET_VERSION ?= $(shell git rev-parse HEAD)
GO_CMD ?= go
ARCH = $(shell uname -m)
OS = $(shell uname -s)

LIBRARY_LDFLAGS = -ldflags="-s -w -X 'github.com/tronbyt/pixlet/runtime.Version=$(PIXLET_VERSION)'"
ifeq ($(OS),Windows_NT)
	BINARY = pixlet.exe
	LIBRARY = pixlet.dll
	LDFLAGS = -ldflags="-s -w '-extldflags=-static' -X 'github.com/tronbyt/pixlet/runtime.Version=$(PIXLET_VERSION)'"
	TAGS = -tags timetzdata,gzip_fonts
else
	BINARY = pixlet
	LIBRARY = libpixlet.so
	ifeq ($(STATIC),1)
		TAGS = -tags netgo,osusergo,gzip_fonts
		LDFLAGS = -ldflags="-s -w '-extldflags=-static' -X 'github.com/tronbyt/pixlet/runtime.Version=$(PIXLET_VERSION)'"
	else
		TAGS = -tags gzip_fonts
		LDFLAGS = -ldflags="-s -w -X 'github.com/tronbyt/pixlet/runtime.Version=$(PIXLET_VERSION)'"
	endif
endif

all: build

test:
	$(GO_CMD) test $(TAGS) -v -cover ./...

clean:
	rm -f $(BINARY)
	rm -rf ./build
	rm -rf ./out

bench:
	$(GO_CMD) test -benchmem -benchtime=20s -bench BenchmarkRunAndRender github.com/tronbyt/pixlet/encode

build: gzip_fonts
	$(GO_CMD) build $(LDFLAGS) $(TAGS) -o $(BINARY) github.com/tronbyt/pixlet
	CGO_ENABLED=1 $(GO_CMD) build $(LIBRARY_LDFLAGS) -tags lib,gzip_fonts -o $(LIBRARY) -buildmode=c-shared ./library

widgets:
	 $(GO_CMD) run ./runtime/gen
	 gofmt -s -w ./

release-macos: clean
	./scripts/release-macos.sh

release-linux: clean
	./scripts/release-linux.sh

release-windows: clean
	./scripts/release-windows.sh

install-buildifier:
	$(GO_CMD) install github.com/bazelbuild/buildtools/buildifier@v0.0.0-20251107112229-e879524f2986

lint:
	@ buildifier --version >/dev/null 2>&1 || $(MAKE) install-buildifier
	buildifier -r ./

format: lint

gzip_fonts:
	$(GO_CMD) generate -x ./fonts
