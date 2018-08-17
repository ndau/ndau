#!/bin/bash

SCRIPTPATH="$(cd "$(dirname "$0")" ; pwd -P )"
ndau=$(which ndau)
if [ -x "$SCRIPTPATH/../cmd/ndau/ndau" ]; then
    ndau="$SCRIPTPATH/../cmd/ndau/ndau"
fi
if [ -x "$SCRIPTPATH/../ndau" ]; then
    ndau="$SCRIPTPATH/../ndau"
fi

if [ -z "$ndau" ]; then
    echo "ndau executable not found"
    exit 1
fi

nn="$SCRIPTPATH/.."
if cd "$nn"; then
    $ndau conf $(bin/defaults.sh docker-compose port tendermint 26657)
else
    echo "ndaunode not found"
    exit 1
fi

