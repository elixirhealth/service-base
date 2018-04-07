package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/cenkalti/backoff"
	"github.com/drausin/libri/libri/common/errors"
	_ "github.com/lib/pq" // loads "postgres" driver for database/sql
	"github.com/mattes/migrate"
	_ "github.com/mattes/migrate/database/postgres" // loads "postgres" driver for migrate
	"github.com/mattes/migrate/source/go-bindata"
	"go.uber.org/zap"
)

const (
	migrationPrefix = "[migration] "

	// these vars assume the particular installation settings and paths found in the Docker
	// gcr.io/elixir-core-prod/service-base-build image
	postgresTestServerDir = "/var/lib/postgresql/10/tests"
	postgresTestServerLog = "/var/log/postgresql/tests.log"
	postgresDBName        = "postgres"
)

// ColDest is a mapping from a column name to a sql.Scan destination type.
type ColDest struct {
	col  string
	dest interface{}
}

// SplitColDests returns a list of column names and their corresponding destination types plus a
// given number of extra destination capacity.
func SplitColDests(nExtraDest int, cds []*ColDest) ([]string, []interface{}) {
	dests := make([]interface{}, len(cds), len(cds)+nExtraDest)
	cols := make([]string, len(cds))
	for i, colDest := range cds {
		cols[i] = colDest.col
		dests[i] = colDest.dest
	}
	return cols, dests
}

// QueryRows is a container for the result of a Select query.
type QueryRows interface {
	Scan(dest ...interface{}) error
	Next() bool
	Close() error
	Err() error
}

// Querier is an interface wrapper around Squirrel query builders and their results for improved
// mocking.
type Querier interface {
	SelectQueryContext(ctx context.Context, b sq.SelectBuilder) (QueryRows, error)
	SelectQueryRowContext(ctx context.Context, b sq.SelectBuilder) sq.RowScanner
	InsertExecContext(ctx context.Context, b sq.InsertBuilder) (sql.Result, error)
	UpdateExecContext(ctx context.Context, b sq.UpdateBuilder) (sql.Result, error)
}

type querierImpl struct{}

// NewQuerier returns a new Querier.
func NewQuerier() Querier {
	return &querierImpl{}
}

func (q *querierImpl) SelectQueryContext(
	ctx context.Context, b sq.SelectBuilder,
) (QueryRows, error) {
	return b.QueryContext(ctx)
}

func (q *querierImpl) SelectQueryRowContext(
	ctx context.Context, b sq.SelectBuilder,
) sq.RowScanner {
	return b.QueryRowContext(ctx)
}

func (q *querierImpl) InsertExecContext(
	ctx context.Context, b sq.InsertBuilder,
) (sql.Result, error) {
	return b.ExecContext(ctx)
}

func (q *querierImpl) UpdateExecContext(
	ctx context.Context, b sq.UpdateBuilder,
) (sql.Result, error) {
	return b.ExecContext(ctx)
}

// Migrator handles Postgres DB migrations. It is a thin wrapper around *Migrate in mattes/migrate
// package.
type Migrator interface {

	// Up migrates the DB up to the latest state.
	Up() error

	// Down migrates the DB all the way to the empty state.
	Down() error
}

type bindataMigrator struct {
	dbURL  string
	as     *bindata.AssetSource
	logger migrate.Logger
}

// NewBindataMigrator creates a new Migrator from the given go-bindata asset source and using the
// given logger.
func NewBindataMigrator(dbURL string, as *bindata.AssetSource, logger migrate.Logger) Migrator {
	return &bindataMigrator{
		dbURL:  dbURL,
		as:     as,
		logger: logger,
	}
}

// Up migrates the DB up to the latest state.
func (bm *bindataMigrator) Up() error {
	m := bm.newInner()
	op := func() error {
		err := m.Up()
		if err == migrate.ErrNoChange {
			return nil
		}
		return err
	}
	if err := backoff.Retry(op, newShortExpBackoff()); err != nil {
		return err
	}
	err1, err2 := m.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// Down migrates the DB down to the empty state.
func (bm *bindataMigrator) Down() error {
	m := bm.newInner()
	if err := m.Down(); err != nil {
		return err
	}
	err1, err2 := m.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

func (bm *bindataMigrator) newInner() *migrate.Migrate {
	d, err := bindata.WithInstance(bm.as)
	errors.MaybePanic(err) // should never happen
	m, err := migrate.NewWithSourceInstance("go-bindata", d, bm.dbURL)
	errors.MaybePanic(err) // should never happen
	m.Log = bm.logger
	return m
}

// LogLogger implements migrate.Logger via log.Printf
type LogLogger struct{}

// Printf prints the given format and args.
func (ll *LogLogger) Printf(format string, v ...interface{}) {
	log.Printf(migrationPrefix+format, v...)
}

// Verbose indicates whether the logger is verbose. Fixed to false.
func (ll *LogLogger) Verbose() bool {
	return false
}

// ZapLogger implements migrate.Logger by wrapper a *zap.Logger
type ZapLogger struct {
	*zap.Logger
}

// Printf prints the given format and args as INFO messages.
func (zl *ZapLogger) Printf(format string, v ...interface{}) {
	zl.Info(migrationPrefix + fmt.Sprintf(strings.TrimSpace(format), v...))
}

// Verbose indicates whether the logger is verbose. Fixed to false.
func (zl *ZapLogger) Verbose() bool {
	return false
}

// SetUpTestPostgres migrates the DB with the given URL and asset source.
func SetUpTestPostgres(t *testing.T, dbURL string, as *bindata.AssetSource) func() {
	logger := &LogLogger{}
	m := NewBindataMigrator(dbURL, as, logger)
	if err := m.Up(); err != nil {
		t.Fatal("migration up error: " + err.Error())
	}
	tearDown := func() {
		if err := m.Down(); err != nil {
			t.Fatal("migration down error: " + err.Error())
		}
	}
	return tearDown
}

// StartTestPostgres starts a Postgres server for tests to use. It assumes that pg_ctl is available
// in the PATH and that the postgresTestServerDir and postgresTestServerLog are valid paths. This
// generally will only be the case when running inside of a
// gcr.io/elixir-core-prod/service-base-build Docker container.
// nolint: gas
func StartTestPostgres() (dbURL string, cleanup func() error, err error) {
	dbURL = fmt.Sprintf("postgres://localhost:5432/%s?sslmode=disable", postgresDBName)
	cleanup = func() error {
		stopCmd := exec.Command("pg_ctl",
			"-D", postgresTestServerDir,
			"-l", postgresTestServerLog,
			"stop")
		return stopCmd.Run()
	}
	if err = dbReady(dbURL, false); err == nil {
		// DB already running (perhaps from previous test, so don't try to start again
		return dbURL, cleanup, nil
	}
	startCmd := exec.Command("pg_ctl",
		"-D", postgresTestServerDir,
		"-l", postgresTestServerLog,
		"start")
	if err = startCmd.Run(); err != nil {
		noCleanup := func() error { return nil }
		return "", noCleanup, err
	}

	// wait for DB to become available
	err = dbReady(dbURL, true)
	return dbURL, cleanup, err
}

func dbReady(dbURL string, retry bool) error {
	op := func() error {
		var err error
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			return err
		}
		return db.Ping()
	}

	if err := op(); err == nil || !retry {
		return err
	}
	return backoff.Retry(op, newShortExpBackoff())
}

func newShortExpBackoff() backoff.BackOff {
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 1
	bo.MaxElapsedTime = 5 * time.Second
	return bo
}
