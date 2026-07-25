package main

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/motchang/marid/internal/config"
	"github.com/motchang/marid/internal/database"
	"github.com/motchang/marid/internal/diagram"
	"github.com/motchang/marid/internal/schema"
	"github.com/motchang/marid/pkg/formatter"
)

func resetGlobals() {
	cfgHost = "localhost"
	cfgPort = 3306
	cfgUser = "root"
	cfgPassword = ""
	cfgDatabase = ""
	cfgTables = ""
	cfgFormat = formatter.DefaultFormat
	cfgPromptPass = false
	cfgUseMyCnf = false
	cfgNoPassword = false

	getMyCnfConfig = config.GetMyCnfConfig
	promptForPassword = config.PromptForPassword
	connect = database.Connect
	extract = schema.Extract
	generate = diagram.Generate
}

func TestMissingDatabaseError(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)

	connectCalled := false
	connect = func(cfg config.Config) (*sql.DB, error) {
		connectCalled = true
		return nil, nil
	}

	cmd := buildRootCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for missing database")
	}

	if err.Error() != "database name is required" {
		t.Fatalf("unexpected error: %v", err)
	}

	if connectCalled {
		t.Fatalf("connect should not be called when database is missing")
	}
}

func TestUseMyCnfMergeSuccess(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)

	getMyCnfConfig = func() (*config.MySQLConfig, error) {
		return &config.MySQLConfig{
			Host:     "file-host",
			Port:     1234,
			User:     "file-user",
			Password: "file-pass",
			Database: "file-db",
		}, nil
	}

	var received config.Config
	connect = func(cfg config.Config) (*sql.DB, error) {
		received = cfg
		return nil, errors.New("stop connect")
	}

	cmd := buildRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--use-mycnf", "--host", "cli-host", "--tables", "foo"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "failed to connect") {
		t.Fatalf("expected connect error, got %v", err)
	}

	if received.Host != "cli-host" {
		t.Fatalf("expected host from command line, got %s", received.Host)
	}

	if received.Database != "file-db" {
		t.Fatalf("expected database from my.cnf, got %s", received.Database)
	}

	if received.Password != "file-pass" {
		t.Fatalf("expected password from my.cnf, got %s", received.Password)
	}

	if received.Tables != "foo" {
		t.Fatalf("expected tables from command line, got %s", received.Tables)
	}
}

func TestUseMyCnfMergeFailure(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)

	getMyCnfConfig = func() (*config.MySQLConfig, error) {
		return nil, errors.New("missing file")
	}

	var received config.Config
	connect = func(cfg config.Config) (*sql.DB, error) {
		received = cfg
		return nil, errors.New("stop connect")
	}

	cmd := buildRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--use-mycnf", "--database", "cli-db", "--host", "cli-host"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "failed to connect") {
		t.Fatalf("expected connect error, got %v", err)
	}

	if !strings.Contains(stderr.String(), "Warning: Could not read .my.cnf") {
		t.Fatalf("expected warning about my.cnf, got %q", stderr.String())
	}

	if received.Database != "cli-db" {
		t.Fatalf("expected database from command line, got %s", received.Database)
	}

	if received.Host != "cli-host" {
		t.Fatalf("expected host from command line, got %s", received.Host)
	}
}

func TestAskPasswordOverrides(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)

	getMyCnfConfig = func() (*config.MySQLConfig, error) {
		return &config.MySQLConfig{Password: "mycnf-pass", Database: "file-db"}, nil
	}

	promptCalled := false
	promptForPassword = func() (string, error) {
		promptCalled = true
		return "prompt-pass", nil
	}

	var received config.Config
	connect = func(cfg config.Config) (*sql.DB, error) {
		received = cfg
		return nil, errors.New("stop connect")
	}

	cmd := buildRootCmd()
	cmd.SetArgs([]string{"--use-mycnf", "--password", "cli-pass", "--database", "cli-db", "--ask-password"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "failed to connect") {
		t.Fatalf("expected connect error, got %v", err)
	}

	if !promptCalled {
		t.Fatalf("expected password prompt to be called")
	}

	if received.Password != "prompt-pass" {
		t.Fatalf("expected password from prompt, got %s", received.Password)
	}

	if received.Database != "cli-db" {
		t.Fatalf("expected database from command line, got %s", received.Database)
	}
}

