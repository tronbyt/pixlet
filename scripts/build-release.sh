#!/bin/bash

if [ -z "$RELEASE_ARCHS" ]; then
	echo "Please set RELEASE_ARCHS"
	exit 1
fi

if [ -z "$RELEASE_PLATFORM" ]; then
	echo "Please set RELEASE_PLATFORM"
	exit 1
fi

for ARCH in $RELEASE_ARCHS
do
	if [[ $ARCH == *arm*  ]]; then
		RELEASE_ARCH=arm64
	else
		RELEASE_ARCH=amd64
	fi

	echo "Building ${RELEASE_PLATFORM}_${RELEASE_ARCH}"

	PIXLET=pixlet
	LIBPIXLET=libpixlet.so
	if [[ $ARCH == "linux-arm64"  ]]; then
		echo "linux-arm64"
		CGO_ENABLED=1 GOOS=$RELEASE_PLATFORM GOARCH=$RELEASE_ARCH go build -ldflags="-s '-extldflags=-static -lsharpyuv' -X 'tidbyt.dev/pixlet/cmd.Version=${PIXLET_VERSION}'" -o "build/${RELEASE_PLATFORM}_${RELEASE_ARCH}/${PIXLET}" tidbyt.dev/pixlet
		CGO_ENABLED=1 GOOS=$RELEASE_PLATFORM GOARCH=$RELEASE_ARCH go build -tags lib -o "build/${RELEASE_PLATFORM}_${RELEASE_ARCH}/${LIBPIXLET}" -buildmode=c-shared library/library.go
	elif [[ $ARCH == "linux-amd64"  ]]; then
		echo "linux-amd64"
		CGO_ENABLED=1 GOOS=$RELEASE_PLATFORM GOARCH=$RELEASE_ARCH go build -ldflags="-s '-extldflags=-static -lsharpyuv' -X 'tidbyt.dev/pixlet/cmd.Version=${PIXLET_VERSION}'" -o "build/${RELEASE_PLATFORM}_${RELEASE_ARCH}/${PIXLET}" tidbyt.dev/pixlet
		CGO_ENABLED=1 GOOS=$RELEASE_PLATFORM GOARCH=$RELEASE_ARCH go build -tags lib -o "build/${RELEASE_PLATFORM}_${RELEASE_ARCH}/${LIBPIXLET}" -buildmode=c-shared library/library.go
	elif [[ $ARCH == "windows-amd64"  ]]; then
		echo "windows-amd64"
		PIXLET=pixlet.exe
		LIBPIXLET=pixlet.dll
		go build -ldflags="-s '-extldflags=-static -lsharpyuv' -X 'tidbyt.dev/pixlet/cmd.Version=${PIXLET_VERSION}'" -tags timetzdata -o build/${RELEASE_PLATFORM}_${RELEASE_ARCH}/${PIXLET} tidbyt.dev/pixlet
		go build -ldflags="-s '-extldflags=-static -lsharpyuv'" -tags lib -o "build/${RELEASE_PLATFORM}_${RELEASE_ARCH}/${LIBPIXLET}" library/library.go
	else
		echo "other"
		CGO_CFLAGS="-I/tmp/${LIBWEBP_VERSION}/${ARCH}/include" CGO_LDFLAGS="-L/tmp/${LIBWEBP_VERSION}/${ARCH}/lib" CGO_ENABLED=1 GOOS=$RELEASE_PLATFORM GOARCH=$RELEASE_ARCH go build -ldflags="-X 'tidbyt.dev/pixlet/cmd.Version=${PIXLET_VERSION}'" -o "build/${RELEASE_PLATFORM}_${RELEASE_ARCH}/${PIXLET}" tidbyt.dev/pixlet
		CGO_CFLAGS="-I/tmp/${LIBWEBP_VERSION}/${ARCH}/include" CGO_LDFLAGS="-L/tmp/${LIBWEBP_VERSION}/${ARCH}/lib" CGO_ENABLED=1 GOOS=$RELEASE_PLATFORM GOARCH=$RELEASE_ARCH go build -tags lib -o build/${RELEASE_PLATFORM}_${RELEASE_ARCH}/libpixlet.so library/library.go
	fi

	echo "Built ./build/${RELEASE_PLATFORM}_${RELEASE_ARCH}/${PIXLET} successfully"
	tar -cvz -C "build/${RELEASE_PLATFORM}_${RELEASE_ARCH}" -f "build/pixlet_${PIXLET_VERSION}_${RELEASE_PLATFORM}_${RELEASE_ARCH}.tar.gz" ${PIXLET} ${LIBPIXLET}
done
