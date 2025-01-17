#!/bin/bash
set -e

root_dir=$(git rev-parse --show-toplevel)

reference() {
    nr=$1
    tag=$2

    echo ".repositories[$nr] |= . * { \"reference\": \"$tag-repository.yaml\" }"
}

YQ=${YQ:-docker run --rm -v "${PWD}":/workdir mikefarah/yq}
set -x

last_commit_snapshot() {
    echo $(docker run --rm quay.io/skopeo/stable list-tags docker://$1 | jq -rc '.Tags | map(select( (. | contains("-repository.yaml")) )) | sort_by(. | sub("v";"") | sub("-repository.yaml";"") | sub("-";"") | split(".") | map(tonumber) ) | .[-1]' | sed "s/-repository.yaml//g")
}

latest_tag=$(last_commit_snapshot quay.io/c3os/packages)
latest_tag_arm64=$(last_commit_snapshot quay.io/c3os/packages-arm64)

$YQ eval "$(reference 0 $latest_tag)" -i repository.yaml
$YQ eval "$(reference 1 $latest_tag_arm64)" -i repository.yaml


