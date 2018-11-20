#! /bin/bash

# Assuming that the txmobile project has been built, this script uploads the artifacts
# to S3 under the desired version number, and allows everyone to read them.

upload_to_s3() {
    name=$1
    prefix=$2
    aws s3 cp ${name} s3://ndau-txmobile/${prefix}/${name}
    # grant read access to all users for this object
    echo "Setting read permissions on https://s3.amazonaws.com/ndau-txmobile/${prefix}/${name}"
    aws s3api put-object-acl \
        --bucket ndau-txmobile \
        --key ${prefix}/${name} \
        --grant-read uri=http://acs.amazonaws.com/groups/global/AllUsers
}

copy1() {
    name=$1
    prefix=$2
    #hash=$(git rev-parse HEAD)
    upload_to_s3 ${name} ${prefix}
}

if [ "$1" ]; then
    version=$1
    copy1 txmobile-sources.jar $version
    copy1 txmobile.aar $version
    copy1 txmobile.zip $version
else
    echo "Version is required."
    echo "Usage: upload_to_s3 vX.Y.Z"
fi
