#!/bin/bash
# This script uses docker-compose to build down the docker-compose containers.

set -e # stop for errors

ROOT="$(cd "$(dirname "$0")"/..; pwd -P )"

# shellcheck source=./defaults.sh
source "$ROOT"/bin/defaults.sh

set -x # echo command
docker-compose build "$@" && "$ROOT"/bin/tool-build.sh
