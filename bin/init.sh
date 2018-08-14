#!/bin/bash

# this script depends on jq, sed, and docker-compose
# let's make sure that each is installed before going anywhere with this
dependencies=(jq sed docker-compose)
for tool in "${dependencies[@]}"; do
    if ! command -v "$tool" > /dev/null  ; then
        (>&2 echo "This script depends on $tool. Install it and try again.")
        exit 1
    fi
done
# gnu sed is required
sed="sed"
if ! sed --version > /dev/null 2>&1 ; then
    if ! command -v gsed >/dev/null; then
        (
            >&2 echo "You have a broken version of sed, and gsed is not installed"
            >&2 echo 'This is common on OSX. Try "brew install gnu-sed"'
            exit 1
        )
    fi
    sed="gsed"
    echo "using $sed as sed"
fi

SCRIPTPATH="$(cd "$(dirname "$0")" && pwd -P )"
source "$SCRIPTPATH/defaults.sh"

# get tendermint initialized
docker-compose run --rm --no-deps tendermint init

# we need tendermint to look for its application at a machine named ndaunode
config=$TMHOME/config/config.toml
config_backup=${config}.bak
cp "$config" "$config_backup"
$sed -E \
    -e '/^proxy_app/s|://[^:]*:|://ndaunode:|' \
    -e '/^create_empty_blocks_interval/s/[[:digit:]]+/10/' \
    -e '/^create_empty_blocks\b/{
            s/true/false/
            s/(.*)/# \1/
            i # tendermint respects create_empty_blocks *OR* create_empty_blocks_interval
        }' \
    "$config_backup" > "$config"
echo "diff config:"
diff "$config_backup" "$config"
rm "$config_backup"

# ndaunode, unlike chaosnode, needs a configuration file to work right
# in a real node, we'd need to specify parameters such as where to connect
# to the chaos chain, and so on.
# We need to support the use case of initting a real node.
# However, most of the time we run these scripts, we're just starting
# a dev server for debugging purposes. In that case, we just want a default
# config file to be put in place in order to
if [ -n "$NDAUNODE_CONFIG" ]; then
    cp -v "$NDAUNODE_CONFIG" "${NDAUHOME}/ndau/config.toml"
else
    echo "INFO: mocking config"
    docker-compose run --rm --no-deps ndaunode --make-mocks
fi

# configure tendermint to recognize the empty app hash
# this only needs to be run once, before genesis
genesis=$TMHOME/config/genesis.json
genesis_backup=${genesis}.bak
# unminify so the diff is cleaner later
jq '.' "$genesis" > "$genesis_backup"

# though we haven't actually started the database yet, it's not empty:
# -make-mocks has added some mock data. We therefore want to use its
# current hash as the base empty hash
empty_hash=$(
    docker-compose run --rm --no-deps ndaunode --echo-hash --use-ndauhome 2>/dev/null |\
    tr -d '\r'
)
echo "Empty hash: $empty_hash"
jq ".app_hash=\"$empty_hash\"" "$genesis_backup" > "$genesis"

echo "diff genesis:"
diff "$genesis_backup" "$genesis"
rm -f "$genesis_backup"
