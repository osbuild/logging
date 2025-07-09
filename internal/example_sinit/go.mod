module github.com/osbuild/logging/internal/example_sinit

go 1.22.1

toolchain go1.23.4

replace github.com/osbuild/logging => ../..

require github.com/osbuild/logging v0.0.0-00010101000000-000000000000

require (
	github.com/aws/aws-sdk-go-v2 v1.32.7 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.7 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.48 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.26 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.26 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.45.1 // indirect
	github.com/aws/smithy-go v1.22.1 // indirect
	github.com/getsentry/sentry-go v0.34.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.4 // indirect
	github.com/lzap/cloudwatchwriter2 v1.4.2 // indirect
	github.com/samber/lo v1.47.0 // indirect
	github.com/samber/slog-common v0.18.1 // indirect
	github.com/samber/slog-sentry/v2 v2.9.3 // indirect
	github.com/systemd/slog-journal v0.1.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)
