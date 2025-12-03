package schema

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/motchang/marid/internal/config"
)

func mustNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func expectError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error: %s", msg)
	}
}

func expectNoRemaining(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExtractSuccessWithFiltering(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	cfg := config.Config{Database: "testdb", Tables: "users"}

	tablesRows := sqlmock.NewRows([]string{"TABLE_NAME"}).
		AddRow("users").
		AddRow("orders")
	mock.ExpectQuery(regexp.QuoteMeta(`
                SELECT TABLE_NAME
                FROM INFORMATION_SCHEMA.TABLES
                WHERE TABLE_SCHEMA = ?
                AND TABLE_TYPE = 'BASE TABLE'
                ORDER BY TABLE_NAME
        `)).
		WithArgs(cfg.Database).
		WillReturnRows(tablesRows)

	mock.ExpectQuery(regexp.QuoteMeta(`
                SELECT
                        TABLE_COMMENT
                FROM
                        INFORMATION_SCHEMA.TABLES
                WHERE
                        TABLE_SCHEMA = DATABASE()
                        AND TABLE_NAME = ?
        `)).
		WithArgs("users").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_COMMENT"}).AddRow("users table"))

	columnsRows := sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_COMMENT"}).
		AddRow("id", "int", "NO", "PRI", "primary id").
		AddRow("email", "varchar", "NO", "UNI", "email column").
		AddRow("org_id", "int", "YES", "", "organization id")
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs("users").
		WillReturnRows(columnsRows)

	pkRows := sqlmock.NewRows([]string{"COLUMN_NAME"}).AddRow("id")
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs("users").
		WillReturnRows(pkRows)

	fkRows := sqlmock.NewRows([]string{"COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME"}).
		AddRow("org_id", "organizations", "id", "fk_users_org")
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs("users").
		WillReturnRows(fkRows)

	schema, err := Extract(db, cfg)
	mustNoError(t, err, "extracting schema")

	if len(schema.Tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(schema.Tables))
	}
	users := schema.Tables[0]
	if users.Name != "users" {
		t.Fatalf("unexpected table name: %s", users.Name)
	}
	if users.Comment != "users table" {
		t.Fatalf("unexpected table comment: %s", users.Comment)
	}

	expectedColumns := []Column{
		{Name: "id", DataType: "int", IsNullable: false, IsPrimary: true, IsUnique: false, Comment: "primary id"},
		{Name: "email", DataType: "varchar", IsNullable: false, IsPrimary: false, IsUnique: true, Comment: "email column"},
		{Name: "org_id", DataType: "int", IsNullable: true, IsPrimary: false, IsUnique: false, Comment: "organization id"},
	}
	if len(users.Columns) != len(expectedColumns) {
		t.Fatalf("unexpected column count: %d", len(users.Columns))
	}
	for i, col := range expectedColumns {
		if users.Columns[i] != col {
			t.Fatalf("unexpected column at %d: %#v", i, users.Columns[i])
		}
	}

	if len(users.PrimaryKey) != 1 || users.PrimaryKey[0] != "id" {
		t.Fatalf("unexpected primary key: %#v", users.PrimaryKey)
	}

	expectedFK := ForeignKey{ColumnName: "org_id", ReferencedTable: "organizations", ReferencedColumn: "id", RelationName: "fk_users_org"}
	if len(users.ForeignKeys) != 1 || users.ForeignKeys[0] != expectedFK {
		t.Fatalf("unexpected foreign keys: %#v", users.ForeignKeys)
	}

	expectNoRemaining(t, mock)
}

func TestGetTablesQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	cfg := config.Config{Database: "db"}
	mock.ExpectQuery(regexp.QuoteMeta(`
                SELECT TABLE_NAME
                FROM INFORMATION_SCHEMA.TABLES
                WHERE TABLE_SCHEMA = ?
                AND TABLE_TYPE = 'BASE TABLE'
                ORDER BY TABLE_NAME
        `)).
		WithArgs(cfg.Database).
		WillReturnError(errors.New("query failed"))

	_, err = getTables(db, cfg)
	expectError(t, err, "getTables query error")
	expectNoRemaining(t, mock)
}

func TestGetTablesScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	cfg := config.Config{Database: "db"}
	rows := sqlmock.NewRows([]string{"TABLE_NAME"}).AddRow(nil)
	mock.ExpectQuery(regexp.QuoteMeta(`
                SELECT TABLE_NAME
                FROM INFORMATION_SCHEMA.TABLES
                WHERE TABLE_SCHEMA = ?
                AND TABLE_TYPE = 'BASE TABLE'
                ORDER BY TABLE_NAME
        `)).
		WithArgs(cfg.Database).
		WillReturnRows(rows)

	_, err = getTables(db, cfg)
	expectError(t, err, "getTables scan error")
	expectNoRemaining(t, mock)
}

func TestGetTablesRowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	cfg := config.Config{Database: "db"}
	rows := sqlmock.NewRows([]string{"TABLE_NAME"}).AddRow("users").RowError(0, errors.New("row error"))
	mock.ExpectQuery(regexp.QuoteMeta(`
                SELECT TABLE_NAME
                FROM INFORMATION_SCHEMA.TABLES
                WHERE TABLE_SCHEMA = ?
                AND TABLE_TYPE = 'BASE TABLE'
                ORDER BY TABLE_NAME
        `)).
		WithArgs(cfg.Database).
		WillReturnRows(rows)

	_, err = getTables(db, cfg)
	expectError(t, err, "getTables row error")
	expectNoRemaining(t, mock)
}

func TestExtractColumnsQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	table := &Table{Name: "users"}
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(table.Name).
		WillReturnError(errors.New("column query failed"))

	err = extractColumns(db, table)
	expectError(t, err, "extractColumns query error")
	expectNoRemaining(t, mock)
}

func TestExtractColumnsScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	table := &Table{Name: "users"}
	rows := sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_COMMENT"}).
		AddRow("id", "int", nil, "PRI", "comment")
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(table.Name).
		WillReturnRows(rows)

	err = extractColumns(db, table)
	expectError(t, err, "extractColumns scan error")
	expectNoRemaining(t, mock)
}

func TestExtractColumnsRowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	table := &Table{Name: "users"}
	rows := sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_COMMENT"}).
		AddRow("id", "int", "NO", "PRI", "comment").
		RowError(0, errors.New("row error"))
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(table.Name).
		WillReturnRows(rows)

	err = extractColumns(db, table)
	expectError(t, err, "extractColumns row error")
	expectNoRemaining(t, mock)
}

func TestExtractPrimaryKeysQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	table := &Table{Name: "users"}
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(table.Name).
		WillReturnError(errors.New("pk query failed"))

	err = extractPrimaryKeys(db, table)
	expectError(t, err, "extractPrimaryKeys query error")
	expectNoRemaining(t, mock)
}

func TestExtractPrimaryKeysScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	table := &Table{Name: "users"}
	rows := sqlmock.NewRows([]string{"COLUMN_NAME"}).AddRow(nil)
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(table.Name).
		WillReturnRows(rows)

	err = extractPrimaryKeys(db, table)
	expectError(t, err, "extractPrimaryKeys scan error")
	expectNoRemaining(t, mock)
}

func TestExtractPrimaryKeysRowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	table := &Table{Name: "users"}
	rows := sqlmock.NewRows([]string{"COLUMN_NAME"}).AddRow("id").RowError(0, errors.New("row error"))
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(table.Name).
		WillReturnRows(rows)

	err = extractPrimaryKeys(db, table)
	expectError(t, err, "extractPrimaryKeys row error")
	expectNoRemaining(t, mock)
}

func TestExtractForeignKeysQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	table := &Table{Name: "users"}
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(table.Name).
		WillReturnError(errors.New("fk query failed"))

	err = extractForeignKeys(db, table)
	expectError(t, err, "extractForeignKeys query error")
	expectNoRemaining(t, mock)
}

func TestExtractForeignKeysScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	table := &Table{Name: "users"}
	rows := sqlmock.NewRows([]string{"COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME"}).
		AddRow(nil, "ref_table", "id", "fk_name")
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(table.Name).
		WillReturnRows(rows)

	err = extractForeignKeys(db, table)
	expectError(t, err, "extractForeignKeys scan error")
	expectNoRemaining(t, mock)
}

func TestExtractForeignKeysRowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	table := &Table{Name: "users"}
	rows := sqlmock.NewRows([]string{"COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME"}).
		AddRow("org_id", "organizations", "id", "fk_users_org").
		RowError(0, errors.New("row error"))
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(table.Name).
		WillReturnRows(rows)

	err = extractForeignKeys(db, table)
	expectError(t, err, "extractForeignKeys row error")
	expectNoRemaining(t, mock)
}

