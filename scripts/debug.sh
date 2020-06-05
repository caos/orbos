#!/bin/bash

set -e

go run ./cmd/gen-executables/*.go --debug --version $(git rev-parse --abbrev-ref HEAD | sed -e "s/heads\///") --commit $(git rev-parse HEAD) --orbctl ./artifacts 
CGO_ENABLED=0 GOOS=linux go build -o ./artifacts/gen-charts  cmd/gen-charts/*.go
dlv exec --api-version 2 --headless --listen 127.0.0.1:5000 ./artifacts/orbctl-Linux-x86_64 -- $@
