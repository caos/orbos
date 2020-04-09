#!/bin/bash

set -e

rm -rf ./artifacts/*

export CGO_ENABLED=0
go run ./cmd/gen-executables/*.go --version "${IMAGE}" --commit "$(date --iso-8601=seconds)" --debug $DEBUG --orbctl ./artifacts

TARGET=prod
if [[ "$DEBUG" == "true" ]]; then
  TARGET=build
fi

docker build --tag $IMAGE --build-arg DEBUG --target $TARGET .
[[ $PUSH_IMAGE=true ]] && docker push $IMAGE
rm -rf ./artifacts/*


