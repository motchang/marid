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
marid -h localhost -P 3306 -u root -p password -d mydatabase
```

### Options

```
Options:
  -h, --host string       MySQL host address (default "localhost")
  -P, --port int          MySQL port (default 3306)
  -u, --user string       MySQL username (default "root")
  -p, --password string   MySQL password (INSECURE: visible in process list)
  --ask-password          Prompt for password (secure, recommended)
  -c, --use-mycnf         Read connection info from ~/.my.cnf
  -n, --no-password       Connect without a password
  -d, --database string   Database name (required)
  -t, --tables string     Comma-separated list of tables (default: all tables)
  --help                  Display help information
```

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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.