# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`nonamedreturns` is a Go static-analysis linter that reports all named return values in function and method signatures. It is built on `golang.org/x/tools/go/analysis` and ships both as a standalone binary (via `singlechecker`) and as an importable `analyzer.Analyzer` that golangci-lint embeds.

## Workflow

- Always run `task lint` and `task test` after each change, and make sure both pass before considering the change done.
- Always add or update tests when adding or changing code.

## Commands

This project uses [Task](https://taskfile.dev) (`Taskfile.yml`), not Make.

- `task build` — runs `go fmt`, `go vet`, then builds the `nonamedreturns` binary (CGO disabled).
- `task test` — runs `go test -race -cover ./...` (CGO enabled, required for `-race`).
- `task lint` — runs `golangci-lint run ./... --timeout=30m` plus `go mod tidy`.
- `task deps` / `task update` — `go mod tidy` / upgrade all modules.
- `task tag` — tags and pushes a release; requires the `TAG` env var (goreleaser handles the actual release via GitHub Actions).

Run a single test directly with the Go toolchain, e.g. `go test -run TestAll ./analyzer/`. There is only one test entry point (`TestAll`).

## Architecture

The entire linter is in `analyzer/analyzer.go` (~140 lines); `main.go` just wires `analyzer.Analyzer` into `singlechecker.Main`.

- `Analyzer.Run` (`run`) uses the `inspect` pass to walk only `*ast.FuncDecl` and `*ast.FuncLit` nodes. For each, it iterates the results field list and reports any result name that isn't `_`.
- The `report-error-in-defer` flag (constant `FlagReportErrorInDefer`) controls one special case: by default a named return of type `error` is **not** reported if it is assigned inside a `defer` closure in the same function body. `findDeferWithVariableAssignment` / `findVariableAssignment` implement this by matching the `types.Object` of the result against assignment LHS identifiers. Set the flag to `true` to report these too.

## Tests

Tests use `analysistest` (the standard x/tools golden-file harness). Test fixtures live in `testdata/src/<config-name>/`:

- `default-config/` — expectations with the defer/error exemption on.
- `report-error-in-defer/` — expectations with the flag set to `true`.

Expected diagnostics are encoded as `// want "..."` comments inside the fixture `.go` files. To add or change behavior, edit the fixture and its `// want` comments rather than asserting in Go test code. `TestAll` runs the default config first, then flips `FlagReportErrorInDefer` and runs the second fixture set — note the flag is process-global state, so order matters within that test.
