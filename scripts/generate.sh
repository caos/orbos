#!/bin/bash

set -e

go install golang.org/x/tools/cmd/stringer
go install ./cmd/gen-kindstubs/gen-kindstubs.go
go install github.com/awalterschulze/goderive
go generate ./...