func TestSuccessfulExecution(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)

	connectCalled := false
	extractCalled := false
	generateCalled := false

	connect = func(cfg config.Config) (*sql.DB, error) {
		connectCalled = true
		if cfg.Database != "cli-db" {
			t.Fatalf("unexpected database: %s", cfg.Database)
		}
		return nil, nil
	}

	extract = func(db *sql.DB, cfg config.Config) (*schema.DatabaseSchema, error) {
		extractCalled = true
		if cfg.Database != "cli-db" {
			t.Fatalf("unexpected database in extract: %s", cfg.Database)
		}
		return &schema.DatabaseSchema{Config: cfg}, nil
	}

	generate = func(dbSchema *schema.DatabaseSchema, format string) (string, error) {
		generateCalled = true
		if dbSchema.Config.Database != "cli-db" {
			t.Fatalf("unexpected database in schema: %s", dbSchema.Config.Database)
		}
		if format != formatter.DefaultFormat {
			t.Fatalf("unexpected format: %s", format)
		}
		return "diagram-output", nil
	}

	cmd := buildRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--database", "cli-db"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected successful execution, got %v", err)
	}

	if !connectCalled || !extractCalled || !generateCalled {
		t.Fatalf("expected all pipeline functions to be called")
	}

	if !strings.Contains(stdout.String(), "diagram-output") {
		t.Fatalf("expected diagram output, got %q", stdout.String())
	}
}

func TestSuccessfulExecutionWithExplicitMermaidFormat(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)

	connectCalled := false
	extractCalled := false
	generateCalled := false

	connect = func(cfg config.Config) (*sql.DB, error) {
		connectCalled = true
		if cfg.Format != "mermaid" {
			t.Fatalf("unexpected format propagated to connect: %s", cfg.Format)
		}
		return nil, nil
	}

	extract = func(db *sql.DB, cfg config.Config) (*schema.DatabaseSchema, error) {
		extractCalled = true
		if cfg.Format != "mermaid" {
			t.Fatalf("unexpected format propagated to extract: %s", cfg.Format)
		}
		return &schema.DatabaseSchema{Config: cfg}, nil
	}

	generate = func(dbSchema *schema.DatabaseSchema, format string) (string, error) {
		generateCalled = true
		if format != "mermaid" {
			t.Fatalf("unexpected format received by generate: %s", format)
		}
		return "mermaid-diagram-output", nil
	}

	cmd := buildRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--database", "cli-db", "--format", "mermaid"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected successful execution, got %v", err)
	}

	if !connectCalled || !extractCalled || !generateCalled {
		t.Fatalf("expected all pipeline functions to be called")
	}

	if !strings.Contains(stdout.String(), "mermaid-diagram-output") {
		t.Fatalf("expected mermaid diagram output, got %q", stdout.String())
	}
}

func TestNoPasswordFlagClearsPassword(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)

	var received config.Config
	connect = func(cfg config.Config) (*sql.DB, error) {
		received = cfg
		return nil, errors.New("stop connect")
	}

	cmd := buildRootCmd()
	cmd.SetArgs([]string{"--database", "cli-db", "--password", "cli-pass", "--no-password"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "failed to connect") {
		t.Fatalf("expected connect error, got %v", err)
	}

	if received.Password != "" {
		t.Fatalf("expected password to be cleared by --no-password, got %q", received.Password)
	}
}

func TestSuccessfulExecutionClosesNonNilDB(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)

	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	connect = func(cfg config.Config) (*sql.DB, error) {
		return mockDB, nil
	}

	extract = func(db *sql.DB, cfg config.Config) (*schema.DatabaseSchema, error) {
		if db != mockDB {
			t.Fatalf("expected extract to receive the connected db")
		}
		return &schema.DatabaseSchema{Config: cfg}, nil
	}

	generate = func(dbSchema *schema.DatabaseSchema, format string) (string, error) {
		return "diagram-output", nil
	}

	cmd := buildRootCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--database", "cli-db"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected successful execution, got %v", err)
	}

	if pingErr := mockDB.Ping(); pingErr == nil || !strings.Contains(pingErr.Error(), "database is closed") {
		t.Fatalf("expected db to be closed by the deferred Close call, ping err: %v", pingErr)
	}
}

