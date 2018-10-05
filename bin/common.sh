#!/bin/bash
# This file contains functions and helpers that are common to many scripts.

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # no color

# errcho prints an error message to stderr.
errcho() {
    >&2 echo -e "$@"
}

# err prints an error message to stderr in red.
err() {
    errcho "${RED}${1}: " "${@:2}" "${NC}"
    exit 1
}

# echo_green prints a message in green
echo_green() {
    errcho "$GREEN" "$@" "$NC"
}

# default tendermint RPC port
export TM_RPC=26657
