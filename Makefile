GIT_COMMIT = $(shell git rev-list -1 HEAD)
ARCH = $(shell uname -m)
OS = $(shell uname -s)
GO_CMD = go

ifeq ($(OS),Windows_NT)
	BINARY = pixlet.exe
	LIBRARY = pixlet.dll
	LDFLAGS = -ldflags="-s '-extldflags=-static -lsharpyuv' -X 'github.com/tronbyt/pixlet/cmd.Version=$(GIT_COMMIT)'"
	TAGS = -tags timetzdata,gzip_fonts
else
	BINARY = pixlet
	LIBRARY = libpixlet.so
	ifeq ($(STATIC),1)
		TAGS = -tags netgo,osusergo,gzip_fonts
		LDFLAGS = -ldflags="-s -w -linkmode=external '-extldflags=-static -lsharpyuv -lm' -X 'github.com/tronbyt/pixlet/cmd.Version=$(GIT_COMMIT)'"
		ifeq ($(OS),Linux)
			CGO_LDFLAGS="-Wl,-Bstatic -lwebp -lwebpdemux -lwebpmux -lsharpyuv -Wl,-Bdynamic"
		endif
	else
		TAGS = -tags gzip_fonts
		LDFLAGS = -ldflags="-s -w -X 'github.com/tronbyt/pixlet/cmd.Version=$(GIT_COMMIT)'"
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
	CGO_LDFLAGS=$(CGO_LDFLAGS) $(GO_CMD) build $(LDFLAGS) -tags lib -o $(LIBRARY) -buildmode=c-shared library/library.go

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
	$(GO_CMD) install github.com/bazelbuild/buildtools/buildifier@latest

lint:
	@ buildifier --version >/dev/null 2>&1 || $(MAKE) install-buildifier
	buildifier -r ./

format: lint

gzip_fonts:
	$(GO_CMD) generate -x ./fonts

emoji:
	$(GO_CMD) run ./render/gen
