#!/bin/bash

set -e

rm -rf ./artifacts
CGO_ENABLED=0 GOOS=linux go build -o ./artifacts/gen-charts cmd/gen-charts/*.go
./artifacts/gen-charts
cp -r dashboards ./artifacts/dashboards
