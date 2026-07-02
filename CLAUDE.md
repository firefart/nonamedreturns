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

The entire linter is in `analyzer/analyzer.go`; `main.go` just wires `analyzer.Analyzer` into `singlechecker.Main`.

- `Analyzer.Run` (`run`) uses the `inspect` pass to walk only `*ast.FuncDecl` and `*ast.FuncLit` nodes. For each, it iterates the results field list and reports any result name that isn't `_`.
- The `report-error-in-defer` flag (constant `FlagReportErrorInDefer`) controls one special case: by default a named return of type `error` is **not** reported when it is referenced inside a `defer` closure **and** assigned somewhere in the function — explicitly (assignment or `for ... = range` statement, inside the defer or anywhere else) or implicitly via a top-level `return` with result values. `collectDeferUsageAndAssignments` implements this with a single closure-aware body walk, matching the `types.Object` of the result. Set the flag to `true` to report these too.
- The `allow-unused-named-returns` flag (constant `FlagAllowUnusedNamedReturns`) inverts the model: named returns are allowed in the signature but reported when referenced anywhere in the body or when the function contains a naked `return`. `collectNamedReturnUsage` implements this. When this flag is set it fully takes over — the error-in-defer exemption and `report-error-in-defer` have no effect.
- Both body walks treat nested closures specially: a `return` inside a closure populates the closure's own results, not the enclosing function's, but references/assignments to captured named returns still count. Shadowing is handled by `types.Object` identity.

## Tests

Tests use `analysistest` (the standard x/tools golden-file harness). Test fixtures live in `testdata/src/<config-name>/`:

- `default-config/` — expectations with the defer/error exemption on.
- `report-error-in-defer/` — expectations with `report-error-in-defer` set to `true`.
- `allow-unused-named-returns/` — expectations with `allow-unused-named-returns` set to `true`.

Expected diagnostics are encoded as `// want "..."` comments inside the fixture `.go` files. To add or change behavior, edit the fixture and its `// want` comments rather than asserting in Go test code. `TestAll` runs the fixture sets in the order above, setting and resetting flags between runs — the flags are process-global state on the Analyzer, so order matters within that test.
