# Contributing to Finggu Performance Insight Tool

Thanks for considering a contribution! This project follows a simple convention
so the codebase stays easy to navigate as it grows:

- Exported types/functions: `Finggu` prefix (e.g. `FingguCollector`, `FingguDetectBottleneck`)
- Unexported helper functions: `fingguFn_` prefix (e.g. `fingguFn_GetEnv`)
- Constants: `FINGGU_` prefix

## Getting started

```bash
git clone https://github.com/sudarshanpjadhav/finggu-performance-insight-tool.git
cd finggu-performance-insight-tool
cp .env.example .env
docker compose up
```

## Running tests

```bash
go mod tidy
go test ./... -v -cover
```

## Adding a new alert channel

Implement the `FingguAlertChannel` interface in `alerts.go`:

```go
type FingguAlertChannel interface {
    Send(message string) error
    Name() string
}
```

Then register it in `FingguNewAlertManager` behind a config flag in `config.go`.

## Pull requests

- Keep PRs focused on a single change.
- Run `gofmt -w .` before committing.
- Add or update tests for any behavior change.
- Describe *why* the change is needed, not just what it does.

## Reporting bugs / suggesting features

Please open a GitHub issue with steps to reproduce (for bugs) or the problem
you're trying to solve (for features).
