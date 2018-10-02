#!/bin/bash

# This script sets the default values for environment variables if they're not already set.
# Note
# Not having any sourced dependencies in this script makes it sourceable outside of a script.

if [ -z "$TMHOME" ]; then
    # tendermint config
    TMHOME=~/.tendermint/
fi

if [ -z "$NDAUHOME" ]; then
    # ndau config--ndwhitelist needs to edit this, etc.
    NDAUHOME=~/.ndau/
fi

if [ -z "$HONEYCOMB_DATASET" ]; then
    # The honeycomb dataset is a bucket for logs
    HONEYCOMB_DATASET=ndau-dev
fi

if [ -z "$HONEYCOMB_KEY" ]; then
    # The honeycomb key must be set externally and not committed to git.
    >&2 echo "defaults.sh:" "HONEYCOMB_KEY is not set -- logging will be local"
    export HONEYCOMB_KEY=invalidkey
else
    >&2 echo "defaults.sh:" "HONEYCOMB_KEY is set"
fi

export TMHOME
export NDAUHOME
export HONEYCOMB_DATASET
