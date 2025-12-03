package sqlmock

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"
)

type mock struct {
	expectations []*ExpectedQuery
	mu           sync.Mutex
}

type ExpectedQuery struct {
	mock      *mock
	regex     *regexp.Regexp
	args      []driver.Value
	rows      *Rows
	err       error
	fulfilled bool
}

type Sqlmock interface {
	ExpectQuery(query string) *ExpectedQuery
	ExpectationsWereMet() error
}

func New(options ...interface{}) (*sql.DB, Sqlmock, error) {
	m := &mock{}
	driverName := fmt.Sprintf("gosqlmock-%d", time.Now().UnixNano())
	sql.Register(driverName, &driverStub{mock: m})
	db, err := sql.Open(driverName, "")
	if err != nil {
		return nil, nil, err
	}
	return db, m, nil
}

func (m *mock) ExpectQuery(query string) *ExpectedQuery {
	m.mu.Lock()
	defer m.mu.Unlock()
	normalized := normalize(query)
	exp := &ExpectedQuery{mock: m, regex: regexp.MustCompile(normalized)}
	m.expectations = append(m.expectations, exp)
	return exp
}

func (m *mock) ExpectationsWereMet() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, exp := range m.expectations {
		if !exp.fulfilled {
			return fmt.Errorf("there is a remaining expectation for query: %s", exp.regex.String())
		}
	}
	return nil
}

func (e *ExpectedQuery) WithArgs(args ...interface{}) *ExpectedQuery {
	e.args = make([]driver.Value, len(args))
	for i, arg := range args {
		e.args[i] = driver.Value(arg)
	}
	return e
}

func (e *ExpectedQuery) WillReturnRows(rows *Rows) *ExpectedQuery {
	e.rows = rows
	return e
}

func (e *ExpectedQuery) WillReturnError(err error) *ExpectedQuery {
	e.err = err
	return e
}

type driverStub struct {
	mock *mock
}

func (d *driverStub) Open(name string) (driver.Conn, error) {
	return &conn{mock: d.mock}, nil
}

type conn struct {
	mock *mock
}

func (c *conn) Prepare(query string) (driver.Stmt, error) { return nil, errors.New("not implemented") }
func (c *conn) Close() error                              { return nil }
func (c *conn) Begin() (driver.Tx, error)                 { return nil, errors.New("not implemented") }

func (c *conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.mock.mu.Lock()
	defer c.mock.mu.Unlock()
	if len(c.mock.expectations) == 0 {
		return nil, fmt.Errorf("unexpected query: %s", query)
	}

	exp := c.mock.expectations[0]
	normalizedQuery := normalize(query)
	if !exp.regex.MatchString(normalizedQuery) {
		return nil, fmt.Errorf("query %s did not match expectation %s", query, exp.regex.String())
	}

	if len(exp.args) > 0 {
		if len(args) != len(exp.args) {
			return nil, fmt.Errorf("query %s args %v do not match expectation %v", query, args, exp.args)
		}
		for i, arg := range args {
			if arg.Value != exp.args[i] {
				return nil, fmt.Errorf("arg %d mismatch: %v vs %v", i, arg.Value, exp.args[i])
			}
		}
	}

	exp.fulfilled = true
	c.mock.expectations = c.mock.expectations[1:]

	if exp.err != nil {
		return nil, exp.err
	}

	if exp.rows == nil {
		return nil, errors.New("no rows set for expectation")
	}

	return exp.rows.clone(), nil
}

func (c *conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return nil, errors.New("exec not supported")
}

func (c *conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}

func normalize(query string) string {
	return strings.Join(strings.Fields(query), " ")
}

var _ driver.QueryerContext = (*conn)(nil)
var _ driver.ExecerContext = (*conn)(nil)
var _ driver.ConnPrepareContext = (*conn)(nil)

// Rows implements a lightweight row set for expectations.
type Rows struct {
	columns []string
	data    [][]driver.Value
	rowErr  map[int]error
	pos     int
	err     error
}

func NewRows(columns []string) *Rows {
	return &Rows{columns: columns, rowErr: map[int]error{}}
}

func (r *Rows) AddRow(values ...interface{}) *Rows {
	row := make([]driver.Value, len(values))
	for i, v := range values {
		row[i] = driver.Value(v)
	}
	r.data = append(r.data, row)
	return r
}

func (r *Rows) RowError(row int, err error) *Rows {
	r.rowErr[row] = err
	return r
}

func (r *Rows) Columns() []string { return append([]string(nil), r.columns...) }

func (r *Rows) Close() error { return nil }

func (r *Rows) Next(dest []driver.Value) error {
	if r.pos >= len(r.data) {
		return io.EOF
	}

	copy(dest, r.data[r.pos])
	if err, ok := r.rowErr[r.pos]; ok {
		r.pos++
		r.err = err
		return err
	}
	r.pos++
	return nil
}

func (r *Rows) clone() *Rows {
	clone := &Rows{columns: append([]string(nil), r.columns...), rowErr: map[int]error{}}
	for _, row := range r.data {
		cloneRow := make([]driver.Value, len(row))
		copy(cloneRow, row)
		clone.data = append(clone.data, cloneRow)
	}
	for idx, err := range r.rowErr {
		clone.rowErr[idx] = err
	}
	return clone
}

// Err returns the error encountered during iteration, if any.
func (r *Rows) Err() error {
	return r.err
}

var _ driver.Rows = (*Rows)(nil)
var _ driver.RowsNextResultSet = (*Rows)(nil)

func (r *Rows) HasNextResultSet() bool { return false }
func (r *Rows) NextResultSet() error   { return io.EOF }
