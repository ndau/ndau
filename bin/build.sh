#!/bin/bash

SCRIPTPATH="$(cd "$(dirname "$0")" ; pwd -P )"
$SCRIPTPATH/defaults.sh docker-compose build
