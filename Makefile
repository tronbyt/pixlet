GIT_COMMIT = $(shell git rev-list -1 HEAD)
ARCH = $(shell uname -m)
OS = $(shell uname -s)
GO_CMD = go

ifeq ($(OS),Windows_NT)
	BINARY = pixlet.exe
	LIBRARY = pixlet.dll
	LDFLAGS = -ldflags="-s '-extldflags=-static -lsharpyuv' -X 'tidbyt.dev/pixlet/cmd.Version=$(GIT_COMMIT)'"
	TAGS = -tags timetzdata
else
	BINARY = pixlet
	LIBRARY = libpixlet.so
	ifeq ($(STATIC),1)
		TAGS = -tags netgo,osusergo
		LDFLAGS = -ldflags="-s -w -linkmode=external '-extldflags=-static -lsharpyuv -lm' -X 'tidbyt.dev/pixlet/cmd.Version=$(GIT_COMMIT)'"
		ifeq ($(OS),Linux)
			CGO_LDFLAGS="-Wl,-Bstatic -lwebp -lwebpdemux -lwebpmux -lsharpyuv -Wl,-Bdynamic"
		endif
	else
		TAGS =
		LDFLAGS = -ldflags="-s -w -X 'tidbyt.dev/pixlet/cmd.Version=$(GIT_COMMIT)'"
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
	$(GO_CMD) test -benchmem -benchtime=20s -bench BenchmarkRunAndRender tidbyt.dev/pixlet/encode

build:
	$(GO_CMD) build $(LDFLAGS) $(TAGS) -o $(BINARY) tidbyt.dev/pixlet
	CGO_LDFLAGS=$(CGO_LDFLAGS) $(GO_CMD) build $(LDFLAGS) -tags lib -o $(LIBRARY) -buildmode=c-shared library/library.go

widgets:
	 $(GO_CMD) run runtime/gen/main.go
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
