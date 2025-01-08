module github.com/osbuild/logging/internal/example_journal

go 1.21

replace github.com/osbuild/logging => ../..

require github.com/osbuild/logging v0.0.0-00010101000000-000000000000

require github.com/coreos/go-systemd/v22 v22.5.0 // indirect
