package storage

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/mattes/migrate"
	_ "github.com/mattes/migrate/database/postgres"
	"github.com/mattes/migrate/source/go-bindata"
	"gopkg.in/ory-am/dockertest.v3"
)

const (
	postgresTestDatabase   = "test"
	postgresTestPassword   = "test-pass"
	postgresDockerImageTag = "10.2-alpine"
)

// SetUpTestPostgresDB migrates the DB with the given URL and asset source.
func SetUpTestPostgresDB(t *testing.T, dbURL string, as *bindata.AssetSource) func() error {
	m, err := NewMigrate(dbURL, as)
	if err != nil {
		t.Fatal(err)
	}
	err = m.Up()
	if err != nil {
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

// StartTestPostgres starts a Postgres Docker container for use in testing. It returns the
// database URL, a function to clean up container after tests are finished, and an error (usually
// nil).
func StartTestPostgres() (dbURL string, cleanup func() error, err error) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	cleanup = func() error { return nil }
	if err != nil {
		return "", cleanup, fmt.Errorf("could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	envVars := []string{
		"POSTGRES_PASSWORD=" + postgresTestPassword,
		"POSTGRES_DB=" + postgresTestDatabase,
	}
	resource, err := pool.Run("postgres", postgresDockerImageTag, envVars)
	if err != nil {
		return "", cleanup, fmt.Errorf("could not start resource: %s", err)
	}
	cleanup = func() error { return pool.Purge(resource) }

	// exponential backoff-retry, because the application in the container might not be ready
	// to accept connections yet
	dbURL = fmt.Sprintf("postgres://postgres:%s@localhost:%s/%s?sslmode=disable",
		postgresTestPassword, resource.GetPort("5432/tcp"), postgresTestDatabase)
	if err := pool.Retry(func() error {
		var err error
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		return "", cleanup, fmt.Errorf("could not connect to docker: %s", err)
	}
	return dbURL, cleanup, nil
}
