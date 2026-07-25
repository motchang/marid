package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

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

func TestWrappedDBDelegatesToUnderlyingSQLDB(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	w := &wrappedDB{DB: mockDB}

	if got := w.SQLDB(); got != mockDB {
		t.Fatalf("SQLDB() returned a different *sql.DB than the one it wraps")
	}

	w.SetMaxOpenConns(7)
	if stats := mockDB.Stats(); stats.MaxOpenConnections != 7 {
		t.Fatalf("expected SetMaxOpenConns to apply to the underlying db, got MaxOpenConnections=%d", stats.MaxOpenConnections)
	}

	// SetMaxIdleConns closes any excess idle connections synchronously, so
	// opening two connections and releasing them back to the pool, then
	// lowering the idle limit below 2, gives an observable effect
	// (Stats().MaxIdleClosed) that proves the call reached the real *sql.DB.
	w.SetMaxIdleConns(5)
	ctx := context.Background()
	c1, err := mockDB.Conn(ctx)
	if err != nil {
		t.Fatalf("failed to acquire connection: %v", err)
	}
	c2, err := mockDB.Conn(ctx)
	if err != nil {
		t.Fatalf("failed to acquire connection: %v", err)
	}
	if err := c1.Close(); err != nil {
		t.Fatalf("failed to release connection: %v", err)
	}
	if err := c2.Close(); err != nil {
		t.Fatalf("failed to release connection: %v", err)
	}
	if stats := mockDB.Stats(); stats.Idle != 2 {
		t.Fatalf("expected 2 idle connections before lowering the idle limit, got %d", stats.Idle)
	}

	w.SetMaxIdleConns(1)
	if stats := mockDB.Stats(); stats.MaxIdleClosed == 0 {
		t.Fatalf("expected SetMaxIdleConns to apply to the underlying db, no excess idle connections were closed")
	}

	// SetConnMaxLifetime only takes effect the next time a connection is
	// reused, so set a lifetime that's already expired by the time the pool
	// checks it and observe Stats().MaxLifetimeClosed.
	w.SetConnMaxLifetime(time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	c3, err := mockDB.Conn(ctx)
	if err != nil {
		t.Fatalf("failed to acquire connection: %v", err)
	}
	if err := c3.Close(); err != nil {
		t.Fatalf("failed to release connection: %v", err)
	}
	if stats := mockDB.Stats(); stats.MaxLifetimeClosed == 0 {
		t.Fatalf("expected SetConnMaxLifetime to apply to the underlying db, no expired connections were closed")
	}

	if err := w.Ping(); err != nil {
		t.Fatalf("expected Ping to succeed, got %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("expected Close to succeed, got %v", err)
	}

	if pingErr := mockDB.Ping(); pingErr == nil {
		t.Fatalf("expected the underlying db to be closed after wrappedDB.Close()")
	}
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
