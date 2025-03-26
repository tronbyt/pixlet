#!/bin/bash

set -e

apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends libwebp-dev
