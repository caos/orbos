#!/bin/bash

set -e

go install golang.org/x/tools/cmd/stringer
go install github.com/awalterschulze/goderive
go generate ./...