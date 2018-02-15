package storage

import (
	"database/sql"
	"os/exec"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	_ "github.com/lib/pq" // loads "postgres" driver for database/sql
	"github.com/mattes/migrate"
	_ "github.com/mattes/migrate/database/postgres" // loads "postgres" driver for migrate
	"github.com/mattes/migrate/source/go-bindata"
)

const (
	// these vars assume the particular installation settings and paths found in the Docker
	// gcr.io/elxir-core-infra/service-base-build image
	postgresTestServerDir = "/var/lib/postgresql/10/tests"
	postgresTestServerLog = "/var/log/postgresql/tests.log"
)

// SetUpTestPostgresDB migrates the DB with the given URL and asset source.
func SetUpTestPostgresDB(t *testing.T, dbURL string, as *bindata.AssetSource) func() error {
	m, err := NewMigrate(dbURL, as)
	if err != nil {
		t.Fatal(err)
	}
	if err := backoff.Retry(m.Up, newShortExpBackoff()); err != nil {
		t.Fatal(err)
	}
	cleanup := func() error {
		if err := m.Down(); err != nil {
			return err
		}
		if err1, err2 := m.Close(); err1 != nil {
			return err1
		} else if err2 != nil {
			return err2
		}
		return nil
	}
	return cleanup
}

// NewMigrate creates a new *migrate.Migrate instance from the given database URL and migration
// assert source.
func NewMigrate(dbURL string, as *bindata.AssetSource) (*migrate.Migrate, error) {
	d, err := bindata.WithInstance(as)
	if err != nil {
		return nil, err
	}
	return migrate.NewWithSourceInstance("go-bindata", d, dbURL)
}

// StartTestPostgres starts a Postgres server for tests to use. It assumes that pg_ctl is available
// in the PATH and that the postgresTestServerDir and postgresTestServerLog are valid paths. This
// generally will only be the case when running inside of a
// gcr.io/elxir-core-infra/service-base-build Docker container.
// nolint: gas
func StartTestPostgres() (dbURL string, cleanup func() error, err error) {
	cleanup = func() error { return nil }
	cmd := exec.Command("pg_ctl",
		"-D", postgresTestServerDir,
		"-l", postgresTestServerLog,
		"start")

	if err := cmd.Run(); err != nil {
		return "", cleanup, err
	}
	cleanup = func() error {
		cmd := exec.Command("pg_ctl",
			"-D", postgresTestServerDir,
			"-l", postgresTestServerLog,
			"stop")
		return cmd.Run()
	}

	dbURL = "postgres://localhost:5432/postgres?sslmode=disable"
	op := func() error {
		var err error
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			return err
		}
		return db.Ping()
	}
	if err := backoff.Retry(op, newShortExpBackoff()); err != nil {
		return "", cleanup, err
	}

	return dbURL, cleanup, nil
}

func newShortExpBackoff() backoff.BackOff {
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 1
	bo.MaxElapsedTime = 10 * time.Second
	return bo
}