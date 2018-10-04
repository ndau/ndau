#!/bin/bash

set -e # quit for errors

ROOT="$(cd "$(dirname "$0")/.." || exit 1; pwd -P )"
me=$(basename "$0")  # get tag

# shellcheck source=./common.sh
source "$ROOT"/bin/common.sh

# shellcheck source=./defaults.sh
source "$ROOT"/bin/defaults.sh

# find the ndau tool
ndau="$(which ndau || echo '')"
if [ -x "$ROOT"/ndau ]; then
    ndau="$ROOT"/ndau
else
    if [ -x "$ROOT/cmd/ndau/ndau" ]; then
        ndau="$ROOT/cmd/ndau/ndau"
    fi
fi
if [ -z "$ndau" ]; then
    err "$me" "ndau executable not found."
fi

errcho "$me" "Using ndau: $ndau"

# configure ndau tool
set -x # echo command
$ndau conf "http://$(docker-compose port tendermint "$TM_RPC")"
