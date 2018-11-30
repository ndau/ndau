#!/bin/bash

ROOT="$(cd "$(dirname "$0")/.." || exit 1; pwd -P )"
me=$(basename "$0") # get tag

# shellcheck source=./common.sh
source "$ROOT"/bin/common.sh

# shellcheck source=./defaults.sh
source "$ROOT"/bin/defaults.sh

# this script depends on jq, sed, and docker-compose
# let's make sure that each is installed before going anywhere with this
dependencies=(jq sed docker-compose)
for tool in "${dependencies[@]}"; do
    if ! command -v "$tool" > /dev/null  ; then
        err "$me" "This script depends on $tool. Install it and try again."
    fi
done

# use working sed or gnu sed
sed="sed"
if ! $sed --version &> /dev/null; then
    if ! which gsed &> /dev/null; then
	    err "$me" "Version of sed not adequate. Try again after: brew install gnu-sed"
    fi
    sed=gsed
    errcho "$me" "using gsed"
fi

errcho "$me" "Initializing tendermint"
docker-compose run --rm --no-deps tendermint init

# make tendermint look for its ABCI app at a machine named ndaunode
config=$TMHOME/config/config.toml
config_backup=${config}.bak
cp "$config" "$config_backup"
$sed -E \
    -e '/^proxy_app/s|://[^:]*:|://ndaunode:|' \
    -e '/^create_empty_blocks_interval/s/[[:digit:]]+/300/' \
    -e '/^create_empty_blocks\b/{
            s/true/false/
            s/(.*)/# \1/
            i # tendermint respects create_empty_blocks *OR* create_empty_blocks_interval
        }' \
    "$config_backup" > "$config"

errcho "$me" "config diff:"
diff "$config_backup" "$config"
rm "$config_backup"

# ndaunode, unlike chaosnode, needs a configuration file to work right
# in a real node, we'd need to specify parameters such as where to connect
# to the chaos chain, and so on.
# We need to support the use case of initting a real node.
# However, most of the time we run these scripts, we're just starting
# a dev server for debugging purposes. In that case, we just want a default
# config file to be put in place
ndauconf="${NDAUHOME}/ndau/config.toml"
if [ -z "$NDAUNODE_CONFIG" ]; then
    # shellcheck disable=SC2016
    errcho "$me" '$NDAUNODE_CONFIG unset'
    # shellcheck disable=SC2016
    errcho "$me" '$NDAUNODE_CONFIG must be the path to a valid ndau node config file'
    exit 1
fi
cp -v "$NDAUNODE_CONFIG" "$ndauconf"
"$ROOT"/bin/update-hash.sh
