#!/bin/bash

set -e

CGO_ENABLED=0 GOOS=linux go build -o ./artifacts/gen-charts cmd/gen-charts/*.go
go build -gcflags "all=-N -l" -o /tmp/boom-debug ./cmd/boom-debug
dlv exec --api-version 2 --headless --listen 127.0.0.1:2345 /tmp/boom-debug -- "$@"
