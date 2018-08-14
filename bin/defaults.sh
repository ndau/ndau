#!/bin/bash

# set default values for TMHOME and NDAUHOME if they are not already set

# ensure that $TMHOME is set
# tendermint config stuff lives here
if [ -z "$TMHOME" ]; then
    export TMHOME=~/.tendermint/
fi
>&2 echo "TMHOME=$TMHOME"

#ensure that $NDAUHOME is set
# ndau config stuff lives here--ndwhitelist needs to edit this, etc.
if [ -z "$NDAUHOME" ]; then
    export NDAUHOME=~/.ndau/
fi
>&2 echo "NDAUHOME=$NDAUHOME"

#set honeycomb variables
# the honeycomb dataset is the basic bucket into which logs will be dumped
if [ -z "$HONEYCOMB_DATASET" ]; then
    export HONEYCOMB_DATASET=ndau-dev
fi
>&2 echo "HONEYCOMB_DATASET=$HONEYCOMB_DATASET"

# the honeycomb key must not enter the repo; it must be set externally
if [ -z "$HONEYCOMB_KEY" ]; then
    >&2 echo "HONEYCOMB_KEY is not set -- logging will be local"
    export HONEYCOMB_KEY=invalidkey
else
    >&2 echo "HONEYCOMB_KEY is set"
fi

"$@"
