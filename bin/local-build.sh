#!/bin/bash

ROOT="$(cd "$(dirname "$0")/.." || exit 1; pwd -P )"

# shellcheck source=./common.sh
source "$ROOT"/bin/common.sh

# shellcheck source=./defaults.sh
source "$ROOT"/bin/defaults.sh

cd "$ROOT" || exit 1

go build -ldflags "-X github.com/oneiro-ndev/ndau/pkg/version.version=${VERSION}" ./cmd/ndaunode
go build -ldflags "-X github.com/oneiro-ndev/ndau/pkg/version.version=$VERSION" ./cmd/ndau
