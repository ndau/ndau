#!/bin/bash

ssh-keyscan github.com > githubKey 2>/dev/null
fingerprint=$(ssh-keygen -lf githubKey | cut -d' ' -f2)
echo "github.com fingerprint: $fingerprint"

# now check that the fingerprint matches a known github fingerprint
# see https://help.github.com/articles/github-s-ssh-key-fingerprints/
declare -a legit_github_fingerprints=(
    "16:27:ac:a5:76:28:2d:36:63:1b:56:4d:eb:df:a6:48"
    "ad:1c:08:a4:40:e3:6f:9c:f5:66:26:5d:4b:33:5d:8c"
    "SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8"
    "SHA256:br9IjFspm1vxR3iA35FWE+4VTyz1hYVLIE2t1/CeyWQ"
)
fp_is_legit=0
for legit_fp in "${legit_github_fingerprints[@]}"; do
    if [ "$fingerprint" == "$legit_fp" ]; then
        fp_is_legit=1
        break
    fi
done
if [ "$fp_is_legit" != "1" ]; then
    echo "FAIL: github fingerprint not valid! MITM?"
    exit 1
fi

kh="$HOME/.ssh/known_hosts"

# if we're here, we believe the fingerprint to be valid
if ! grep github.com "$kh" >/dev/null 2>&1; then
    echo "Adding github.com to $kh"
    echo "$(cat githubKey)" >> "$kh"
else
    echo "github.com already present in $kh"
fi
rm githubKey
