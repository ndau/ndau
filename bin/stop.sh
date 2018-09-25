#!/bin/bash
# This script shuts down the docker-compose containers.
set -e # stop for errors

ROOT="$(cd "$(dirname "$0")/.." || exit 1; pwd -P )"

# shellcheck source=./common.sh
source "$ROOT"/bin/common.sh

# shellcheck source=./defaults.sh
source "$ROOT"/bin/defaults.sh

set -x # echo last command
docker-compose down
