package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/motchang/marid/internal/config"
)

type sqlDB interface {
	Ping() error
	Close() error
	SetMaxOpenConns(int)
	SetMaxIdleConns(int)
	SetConnMaxLifetime(time.Duration)
}

type dbHandle interface {
	sqlDB
	SQLDB() *sql.DB
}

type wrappedDB struct {
	DB *sql.DB
}

func (w *wrappedDB) Ping() error {
	return w.DB.Ping()
}

func (w *wrappedDB) Close() error {
	return w.DB.Close()
}

func (w *wrappedDB) SetMaxOpenConns(n int) {
	w.DB.SetMaxOpenConns(n)
}

func (w *wrappedDB) SetMaxIdleConns(n int) {
	w.DB.SetMaxIdleConns(n)
}

func (w *wrappedDB) SetConnMaxLifetime(d time.Duration) {
	w.DB.SetConnMaxLifetime(d)
}

func (w *wrappedDB) SQLDB() *sql.DB {
	return w.DB
}

var openDB = func(driverName, dataSourceName string) (dbHandle, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return &wrappedDB{DB: db}, nil
}

var defaultOpenDB = openDB

// Connect establishes a connection to the MySQL database
func Connect(cfg config.Config) (*sql.DB, error) {
	// Create DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&timeout=30s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	// Open connection
	db, err := openDB("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 3)

	// Test connection
	err = db.Ping()
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return db.SQLDB(), nil
}
