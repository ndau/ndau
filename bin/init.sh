#!/bin/bash

# this script depends on jq, sed, and docker-compose
# let's make sure that each is installed before going anywhere with this
dependencies=(jq sed docker-compose)
for tool in ${dependencies[@]}; do
    which $tool > /dev/null
    if [ $? != 0 ]; then
        (>&2 echo "This script depends on $tool. Install it and try again.")
        exit 1
    fi
done
# gnu sed is required
sed=sed
sed --version > /dev/null 2>&1
if [ $? != 0 ]; then
    which gsed >/dev/null
    if [ $? != 0 ]; then
        (
            >&2 echo "You have a broken version of sed, and gsed is not installed"
            >&2 echo "This is common on OSX. Try `brew install gnu-sed`"
            exit 1
        )
    fi
    sed=gsed
    echo "using $sed as sed"
fi

SCRIPTPATH="$(cd "$(dirname "$0")" ; pwd -P )"
source $SCRIPTPATH/defaults.sh

# get tendermint initialized
docker-compose run --rm --no-deps tendermint init

# we need tendermint to look for its application at a machine named ndaunode
config=$TMHOME/config/config.toml
config_backup=${config}.bak
cp $config $config_backup
$sed -E \
    -e '/^proxy_app/s|://[^:]*:|://ndaunode:|' \
    -e '/^create_empty_blocks_interval/s/[[:digit:]]+/10/' \
    -e '/^create_empty_blocks\b/{
            s/true/false/
            s/(.*)/# \1/
            i # tendermint respects create_empty_blocks *OR* create_empty_blocks_interval
        }' \
    $config_backup > $config
echo "diff config:"
diff $config_backup $config
rm $config_backup

# configure tendermint to recognize the empty app hash
# this only needs to be run once, before genesis
genesis=$TMHOME/config/genesis.json
genesis_backup=${genesis}.bak
# unminify so the diff is cleaner later
jq '.' $genesis > $genesis_backup

# we've built the capability of simply writing the hexadecimal encoding
# of the hash of its empty value to stdout into ndaunode.
# We use that to populate tendermint's expectations for the genesis hash.
empty_hash=$(
    docker-compose run --rm --no-deps ndaunode --echo-empty-hash |\
    tr -d '\r'
)
echo "Empty hash: $empty_hash"
jq ".app_hash=\"$empty_hash\"" $genesis_backup > $genesis

echo "diff genesis:"
diff $genesis_backup $genesis
rm -f $genesis_backup
