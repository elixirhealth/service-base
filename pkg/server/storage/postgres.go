package storage

import (
	"database/sql"
	"testing"

	"os/exec"
	"time"

	"github.com/cenkalti/backoff"
	_ "github.com/lib/pq"
	"github.com/mattes/migrate"
	_ "github.com/mattes/migrate/database/postgres"
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

	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 1
	bo.MaxElapsedTime = 10 * time.Second
	if err := backoff.Retry(m.Up, bo); err != nil {
		t.Fatal(err)
	}
	return m.Down
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
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 1
	bo.MaxElapsedTime = 10 * time.Second
	if err := backoff.Retry(op, bo); err != nil {
		return "", cleanup, err
	}

	return dbURL, cleanup, nil
}