func TestExtractTableInfoCommentError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	tableName := "users"
	mock.ExpectQuery(regexp.QuoteMeta(`
                SELECT
                        TABLE_COMMENT
                FROM
                        INFORMATION_SCHEMA.TABLES
                WHERE
                        TABLE_SCHEMA = DATABASE()
                        AND TABLE_NAME = ?
        `)).
		WithArgs(tableName).
		WillReturnError(errors.New("comment query failed"))

	_, err = extractTableInfo(db, tableName)
	expectError(t, err, "table comment query error")
	expectNoRemaining(t, mock)
}

func TestExtractTableInfoColumnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	tableName := "users"
	mock.ExpectQuery(regexp.QuoteMeta(`
                SELECT
                        TABLE_COMMENT
                FROM
                        INFORMATION_SCHEMA.TABLES
                WHERE
                        TABLE_SCHEMA = DATABASE()
                        AND TABLE_NAME = ?
        `)).
		WithArgs(tableName).
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_COMMENT"}).AddRow("comment"))

	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(tableName).
		WillReturnError(errors.New("column extraction failed"))

	_, err = extractTableInfo(db, tableName)
	expectError(t, err, "column extraction failure")
	expectNoRemaining(t, mock)
}

func TestExtractTableInfoPrimaryKeyError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	tableName := "users"
	mock.ExpectQuery(regexp.QuoteMeta(`
                SELECT
                        TABLE_COMMENT
                FROM
                        INFORMATION_SCHEMA.TABLES
                WHERE
                        TABLE_SCHEMA = DATABASE()
                        AND TABLE_NAME = ?
        `)).
		WithArgs(tableName).
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_COMMENT"}).AddRow("comment"))

	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(tableName).
		WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_COMMENT"}).AddRow("id", "int", "NO", "PRI", "comment"))

	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(tableName).
		WillReturnError(errors.New("pk extraction failed"))

	_, err = extractTableInfo(db, tableName)
	expectError(t, err, "primary key extraction failure")
	expectNoRemaining(t, mock)
}

func TestExtractTableInfoForeignKeyError(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	tableName := "users"
	mock.ExpectQuery(regexp.QuoteMeta(`
                SELECT
                        TABLE_COMMENT
                FROM
                        INFORMATION_SCHEMA.TABLES
                WHERE
                        TABLE_SCHEMA = DATABASE()
                        AND TABLE_NAME = ?
        `)).
		WithArgs(tableName).
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_COMMENT"}).AddRow("comment"))

	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(tableName).
		WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_COMMENT"}).AddRow("id", "int", "NO", "PRI", "comment"))

	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(tableName).
		WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME"}).AddRow("id"))

	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(tableName).
		WillReturnError(errors.New("fk extraction failed"))

	_, err = extractTableInfo(db, tableName)
	expectError(t, err, "foreign key extraction failure")
	expectNoRemaining(t, mock)
}

func TestExtractColumnsSetsFlags(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	table := &Table{Name: "users"}
	rows := sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_COMMENT"}).
		AddRow("id", "int", "NO", "PRI", "primary id").
		AddRow("code", "varchar", "YES", "UNI", "unique code").
		AddRow("group_id", "int", "YES", "", "fk to groups")
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(table.Name).
		WillReturnRows(rows)

	err = extractColumns(db, table)
	mustNoError(t, err, "extracting columns")

	expected := []Column{
		{Name: "id", DataType: "int", IsNullable: false, IsPrimary: true, IsUnique: false, Comment: "primary id"},
		{Name: "code", DataType: "varchar", IsNullable: true, IsPrimary: false, IsUnique: true, Comment: "unique code"},
		{Name: "group_id", DataType: "int", IsNullable: true, IsPrimary: false, IsUnique: false, Comment: "fk to groups"},
	}
	for i, col := range expected {
		if table.Columns[i] != col {
			t.Fatalf("unexpected column at %d: %#v", i, table.Columns[i])
		}
	}

	expectNoRemaining(t, mock)
}

