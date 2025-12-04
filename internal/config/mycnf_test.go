package config

import (
	"os"
	"path/filepath"
	"testing"
)

func withStubUserHomeDir(t *testing.T, dir string) {
	t.Helper()

	original := userHomeDir
	userHomeDir = func() (string, error) {
		return dir, nil
	}

	t.Cleanup(func() {
		userHomeDir = original
	})
}

func TestGetMyCnfConfigMissingFile(t *testing.T) {
	tempDir := t.TempDir()
	withStubUserHomeDir(t, tempDir)

	if _, err := GetMyCnfConfig(); err == nil {
		t.Fatalf("expected error when .my.cnf is missing")
	}
}

func TestGetMyCnfConfigDefaults(t *testing.T) {
	tempDir := t.TempDir()
	withStubUserHomeDir(t, tempDir)

	myCnfPath := filepath.Join(tempDir, ".my.cnf")
	content := []byte("[client]\nuser=test-user\n")
	if err := os.WriteFile(myCnfPath, content, 0o600); err != nil {
		t.Fatalf("failed to write .my.cnf: %v", err)
	}

	cfg, err := GetMyCnfConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Fatalf("expected default host 'localhost', got %q", cfg.Host)
	}

	if cfg.Port != 3306 {
		t.Fatalf("expected default port 3306, got %d", cfg.Port)
	}

	if cfg.User != "test-user" {
		t.Fatalf("expected user 'test-user', got %q", cfg.User)
	}
}

func TestGetMyCnfConfigClientSection(t *testing.T) {
	tempDir := t.TempDir()
	withStubUserHomeDir(t, tempDir)

	myCnfPath := filepath.Join(tempDir, ".my.cnf")
	content := []byte("[client]\nhost=db.example.com\nport=3307\nuser=my-user\npassword=my-pass\ndatabase=my-db\n")
	if err := os.WriteFile(myCnfPath, content, 0o600); err != nil {
		t.Fatalf("failed to write .my.cnf: %v", err)
	}

	cfg, err := GetMyCnfConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "db.example.com" {
		t.Fatalf("expected host 'db.example.com', got %q", cfg.Host)
	}

	if cfg.Port != 3307 {
		t.Fatalf("expected port 3307, got %d", cfg.Port)
	}

	if cfg.User != "my-user" {
		t.Fatalf("expected user 'my-user', got %q", cfg.User)
	}

	if cfg.Password != "my-pass" {
		t.Fatalf("expected password 'my-pass', got %q", cfg.Password)
	}

	if cfg.Database != "my-db" {
		t.Fatalf("expected database 'my-db', got %q", cfg.Database)
	}
}
