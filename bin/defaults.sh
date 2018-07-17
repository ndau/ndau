#!/bin/bash

# set default values for TMHOME and NDAUHOME if they are not already set

# ensure that $TMHOME is set
# tendermint config stuff lives here
if [ -z $TMHOME ]; then
    export TMHOME=~/.tendermint/
fi
>&2 echo "TMHOME=$TMHOME"

#ensure that $NDAUHOME is set
# ndau config stuff lives here--ndwhitelist needs to edit this, etc.
if [ -z $NDAUHOME ]; then
    export NDAUHOME=~/.ndau/
fi
>&2 echo "NDAUHOME=$NDAUHOME"

"$@"