func TestExtractForeignKeysCaptured(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	table := &Table{Name: "orders"}
	rows := sqlmock.NewRows([]string{"COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME"}).
		AddRow("user_id", "users", "id", "fk_orders_users").
		AddRow("org_id", "organizations", "id", "fk_orders_orgs")
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(table.Name).
		WillReturnRows(rows)

	err = extractForeignKeys(db, table)
	mustNoError(t, err, "extracting foreign keys")

	expected := []ForeignKey{
		{ColumnName: "user_id", ReferencedTable: "users", ReferencedColumn: "id", RelationName: "fk_orders_users"},
		{ColumnName: "org_id", ReferencedTable: "organizations", ReferencedColumn: "id", RelationName: "fk_orders_orgs"},
	}
	for i, fk := range expected {
		if table.ForeignKeys[i] != fk {
			t.Fatalf("unexpected foreign key at %d: %#v", i, table.ForeignKeys[i])
		}
	}

	expectNoRemaining(t, mock)
}

func TestExtractPrimaryKeysPopulate(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	table := &Table{Name: "users"}
	rows := sqlmock.NewRows([]string{"COLUMN_NAME"}).AddRow("id").AddRow("email")
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs(table.Name).
		WillReturnRows(rows)

	err = extractPrimaryKeys(db, table)
	mustNoError(t, err, "extracting primary keys")

	expected := []string{"id", "email"}
	for i, v := range expected {
		if table.PrimaryKey[i] != v {
			t.Fatalf("unexpected primary key at %d: %s", i, table.PrimaryKey[i])
		}
	}

	expectNoRemaining(t, mock)
}

func TestExtractHandlesMultipleTables(t *testing.T) {
	db, mock, err := sqlmock.New()
	mustNoError(t, err, "creating mock")
	defer db.Close()

	cfg := config.Config{Database: "testdb"}

	tablesRows := sqlmock.NewRows([]string{"TABLE_NAME"}).AddRow("users").AddRow("orders")
	mock.ExpectQuery(regexp.QuoteMeta(`
                SELECT TABLE_NAME
                FROM INFORMATION_SCHEMA.TABLES
                WHERE TABLE_SCHEMA = ?
                AND TABLE_TYPE = 'BASE TABLE'
                ORDER BY TABLE_NAME
        `)).
		WithArgs(cfg.Database).
		WillReturnRows(tablesRows)

	// users table setup
	mock.ExpectQuery(regexp.QuoteMeta(`
                SELECT
                        TABLE_COMMENT
                FROM
                        INFORMATION_SCHEMA.TABLES
                WHERE
                        TABLE_SCHEMA = DATABASE()
                        AND TABLE_NAME = ?
        `)).
		WithArgs("users").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_COMMENT"}).AddRow("users table"))

	usersColumns := sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_COMMENT"}).
		AddRow("id", "int", "NO", "PRI", "id").
		AddRow("name", "varchar", "YES", "", "name")
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs("users").
		WillReturnRows(usersColumns)

	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs("users").
		WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME"}).AddRow("id"))

	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs("users").
		WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME"}))

	// orders table setup
	mock.ExpectQuery(regexp.QuoteMeta(`
                SELECT
                        TABLE_COMMENT
                FROM
                        INFORMATION_SCHEMA.TABLES
                WHERE
                        TABLE_SCHEMA = DATABASE()
                        AND TABLE_NAME = ?
        `)).
		WithArgs("orders").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_COMMENT"}).AddRow("orders table"))

	ordersColumns := sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_COMMENT"}).
		AddRow("id", "int", "NO", "PRI", "id").
		AddRow("user_id", "int", "NO", "", "user id")
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs("orders").
		WillReturnRows(ordersColumns)

	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs("orders").
		WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME"}).AddRow("id"))

	mock.ExpectQuery(regexp.QuoteMeta(`
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
        `)).
		WithArgs("orders").
		WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME"}).AddRow("user_id", "users", "id", "fk_orders_users"))

	schema, err := Extract(db, cfg)
	mustNoError(t, err, "extracting schema")

	if len(schema.Tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(schema.Tables))
	}
	if schema.Tables[0].Name != "users" || schema.Tables[1].Name != "orders" {
		t.Fatalf("unexpected table order: %#v", []string{schema.Tables[0].Name, schema.Tables[1].Name})
	}
	expectedFKs := []ForeignKey{{ColumnName: "user_id", ReferencedTable: "users", ReferencedColumn: "id", RelationName: "fk_orders_users"}}
	if len(schema.Tables[1].ForeignKeys) != 1 || schema.Tables[1].ForeignKeys[0] != expectedFKs[0] {
		t.Fatalf("unexpected foreign keys: %#v", schema.Tables[1].ForeignKeys)
	}

	expectNoRemaining(t, mock)
}

var _ = sql.DB{}
