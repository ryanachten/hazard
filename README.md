# hazard

Event-driven hazard simulation system.

## Prerequisites

- [Go](https://go.dev/dl/) 1.26.4
- [golangci-lint](https://golangci-lint.run/) v2.12.2 (for linting)
- [pre-commit](https://pre-commit.com/) (for git hooks)

## Setup

```sh
pre-commit install
pre-commit install --hook-type pre-push
```

## Running locally
- Run Go application: `DEBUG=1 go run .`
    - Set `DEBUG=1` environment variable for log output

## Tests

- Get code coverage for path
```bash
go test -coverprofile=cov.out ./internal/pathfinding/ && go tool cover -func=cov.out
```