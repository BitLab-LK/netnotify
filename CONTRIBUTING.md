# Contributing

Run `go test ./...`, `go vet ./...`, `gofmt -w` and shellcheck for scripts before opening a pull request. Providers must implement `pkg/provider.Notifier`; sources must implement `pkg/source.Parser`.
