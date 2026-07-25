# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Marid is a Go CLI that connects to a MySQL database, extracts its schema, and renders an ER diagram (Mermaid by default). Single binary, built with `spf13/cobra`.

## Commands

```bash
# Build
go build -o marid ./cmd/marid

# Run all tests
go test ./...

# Run a single package's tests
go test ./pkg/formatter/mermaid/...

# Run a single test by name
go test ./internal/diagram/... -run TestGenerate

# Lint (matches CI's golangci-lint-action version)
golangci-lint run ./...

# Measure code complexity (matches CI's cccc job)
# requires https://github.com/moznion/cccc

# Coverage, matching the CI build.yml job
mkdir -p coverage
go test -coverpkg=./... ./... -v -coverprofile=coverage/coverage.out -timeout=5m
go tool cover -html=coverage/coverage.out -o coverage/coverage.html
./scripts/check-coverage-threshold.sh coverage/coverage.out 30   # CI fails below 30%
./scripts/coverage-summary.sh coverage/coverage.out               # prints just the % total
rm -rf coverage                                                   # clean up when done

# Manual smoke test against a real MySQL instance
marid -H localhost -P 3306 -u root -p password -d mydatabase
```

CI runs static analysis and linting *before* tests (`.github/workflows/static-analysis.yml`, then `.github/workflows/build.yml`); do the same locally: `golangci-lint run ./...` first, then `go test ./...`.

## Architecture

Data flows in one direction through four layers, wired together in `cmd/marid/main.go`:

```
cmd/marid (cobra CLI, flags)
  -> internal/config   (Config struct, ~/.my.cnf merge, password prompt)
  -> internal/database (sql.DB connection)
  -> internal/schema   (Extract: DB -> format-agnostic DatabaseSchema)
  -> internal/diagram  (Generate: DatabaseSchema -> formatter.RenderData -> rendered string)
       -> pkg/formatter          (Formatter interface, Factory, registry)
       -> pkg/formatter/mermaid  (the only built-in Formatter implementation, default)
```

Key boundary: `internal/schema` and `internal/diagram` know nothing about Mermaid syntax — `internal/diagram.toRenderData` converts `schema.DatabaseSchema` into the format-agnostic `formatter.RenderData`/`Table`/`Column`/`ForeignKey` types, and only `pkg/formatter/mermaid` turns that into `erDiagram` text. `docs/mermaid_analysis.md` documents this split in more detail if you're touching the diagram/schema boundary.

### Formatter plugin pattern

Formats are self-registering via `init()`:

```go
func init() {
    formatter.Register("plantuml", func() formatter.Formatter { return New() })
}
```

`internal/diagram/generate.go` blank-imports each formatter package (`_ "github.com/motchang/marid/pkg/formatter/mermaid"`) purely for this registration side effect. To add a formatter:
1. Create `pkg/formatter/<format>/formatter.go` implementing `formatter.Formatter` (`Name`/`MediaType`/`Render`).
2. Register it in `init()` with a unique name — `formatter.Register` panics on empty names, nil factories, or duplicates.
3. Add a matching entry to `pkg/formatter/formatter_contract_test.go`'s table (name, media type, expected output snippet) — this is the standard way new formatters get tested.
4. No CLI wiring is needed; `--format <name>` picks it up automatically via `formatter.Get`.

Full walkthrough with a PlantUML skeleton: README.md's "Formatter Development Guide" section. Testing helpers (`formattertest.SampleRenderData()`, `formattertest.MockFormatter`) are documented in `docs/formatters/testing_template.md`.

### The `internal/testsqlmock` module

`go.mod` has `replace github.com/DATA-DOG/go-sqlmock => ./internal/testsqlmock`, and `internal/testsqlmock` is its own Go module. All DB-facing tests (`internal/database`, `internal/schema`) depend on `go-sqlmock` but resolve to this local, in-repo replacement rather than the upstream module — keep that in mind when touching sqlmock behavior or debugging mock-related test failures, since editing upstream `go-sqlmock` sources elsewhere won't affect this repo.

### Golden-file testing

`internal/diagram/generate_test.go` and the formatter contract tests compare generated output against literal expected strings. There's also an end-to-end snapshot: `testdata/ddl/ecommerce.sql` is loaded into a real MySQL instance in CI (`.github/workflows/mysql-mermaid.yml`), `marid` is run against it, and the output is diffed against `testdata/expected/ecommerce.mmd`. Regenerate that snapshot with:

```bash
marid -H localhost -P 3306 -u root -p password -d ecommerce > testdata/expected/ecommerce.mmd
```

## Conventions (from AGENTS.md)

- Business logic belongs in `internal/`/`pkg/`; keep `cmd/marid` limited to flag wiring.
- Wrap errors with `%w` and make messages actionable; avoid logging an error and also returning it.
- User-facing output goes to stdout, diagnostics to stderr.
- Table-driven tests are preferred; use golden files under `testdata/` for stable text output.
- `gofmt`/`goimports` on all touched files.
