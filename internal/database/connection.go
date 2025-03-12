package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/motchang/marid/internal/config"
)

// Connect establishes a connection to the MySQL database
func Connect(cfg config.Config) (*sql.DB, error) {
	// Create DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&timeout=30s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	// Open connection
	db, err := sql.Open("mysql", dsn)
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

	return db, nil
}
