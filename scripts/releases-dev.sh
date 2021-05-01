#!/bin/bash

set -e

VERSION="$(git rev-parse --abbrev-ref HEAD | sed -e "s/heads\///" | rev | cut -d "/" -f 1 | rev)"
IMAGE="ghcr.io/caos/orbos:${VERSION}"
go run -race ./cmd/gen-executables/*.go --version "$VERSION" --commit $(git rev-parse HEAD) --orbctl ./artifacts 1>&2
CGO_ENABLED=0 GOOS=linux go build -o ./artifacts/gen-charts  cmd/gen-charts/*.go
docker build --tag "$IMAGE" --file ./build/orbos/Dockerfile .
docker push "$IMAGE"