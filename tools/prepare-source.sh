#!/bin/sh
set -eu

GO_VERSION=1.24.12 # also update .github/workflows/gobump.yml
export GOWORK=off

# Pin Go and toolchain versions at a reasonable version
go get go@$GO_VERSION toolchain@$GO_VERSION

# Generate source
go generate -x ./pkg/...

# Reformat source
go run golang.org/x/tools/cmd/goimports@latest -w ./pkg
go fmt ./pkg/...

# Update go.mod and go.sum (keep it as the last)
go mod tidy
