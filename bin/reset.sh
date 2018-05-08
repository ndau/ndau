#!/bin/bash

SCRIPTPATH="$(cd "$(dirname "$0")" ; pwd -P )"
$SCRIPTPATH/stop.sh
source $SCRIPTPATH/defaults.sh 

tendermint unsafe_reset_all
rm -rfv $TMHOME
rm -rfv $NDAUHOME
