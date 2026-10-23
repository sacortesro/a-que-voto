// Package fakedb provides a scripted database/sql driver for handler tests.
// Every statement the code under test runs must match the next Step, in order.
package fakedb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

// Any matches any bind value in Step.Args.
var Any = anyValue{}

type anyValue struct{}

// Step is one expected statement and the result it returns.
type Step struct {
	// Query is a substring the whitespace-normalized SQL must contain.
	Query string
	// Args are the expected bind values. Nil skips the check.
	Args []driver.Value
	// Columns and Rows are returned to Query and QueryRow callers.
	// No rows on QueryRow means sql.ErrNoRows.
	Columns []string
	Rows    [][]driver.Value
	// Err, when set, is returned instead of a result.
	Err error
}

// Fake records whether the transaction committed or rolled back.
type Fake struct {
	t          testing.TB
	steps      []Step
	calls      int
	Committed  bool
	RolledBack bool
}

// Open returns a *sql.DB that answers statements from steps, in order.
// On cleanup it fails the test if a statement was unexpected or if steps were left unused.
func Open(t testing.TB, steps ...Step) (*sql.DB, *Fake) {
	t.Helper()

	f := &Fake{t: t, steps: steps}
	db := sql.OpenDB(connector{f: f})

	t.Cleanup(func() {
		db.Close()
		if f.calls != len(f.steps) {
			t.Errorf("fakedb: %d of %d expected statements ran", f.calls, len(f.steps))
		}
	})

	return db, f
}

func (f *Fake) next(query string, args []driver.NamedValue) (Step, error) {
	got := normalize(query)
	values := make([]driver.Value, len(args))
	for i, a := range args {
		values[i] = a.Value
	}

	if f.calls >= len(f.steps) {
		err := fmt.Errorf("fakedb: unexpected statement %q", got)
		f.t.Error(err)
		return Step{}, err
	}

	s := f.steps[f.calls]
	f.calls++

	if !strings.Contains(got, normalize(s.Query)) {
		err := fmt.Errorf("fakedb: statement %d: want query containing %q, got %q", f.calls, s.Query, got)
		f.t.Error(err)
		return Step{}, err
	}

	if s.Args != nil && !argsMatch(s.Args, values) {
		err := fmt.Errorf("fakedb: statement %d: want args %v, got %v", f.calls, s.Args, values)
		f.t.Error(err)
		return Step{}, err
	}

	return s, s.Err
}

func argsMatch(want, got []driver.Value) bool {
	if len(want) != len(got) {
		return false
	}
	for i := range want {
		if want[i] == Any {
			continue
		}
		if !reflect.DeepEqual(want[i], got[i]) {
			return false
		}
	}
	return true
}

func normalize(q string) string {
	return strings.Join(strings.Fields(q), " ")
}

type connector struct{ f *Fake }

func (c connector) Connect(context.Context) (driver.Conn, error) {
	return conn{f: c.f}, nil
}

func (c connector) Driver() driver.Driver { return driverStub{} }

type driverStub struct{}

func (driverStub) Open(string) (driver.Conn, error) {
	return nil, errors.New("fakedb: open through the connector only")
}

type conn struct{ f *Fake }

func (c conn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("fakedb: prepared statements are not supported")
}

func (c conn) Close() error { return nil }

func (c conn) Begin() (driver.Tx, error) {
	return tx{f: c.f}, nil
}

func (c conn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return tx{f: c.f}, nil
}

func (c conn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	s, err := c.f.next(query, args)
	if err != nil {
		return nil, err
	}
	return &rows{columns: s.Columns, data: s.Rows}, nil
}

func (c conn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if _, err := c.f.next(query, args); err != nil {
		return nil, err
	}
	return driver.RowsAffected(1), nil
}

type tx struct{ f *Fake }

func (t tx) Commit() error {
	t.f.Committed = true
	return nil
}

func (t tx) Rollback() error {
	t.f.RolledBack = true
	return nil
}

type rows struct {
	columns []string
	data    [][]driver.Value
	i       int
}

func (r *rows) Columns() []string { return r.columns }

func (r *rows) Close() error { return nil }

func (r *rows) Next(dest []driver.Value) error {
	if r.i >= len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.i])
	r.i++
	return nil
}
