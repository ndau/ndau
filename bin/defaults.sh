#!/bin/bash

# This script sets the default values for environment variables if they're not already set.
# Note
# Not having any sourced dependencies in this script makes it sourceable outside of a script.

VERSION=$(git describe --long --tags)
>&2 echo "defaults.sh:" "VERSION=$VERSION"

if [ -z "$TMHOME" ]; then
    # tendermint config
    TMHOME=~/.tendermint/
fi
>&2 echo "defaults.sh:" "TMHOME=$TMHOME"

if [ -z "$NDAUHOME" ]; then
    # ndau config--ndwhitelist needs to edit this, etc.
    NDAUHOME=~/.ndau/
fi
>&2 echo "defaults.sh:" "NDAUHOME=$NDAUHOME"

if [ -z "$HONEYCOMB_DATASET" ]; then
    # The honeycomb dataset is a bucket for logs
    HONEYCOMB_DATASET=ndau-dev
fi
>&2 echo "defaults.sh:" "HONEYCOMB_DATASET=$HONEYCOMB_DATASET"

if [ -z "$HONEYCOMB_KEY" ]; then
    # The honeycomb key must be set externally and not committed to git.
    >&2 echo "defaults.sh:" "HONEYCOMB_KEY is not set -- logging will be local"
    export HONEYCOMB_KEY=invalidkey
else
    >&2 echo "defaults.sh:" "HONEYCOMB_KEY is set"
fi

export VERSION
export TMHOME
export NDAUHOME
export HONEYCOMB_DATASET

# after defaults are set, call all arguments.
# this allows inline usage on the command line, i.e.
# $ bin/defaults.sh docker-compose run --rm --no-deps ndaunode --version
"$@"
