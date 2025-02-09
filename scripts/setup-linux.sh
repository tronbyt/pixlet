#!/bin/bash

set -e

dpkg --add-architecture arm64

cat <<EOT > /etc/apt/sources.list.d/ubuntu.sources
Types: deb
URIs: http://archive.ubuntu.com/ubuntu/
Suites: noble
Components: main restricted
Architectures: amd64

Types: deb
URIs: http://archive.ubuntu.com/ubuntu/
Suites: noble-updates
Components: main restricted
Architectures: amd64

Types: deb
URIs: http://archive.ubuntu.com/ubuntu/
Suites: noble
Components: universe
Architectures: amd64

Types: deb
URIs: http://archive.ubuntu.com/ubuntu/
Suites: noble-updates
Components: universe
Architectures: amd64

Types: deb
URIs: http://archive.ubuntu.com/ubuntu/
Suites: noble
Components: multiverse
Architectures: amd64

Types: deb
URIs: http://archive.ubuntu.com/ubuntu/
Suites: noble-updates
Components: multiverse
Architectures: amd64

Types: deb
URIs: http://archive.ubuntu.com/ubuntu/
Suites: noble-backports
Components: main restricted universe multiverse
Architectures: amd64

Types: deb
URIs: http://security.ubuntu.com/ubuntu/
Suites: noble-security
Components: main restricted
Architectures: amd64

Types: deb
URIs: http://security.ubuntu.com/ubuntu/
Suites: noble-security
Components: universe
Architectures: amd64

Types: deb
URIs: http://security.ubuntu.com/ubuntu/
Suites: noble-security
Components: multiverse
Architectures: amd64

Types: deb
URIs: http://ports.ubuntu.com/ubuntu-ports
Suites: noble
Components: main restricted
Architectures: arm64

Types: deb
URIs: http://ports.ubuntu.com/ubuntu-ports
Suites: noble-updates
Components: main restricted
Architectures: arm64

Types: deb
URIs: http://ports.ubuntu.com/ubuntu-ports
Suites: noble
Components: universe
Architectures: arm64

Types: deb
URIs: http://ports.ubuntu.com/ubuntu-ports
Suites: noble-updates
Components: universe
Architectures: arm64

Types: deb
URIs: http://ports.ubuntu.com/ubuntu-ports
Suites: noble
Components: multiverse
Architectures: arm64

Types: deb
URIs: http://ports.ubuntu.com/ubuntu-ports
Suites: noble-updates
Components: multiverse
Architectures: arm64

Types: deb
URIs: http://ports.ubuntu.com/ubuntu-ports
Suites: noble-backports
Components: main restricted universe multiverse
Architectures: arm64

Types: deb
URIs: http://ports.ubuntu.com/ubuntu-ports
Suites: noble-security
Components: main restricted
Architectures: arm64

Types: deb
URIs: http://ports.ubuntu.com/ubuntu-ports
Suites: noble-security
Components: universe
Architectures: arm64

Types: deb
URIs: http://ports.ubuntu.com/ubuntu-ports
Suites: noble-security
Components: multiverse
Architectures: arm64
EOT

apt-get update 
apt-get install -y \
    libwebp-dev \
    libwebp-dev:arm64 \
    crossbuild-essential-arm64