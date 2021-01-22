#!/bin/bash

set -e

go build -gcflags "all=-N -l" -o /tmp/db-debug ./cmd/database-debug
dlv exec --api-version 2 --headless --listen 127.0.0.1:2345 /tmp/db-debug -- "$@"
