module github.com/lzap/strc/internal/example_export

go 1.23.3

replace github.com/lzap/strc => ../..

require (
	github.com/lzap/strc v0.0.0-00010101000000-000000000000
	github.com/samber/slog-multi v1.2.4
)

require (
	github.com/samber/lo v1.47.0 // indirect
	golang.org/x/text v0.18.0 // indirect
)
