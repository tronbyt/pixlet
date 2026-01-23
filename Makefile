PIXLET_VERSION ?= $(shell git rev-parse HEAD)
GO_CMD ?= go
ARCH = $(shell uname -m)
OS = $(shell uname -s)

ifeq ($(OS),Windows_NT)
	BINARY = pixlet.exe
	LDFLAGS = -ldflags="-s '-extldflags=-static -lsharpyuv' -X 'github.com/tronbyt/pixlet/runtime.Version=$(PIXLET_VERSION)'"
	TAGS = -tags timetzdata,gzip_fonts,netgo,osusergo
else
	BINARY = pixlet
	ifeq ($(STATIC),1)
		TAGS = -tags timetzdata,gzip_fonts,netgo,osusergo
		LDFLAGS = -ldflags="-s -w -linkmode=external '-extldflags=-static -lsharpyuv -lm' -X 'github.com/tronbyt/pixlet/runtime.Version=$(PIXLET_VERSION)'"
		ifeq ($(OS),Linux)
			CGO_LDFLAGS="-Wl,-Bstatic -lwebp -lwebpdemux -lwebpmux -lsharpyuv -Wl,-Bdynamic"
		endif
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

widgets:
	 $(GO_CMD) run ./runtime/gen
	 gofmt -s -w ./

install-buildifier:
	$(GO_CMD) install github.com/bazelbuild/buildtools/buildifier@v0.0.0-20260113134051-f026de8858b3

lint:
	@ buildifier --version >/dev/null 2>&1 || $(MAKE) install-buildifier
	buildifier -r ./

format: lint

gzip_fonts:
	$(GO_CMD) generate -x ./fonts