func TestAskPasswordPromptError(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)

	promptErr := errors.New("tty unavailable")
	promptForPassword = func() (string, error) {
		return "", promptErr
	}

	connectCalled := false
	connect = func(cfg config.Config) (*sql.DB, error) {
		connectCalled = true
		return nil, nil
	}

	cmd := buildRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--database", "cli-db", "--ask-password"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error when the password prompt fails")
	}

	if !errors.Is(err, promptErr) {
		t.Errorf("expected the prompt error to be wrapped, got %v", err)
	}

	if !strings.Contains(err.Error(), "failed to read password") {
		t.Errorf("expected an actionable message, got %v", err)
	}

	if connectCalled {
		t.Errorf("connect should not be called when the password prompt fails")
	}
}

// TestNoPasswordTakesPrecedenceOverAskPassword pins the flag precedence in
// resolveConfig: --no-password short-circuits the prompt entirely rather than
// prompting and then discarding the result.
func TestNoPasswordTakesPrecedenceOverAskPassword(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)

	promptCalled := false
	promptForPassword = func() (string, error) {
		promptCalled = true
		return "prompt-pass", nil
	}

	var received config.Config
	connect = func(cfg config.Config) (*sql.DB, error) {
		received = cfg
		return nil, errors.New("stop connect")
	}

	cmd := buildRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--database", "cli-db", "--password", "cli-pass", "--no-password", "--ask-password"})

	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "failed to connect") {
		t.Fatalf("expected connect error, got %v", err)
	}

	if promptCalled {
		t.Errorf("expected --no-password to skip the password prompt")
	}

	if received.Password != "" {
		t.Errorf("expected password to be cleared, got %q", received.Password)
	}
}

func TestExtractError(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)

	extractErr := errors.New("information_schema unreadable")

	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	connect = func(cfg config.Config) (*sql.DB, error) {
		return mockDB, nil
	}

	extract = func(db *sql.DB, cfg config.Config) (*schema.DatabaseSchema, error) {
		return nil, extractErr
	}

	generateCalled := false
	generate = func(dbSchema *schema.DatabaseSchema, format string) (string, error) {
		generateCalled = true
		return "", nil
	}

	cmd := buildRootCmd()
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--database", "cli-db"})

	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected error when schema extraction fails")
	}

	if !errors.Is(err, extractErr) {
		t.Errorf("expected the extract error to be wrapped, got %v", err)
	}

	if !strings.Contains(err.Error(), "failed to extract schema") {
		t.Errorf("expected an actionable message, got %v", err)
	}

	if generateCalled {
		t.Errorf("generate should not be called after extraction fails")
	}

	if pingErr := mockDB.Ping(); pingErr == nil || !strings.Contains(pingErr.Error(), "database is closed") {
		t.Errorf("expected db to be closed on the error path, ping err: %v", pingErr)
	}
}

// failingWriter stands in for a stdout that cannot be written to, e.g. a closed
// pipe when marid's output is piped into a command that exits early.
type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestOutputWriteErrorIsReturned(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)

	writeErr := errors.New("broken pipe")

	connect = func(cfg config.Config) (*sql.DB, error) {
		return nil, nil
	}

	extract = func(db *sql.DB, cfg config.Config) (*schema.DatabaseSchema, error) {
		return &schema.DatabaseSchema{Config: cfg}, nil
	}

	generate = func(dbSchema *schema.DatabaseSchema, format string) (string, error) {
		return "diagram-output", nil
	}

	cmd := buildRootCmd()
	var stderr bytes.Buffer
	cmd.SetOut(failingWriter{err: writeErr})
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--database", "cli-db"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected the diagram write error to be reported")
	}

	if !errors.Is(err, writeErr) {
		t.Errorf("expected the write error to be returned, got %v", err)
	}
}

func TestUnknownFormatError(t *testing.T) {
	resetGlobals()
	t.Cleanup(resetGlobals)

	connect = func(cfg config.Config) (*sql.DB, error) {
		if cfg.Format != "unknown" {
			t.Fatalf("expected format to be forwarded, got %q", cfg.Format)
		}
		return nil, nil
	}

	extract = func(db *sql.DB, cfg config.Config) (*schema.DatabaseSchema, error) {
		return &schema.DatabaseSchema{Config: cfg, Tables: []schema.Table{{Name: "users"}}}, nil
	}

	generate = diagram.Generate

	cmd := buildRootCmd()
	cmd.SetArgs([]string{"--database", "cli-db", "--format", "unknown"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for unknown format")
	}

	const want = "failed to generate diagram: unknown format \"unknown\". Available formats: mermaid"
	if err.Error() != want {
		t.Fatalf("unexpected error: %v", err)
	}
}
