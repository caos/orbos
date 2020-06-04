#!/bin/bash

set -e

rm -rf ./artifacts/*

export CGO_ENABLED=0
go run ./cmd/gen-executables/*.go --version "${IMAGE}" --commit "$(git log --pretty=format:'%h' -n 1)" "--debug=$DEBUG" --orbctl ./artifacts

TARGET=prod
if [[ "$DEBUG" == "true" ]]; then
  TARGET=build
fi

docker build --tag $IMAGE --target $TARGET .
[[ "$PUSH_IMAGE" == "true" ]] && docker push $IMAGE
rm -rf ./artifacts/*


