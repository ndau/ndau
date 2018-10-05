#!/bin/bash

ROOT="$(cd "$(dirname "$0")/.." || exit 1; pwd -P )"
me=$(basename "$0") # get tag

# shellcheck source=./common.sh
source "$ROOT"/bin/common.sh

# shellcheck source=./defaults.sh
source "$ROOT"/bin/defaults.sh

dependencies=(jq docker-compose)
for tool in "${dependencies[@]}"; do
    if ! command -v "$tool" > /dev/null  ; then
        err "$me" "This script depends on $tool. Install it and try again."
    fi
done

# configure tendermint to recognize the empty app hash
# this only needs to be run once, before genesis
genesis="$TMHOME"/config/genesis.json
genesis_backup=${genesis}.bak
# unminify so the diff is cleaner later
jq '.' "$genesis" > "$genesis_backup"

# though we haven't actually started the database yet, it's not empty:
# -make-mocks has added some mock data. We therefore want to use its
# current hash as the base empty hash
empty_hash=$(
    docker-compose run --rm --no-deps ndaunode --echo-hash --use-ndauhome 2> /dev/null |\
    tr -d '\r'
)

errcho "$me" "Empty hash: $empty_hash"
jq ".app_hash=\"$empty_hash\"" "$genesis_backup" > "$genesis"

errcho "$me" "genesis.json diff:"
diff "$genesis_backup" "$genesis"
rm -f "$genesis_backup"
