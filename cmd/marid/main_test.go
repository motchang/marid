package main

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"
	"testing"

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
