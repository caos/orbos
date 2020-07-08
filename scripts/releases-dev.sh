#!/bin/bash

set -e

VERSION="$(git rev-parse --abbrev-ref HEAD | sed -e "s/heads\///")"
IMAGE="docker.pkg.github.com/caos/orbos/orbos:${VERSION}"
go run -race ./cmd/gen-executables/*.go --version "$VERSION" --commit $(git rev-parse HEAD) --orbctl ./artifacts 1>&2
CGO_ENABLED=0 GOOS=linux go build -o ./artifacts/gen-charts  cmd/gen-charts/*.go
docker build --tag "$IMAGE" --file ./Dockerfile .
docker push "$IMAGE"