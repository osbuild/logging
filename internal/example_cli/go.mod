module github.com/osbuild/logging/internal/example_cli

go 1.22.1

toolchain go1.23.4

replace github.com/osbuild/logging => ../..

require (
	github.com/google/pprof v0.0.0-20241210010833-40e02aabc2ad
	github.com/osbuild/logging v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.36.2
)
