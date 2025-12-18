# Contribution and CLI Development Guide

## Scope and purpose
These instructions apply to the entire repository. They outline expectations for contributors and give Golang CLI guidance to keep tooling consistent and maintainable.

## General contribution expectations
- Run `gofmt` and `goimports` on all touched files. Prefer standard library over external dependencies unless a clear benefit is documented in the PR description.
- Keep changes small and focused. New files should include package-level comments where it improves clarity.
- Tests are required for new behavior; add table-driven tests when practical. Use `go test ./...` before submitting.
- PR descriptions should include: a one-line summary, a short bullet list of changes, and notes on testing commands that were run.
- Avoid try/catch equivalents around imports; imports should fail fast.

## CLI-specific guidance (Golang)
Use these practices when creating or modifying CLI commands (e.g., under `cmd/`):
- **Command layout:** Organize commands under `cmd/<app>/` with `main.go` wiring and subcommands in discrete files. Prefer the standard library `flag` package or a well-maintained library (e.g., `cobra`) when subcommand trees are complex.
- **Argument handling:** Provide sensible defaults, short and long flag forms when supported, and reject conflicting or ambiguous flags with clear errors. Validate user inputs early and return non-zero exit codes on failure.
- **Configuration:** Support environment variable overrides and config files only when necessary; document precedence (flags > env > config) explicitly in help text.
- **Logging and output:** Use structured logging where possible. Send user-facing results to `stdout` and diagnostics to `stderr`. Avoid noisy output; include `--quiet` or `--verbose` where appropriate.
- **Error handling:** Wrap errors with context using `%w` and ensure messages are actionable. Do not log-and-return the same error unless the log adds value. Exit with consistent codes (e.g., `0` success, `1` usage or general failure, `2` configuration issues, `3` external/system errors).
- **Dependency boundaries:** Keep CLI wiring thin; business logic should live in reusable packages under `internal/` or `pkg/`. Subcommands should call into those packages rather than duplicating logic.
- **I/O and paths:** Accept `-` to read from stdin when applicable and use absolute paths only when required. Handle Windows and POSIX path differences when working with user-supplied file paths.
- **Concurrency:** Guard concurrent writes to shared outputs. Use contexts with timeouts for network or long-running operations, and honor `SIGINT`/`SIGTERM` for graceful shutdowns.
- **Security:** Prefer least-privilege defaults, avoid logging secrets, and validate any external input before use. When spawning processes, use explicit argument lists instead of shell invocation.
- **Testing:** Provide unit tests for command parsing and integration-style tests for end-to-end workflows. Use golden files for stable text output and keep fixtures under `testdata/`.
- **Documentation:** Keep `README` or command-specific docs up to date. Ensure `--help` output is clear and includes examples for common tasks.

## Documentation-derived expectations
The `docs/` directory contains architectural notes and testing templates; treat them as normative when working in the related areas.

- For Mermaid ER diagram generation, preserve the responsibility boundaries spelled out in `docs/mermaid_analysis.md`: keep the CLI (`cmd/marid/main.go`), renderer (`internal/diagram/generate.go`), and schema extraction (`internal/schema/extract.go`) concerns separate. Split Mermaid-specific logic from format-agnostic logic, and remember to refresh expectations in `internal/diagram/generate_test.go` when behavior changes.
- When adding or modifying a formatter, follow `docs/formatters/testing_template.md`: include rendering tests using `formattertest.SampleRenderData()`, DI behavior checks using `formattertest.MockFormatter`, and add corresponding cases to the contract tests in `pkg/formatter/formatter_contract_test.go`.

## Release hygiene
- Update version strings and changelog entries together. Keep backwards compatibility where possible; document breaking changes prominently.
- Ensure binaries build on Linux, macOS, and Windows (amd64/arm64). Run `go vet` and `staticcheck` when available to catch common issues before release.
