#!/bin/bash

contains_detach_command=0
for arg in "$@"; do
    if [ $arg == "-d" -o $arg == "--detach" ]; then
        contains_detach_command=1
        break
    fi
done
if [ contains_detach_command == 0 ]; then
    aoce="--abort-on-container-exit --exit-code-from tendermint"
fi

SCRIPTPATH="$(cd "$(dirname "$0")" ; pwd -P )"
$SCRIPTPATH/defaults.sh docker-compose up $aoce "$@"
