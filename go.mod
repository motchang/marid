module github.com/motchang/marid

go 1.25.0

require (
	github.com/DATA-DOG/go-sqlmock v0.0.0
	github.com/go-ini/ini v1.67.0
	github.com/go-sql-driver/mysql v1.10.0
	github.com/spf13/cobra v1.10.2
	golang.org/x/term v0.44.0
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/sys v0.46.0 // indirect
)

replace github.com/DATA-DOG/go-sqlmock => ./internal/testsqlmock
