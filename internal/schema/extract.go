package schema

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/motchang/marid/internal/config"
)

// Column represents a database column
type Column struct {
	Name       string
	DataType   string
	IsNullable bool
	IsPrimary  bool
	IsUnique   bool
	Comment    string
}

// ForeignKey represents a foreign key relationship
type ForeignKey struct {
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
	RelationName     string
}

// Table represents a database table
// Table represents a database table
type Table struct {
	Name        string
	Comment     string // Add this field
	Columns     []Column
	PrimaryKey  []string
	ForeignKeys []ForeignKey
}

// DatabaseSchema represents the complete database schema
type DatabaseSchema struct {
	Tables []Table
	Config config.Config
}

// Extract extracts the database schema
func Extract(db *sql.DB, cfg config.Config) (*DatabaseSchema, error) {
	schema := &DatabaseSchema{
		Tables: []Table{},
		Config: cfg,
	}

	// Get list of tables
	tables, err := getTables(db, cfg)
	if err != nil {
		return nil, err
	}

	// Extract table details
	for _, tableName := range tables {
		table, err := extractTableInfo(db, tableName)
		if err != nil {
			return nil, err
		}
		schema.Tables = append(schema.Tables, *table)
	}

	return schema, nil
}

// getTables gets the list of tables from the database
func getTables(db *sql.DB, cfg config.Config) ([]string, error) {
	filterTables := cfg.GetTablesList()
	hasTableFilter := len(filterTables) > 0

	// Query to get all tables in the database
	query := `
		SELECT TABLE_NAME
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = ?
		AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`

	rows, err := db.Query(query, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("error querying tables: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("error scanning table name: %w", err)
		}

		// Apply table filter if specified
		if hasTableFilter {
			for _, filterTable := range filterTables {
				if tableName == filterTable {
					tables = append(tables, tableName)
					break
				}
			}
		} else {
			tables = append(tables, tableName)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table rows: %w", err)
	}

	return tables, nil
}

// extractTableInfo extracts detailed information about a table
// extractTableInfo extracts detailed information about a table
func extractTableInfo(db *sql.DB, tableName string) (*Table, error) {
	table := &Table{
		Name:        tableName,
		Columns:     []Column{},
		PrimaryKey:  []string{},
		ForeignKeys: []ForeignKey{},
	}

	// Get table comment
	commentQuery := `
		SELECT 
			TABLE_COMMENT
		FROM 
			INFORMATION_SCHEMA.TABLES
		WHERE 
			TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
	`

	err := db.QueryRow(commentQuery, tableName).Scan(&table.Comment)
	if err != nil {
		return nil, fmt.Errorf("error querying table comment for %s: %w", tableName, err)
	}

	// Get column information
	if err := extractColumns(db, table); err != nil {
		return nil, err
	}

	// Get primary key information
	if err := extractPrimaryKeys(db, table); err != nil {
		return nil, err
	}

	// Get foreign key information
	if err := extractForeignKeys(db, table); err != nil {
		return nil, err
	}

	return table, nil
}

// extractColumns extracts column information for the table
// extractColumns extracts column information for the table
func extractColumns(db *sql.DB, table *Table) error {
	query := `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_KEY,
			COLUMN_COMMENT
		FROM 
			INFORMATION_SCHEMA.COLUMNS
		WHERE 
			TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
		ORDER BY 
			ORDINAL_POSITION
	`

	rows, err := db.Query(query, table.Name)
	if err != nil {
		return fmt.Errorf("error querying columns for table %s: %w", table.Name, err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var column Column
		var isNullable, columnKey string

		if err := rows.Scan(
			&column.Name,
			&column.DataType,
			&isNullable,
			&columnKey,
			&column.Comment,
		); err != nil {
			return fmt.Errorf("error scanning column: %w", err)
		}

		column.IsNullable = strings.ToUpper(isNullable) == "YES"
		column.IsPrimary = strings.ToUpper(columnKey) == "PRI"
		column.IsUnique = strings.ToUpper(columnKey) == "UNI"

		table.Columns = append(table.Columns, column)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating column rows: %w", err)
	}

	return nil
}

// extractPrimaryKeys extracts primary key information for the table
func extractPrimaryKeys(db *sql.DB, table *Table) error {
	query := `
		SELECT
			COLUMN_NAME
		FROM
			INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE
			TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			AND CONSTRAINT_NAME = 'PRIMARY'
		ORDER BY
			ORDINAL_POSITION
	`

	rows, err := db.Query(query, table.Name)
	if err != nil {
		return fmt.Errorf("error querying primary keys for table %s: %w", table.Name, err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return fmt.Errorf("error scanning primary key: %w", err)
		}
		table.PrimaryKey = append(table.PrimaryKey, columnName)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating primary key rows: %w", err)
	}

	return nil
}

// extractForeignKeys extracts foreign key information for the table
func extractForeignKeys(db *sql.DB, table *Table) error {
	query := `
		SELECT
			COLUMN_NAME,
			REFERENCED_TABLE_NAME,
			REFERENCED_COLUMN_NAME,
			CONSTRAINT_NAME
		FROM
			INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE
			TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			AND REFERENCED_TABLE_NAME IS NOT NULL
		ORDER BY
			ORDINAL_POSITION
	`

	rows, err := db.Query(query, table.Name)
	if err != nil {
		return fmt.Errorf("error querying foreign keys for table %s: %w", table.Name, err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var fk ForeignKey
		if err := rows.Scan(
			&fk.ColumnName,
			&fk.ReferencedTable,
			&fk.ReferencedColumn,
			&fk.RelationName,
		); err != nil {
			return fmt.Errorf("error scanning foreign key: %w", err)
		}
		table.ForeignKeys = append(table.ForeignKeys, fk)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating foreign key rows: %w", err)
	}

	return nil
}
