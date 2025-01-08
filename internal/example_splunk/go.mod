module github.com/osbuild/logging/internal/example_splunk

go 1.21

replace github.com/osbuild/logging => ../..

require github.com/osbuild/logging v0.0.0-00010101000000-000000000000

require (
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
)
