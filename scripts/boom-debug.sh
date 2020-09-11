#!/bin/bash

set -e

rm -rf ./artifacts
CGO_ENABLED=0 GOOS=linux go build -o ./artifacts/gen-charts cmd/gen-charts/*.go
./artifacts/gen-charts
cp -r dashboards ./artifacts/dashboards
go build -gcflags "all=-N -l" -o /tmp/boom-debug ./cmd/boom-debug
dlv exec --api-version 2 --headless --listen 127.0.0.1:2345 /tmp/boom-debug -- "$@"
