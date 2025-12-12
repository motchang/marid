# Marid: MySQL to Mermaid ER Diagram Generator

Marid is a command-line tool that connects to a MySQL database, extracts table definitions, and generates Mermaid ER diagrams. The tool is written in Go and distributed as a single binary.

[![Go Report Card](https://goreportcard.com/badge/github.com/motchang/marid)](https://goreportcard.com/report/github.com/motchang/marid)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

## Features

- Connect to MySQL servers using command-line parameters
- Extract table structure and relationships from database schema
- Generate correct Mermaid ER diagram syntax
- Output the diagram text to stdout or to a file
- Filter tables by name

## Installation

### Using Go Install

If you have Go installed, you can install Marid directly:

```bash
go install github.com/motchang/marid/cmd/marid@latest
```

### Binary Downloads

Pre-built binaries for Linux, macOS, and Windows are available on the [releases page](https://github.com/motchang/marid/releases).

## Usage

Basic usage:

```bash
marid -H localhost -P 3306 -u root -p password -d mydatabase
```

### Options

```
Options:
  -H, --host string       MySQL host address (default "localhost")
  -P, --port int          MySQL port (default 3306)
  -u, --user string       MySQL username (default "root")
  -p, --password string   MySQL password (insecure, prefer --ask-password)
  --ask-password          Prompt for password (secure)
  -c, --use-mycnf         Read connection info from ~/.my.cnf
  -n, --no-password       Connect without a password
  -d, --database string   Database name (required)
  -t, --tables string     Comma-separated list of tables (default: all tables)
  -f, --format string     Output format (default: mermaid; available: mermaid)
  -h, --help              Display help information

Note: `-h` is reserved for help output; use `-H` for the host shorthand.
```

### Output formats

- Mermaid is the default formatter.
- Use `--format` (or `-f`) to choose another registered formatter.
- When an unknown format is provided, Marid returns an error listing the available formatters so you can pick a supported one.

### Example

Connect to a local MySQL database and generate an ER diagram for specific tables:

```bash
marid -u dbuser -p dbpass -d myapp -t users,orders,products
```

Output:

```
erDiagram
    users {
        PK id integer NOT NULL
        username string NOT NULL
        email string NOT NULL
        created_at datetime
    }
    orders {
        PK id integer NOT NULL
        FK user_id integer NOT NULL
        total_amount float NOT NULL
        status string
        created_at datetime
    }
    products {
        PK id integer NOT NULL
        name string NOT NULL
        price float NOT NULL
        stock integer
        category string
    }
    users ||--o{ orders : "FK_user_orders"
```

You can pipe the output to a file:

```bash
marid -d myapp > myapp_diagram.mmd
```

## Rendering the Diagram

You can render the Mermaid diagram using:

1. The [Mermaid Live Editor](https://mermaid.live/)
2. VS Code with the Mermaid extension
3. GitLab or GitHub Markdown (both support Mermaid syntax)
4. Any Mermaid-compatible tool

## Building from Source

Clone the repository:

```bash
git clone https://github.com/yourusername/marid.git
cd marid
```

Build:

```bash
go build -o marid ./cmd/marid
```

Run tests:

```bash
go test ./...
```

## Formatter Development Guide

### Package layout and responsibilities

- `cmd/marid/`: CLI entrypoint. Responsible only for reading configuration and booting the application.
- `internal/`: Non-public implementations such as database connections, schema retrieval, and domain logic.
- `pkg/formatter/`: Public API layer that collects rendering implementations for each output format (Mermaid, PlantUML, etc.). Add new formatters here and consume them from `internal` through the interface to keep responsibilities separated and minimize the blast radius of new formats.

### How to add a formatter

1. Create a dedicated package such as `pkg/formatter/<format>/formatter.go` and implement the `formatter.Formatter` interface (`Name`/`MediaType`/`Render`).
2. Call `formatter.Register("<format>", func() formatter.Formatter { return New() })` from an `init` function to register your factory.
3. No new CLI option is required—pass the registered name to `--format` to enable your formatter.

### Registering with the formatter registry

- `formatter.Register` accepts a unique format name and a `formatter.Factory`. Avoid empty names, duplicates, or `nil` factories; they will panic.
- The name is the user-facing identifier passed via `--format`. Prefer a short slug that does not conflict with the existing `mermaid` entry.
- The registry is built automatically at startup; `formatter.Get` returns `mermaid` (the default) when no name is provided.

### How to add tests

- The quickest path is to add a single expectation to the existing contract tests. In `pkg/formatter/formatter_contract_test.go`, extend the table with the new constructor (the equivalent of `formatter.New()`) plus the expected name, media type, and a snippet of representative output.
- For formatter-specific behaviors, use the mocks in `pkg/formatter/formattertest` or the sample data described in [`docs/formatters/testing_template.md`](docs/formatters/testing_template.md).
- Run `go test ./...` to guard against regressions while verifying rendering and registry integration for your new formatter.

### PlantUML skeleton

Here is the minimal structure for a PlantUML formatter. Create `pkg/formatter/plantuml/formatter.go` and adapt the rendering logic as needed.

```go
package plantuml

import "github.com/motchang/marid/pkg/formatter"

func init() {
    formatter.Register("plantuml", func() formatter.Formatter { return New() })
}

type Formatter struct{}

func New() Formatter { return Formatter{} }

func (Formatter) Name() string      { return "plantuml" }
func (Formatter) MediaType() string { return "text/plain" }

func (Formatter) Render(data formatter.RenderData) (string, error) {
    // Iterate over data.Tables to generate a PlantUML string with @startuml / @enduml markers
    return "@startuml\n...\n@enduml", nil
}
```

For richer sample data and mock usage, see [`docs/formatters/testing_template.md`](docs/formatters/testing_template.md).

## Coverage

- Generate local coverage with `go test -coverpkg=./... ./... -v -coverprofile=coverage/coverage.out -timeout=5m` and create an HTML report with `go tool cover -html=coverage/coverage.out -o coverage/coverage.html`.
- CI uploads the `coverage/` directory (including `coverage/coverage.out` and `coverage/coverage.html`) as an artifact and comments on PRs with the total coverage and a download link for the HTML report.
- Use `rm -rf coverage` to clean up generated coverage files when finished.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.