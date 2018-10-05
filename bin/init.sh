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
    -e '/^create_empty_blocks_interval/s/[[:digit:]]+/10/' \
    -e '/^create_empty_blocks\b/{
            s/true/false/
            s/(.*)/# \1/
            i # tendermint respects create_empty_blocks *OR* create_empty_blocks_interval
        }' \
    "$config_backup" > "$config"

errcho "$me" "config diff:"
diff "$config_backup" "$config"
rm "$config_backup"

# detect if the chaos chain is currently running
# if so, we want to connect to that chain instead of using a mockfile
chaospath="$ROOT/../chaos"
if [ -d "$chaospath" ]; then
    errcho "$me" "found chaos path: $chaospath"
    cn_port=$(
        cd "$chaospath"
        bin/defaults.sh docker-compose port tendermint "$TM_RPC" 2>/dev/null |\
        cut -d: -f2
    )
    local_ip=$(ipconfig getifaddr en0)
    cn_addr=$(printf '%s:%s' "$local_ip" "$cn_port")
else
    errcho "$me" "chaosnode not found at $chaospath"
fi
if [ -n "$cn_port" ]; then
    errcho "$me" "chaosnode appears to be running at $cn_addr"
else
    errcho "$me" "chaosnode appears not to be running"
fi

# ndaunode, unlike chaosnode, needs a configuration file to work right
# in a real node, we'd need to specify parameters such as where to connect
# to the chaos chain, and so on.
# We need to support the use case of initting a real node.
# However, most of the time we run these scripts, we're just starting
# a dev server for debugging purposes. In that case, we just want a default
# config file to be put in place
ndauconf="${NDAUHOME}/ndau/config.toml"
if [ -n "$NDAUNODE_CONFIG" ]; then
    cp -v "$NDAUNODE_CONFIG" "$ndauconf"
else
    errcho "$me" "ndaunode making mocks"
    docker-compose run --rm --no-deps ndaunode --make-mocks

    if [ -n "$cn_port" ]; then
        errcho "$me" "updating config with chaos port"
        $sed -E \
            -e "/^ChaosAddress/s/\"[^\"]*\"/\"$cn_addr\"/" \
            -i "$ndauconf"
        errcho "$me" "ndaunode making chaos mocks"
        docker-compose run --rm --no-deps ndaunode --make-chaos-mocks
    fi
fi

"$ROOT"/bin/update-hash.sh
