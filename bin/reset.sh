#!/bin/bash

ROOT="$(cd "$(dirname "$0")/.." || exit 1; pwd -P )"

# shellcheck source=./common.sh
source "$ROOT"/bin/common.sh

"$ROOT"/bin/stop.sh

# shellcheck source=./defaults.sh
source "$ROOT"/bin/defaults.sh

# completely reset tendermint
tendermint unsafe_reset_all

# remove tendermint home
rm -rfv $TMHOME

# remove ndau home
rm -rfv $NDAUHOME/ndau
