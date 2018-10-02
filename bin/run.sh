#!/bin/bash
# This script runs the ndau node using docker-compose.

ROOT="$(cd "$(dirname "$0")/.." || exit 1; pwd -P )"

# shellcheck source=./common.sh
source "$ROOT"/bin/common.sh

# shellcheck source=./defaults.sh
source "$ROOT"/bin/defaults.sh


for arg in "$@"; do
    if [ $arg == "-d" -o $arg == "--detach" ]; then
        aoce="--abort-on-container-exit --exit-code-from tendermint"
    fi
done

set -x # echo command
# shellcheck disable=SC2086 disable=SC2068
docker-compose up $aoce $@
