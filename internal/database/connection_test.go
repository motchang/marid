package database

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/motchang/marid/internal/config"
)

type stubDB struct {
	pingErr     error
	closeCalled bool
	maxOpen     int
	maxIdle     int
	maxLifetime time.Duration
	pingCalled  bool
}

func (s *stubDB) Ping() error {
	s.pingCalled = true
	return s.pingErr
}

func (s *stubDB) Close() error {
	s.closeCalled = true
	return nil
}

func (s *stubDB) SetMaxOpenConns(n int) {
	s.maxOpen = n
}

func (s *stubDB) SetMaxIdleConns(n int) {
	s.maxIdle = n
}

func (s *stubDB) SetConnMaxLifetime(d time.Duration) {
	s.maxLifetime = d
}

func (s *stubDB) SQLDB() *sql.DB {
	return &sql.DB{}
}

func TestConnectConfiguresDatabase(t *testing.T) {
	t.Cleanup(func() { openDB = defaultOpenDB })

	stub := &stubDB{}
	var gotDriver, gotDSN string

	openDB = func(driverName, dataSourceName string) (dbHandle, error) {
		gotDriver = driverName
		gotDSN = dataSourceName
		return stub, nil
	}

	cfg := config.Config{
		Host:     "localhost",
		Port:     3306,
		User:     "user",
		Password: "pass",
		Database: "db",
	}

	db, err := Connect(cfg)
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}

	if db == nil {
		t.Fatalf("Connect returned nil database")
	}

	expectedDSN := "user:pass@tcp(localhost:3306)/db?parseTime=true&timeout=30s"
	if gotDSN != expectedDSN {
		t.Fatalf("DSN mismatch. got %q, want %q", gotDSN, expectedDSN)
	}

	if gotDriver != "mysql" {
		t.Fatalf("driver mismatch. got %q, want %q", gotDriver, "mysql")
	}

	if !stub.pingCalled {
		t.Fatalf("Ping was not called")
	}

	if stub.maxOpen != 10 {
		t.Errorf("SetMaxOpenConns not applied. got %d", stub.maxOpen)
	}

	if stub.maxIdle != 5 {
		t.Errorf("SetMaxIdleConns not applied. got %d", stub.maxIdle)
	}

	if stub.maxLifetime != 3*time.Minute {
		t.Errorf("SetConnMaxLifetime not applied. got %s", stub.maxLifetime)
	}

	if stub.closeCalled {
		t.Errorf("Close should not be called on successful connection")
	}
}

func TestConnectReturnsOpenError(t *testing.T) {
	t.Cleanup(func() { openDB = defaultOpenDB })

	openErr := errors.New("open failure")
	openDB = func(string, string) (dbHandle, error) {
		return nil, openErr
	}

	_, err := Connect(config.Config{})
	if err == nil || !errors.Is(err, openErr) {
		t.Fatalf("expected wrapped open error, got %v", err)
	}
}

func TestConnectReturnsPingErrorAndCloses(t *testing.T) {
	t.Cleanup(func() { openDB = defaultOpenDB })

	pingErr := errors.New("ping failure")
	stub := &stubDB{pingErr: pingErr}

	openDB = func(string, string) (dbHandle, error) {
		return stub, nil
	}

	_, err := Connect(config.Config{})
	if err == nil || !errors.Is(err, pingErr) {
		t.Fatalf("expected wrapped ping error, got %v", err)
	}

	if !stub.closeCalled {
		t.Fatalf("expected Close to be called after ping failure")
	}
}
