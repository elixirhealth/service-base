package storage

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/elxirhealth/service-base/pkg/server/storage/test"
	_ "github.com/lib/pq"
	_ "github.com/mattes/migrate/database/postgres"
	"github.com/mattes/migrate/source/go-bindata"
	"github.com/stretchr/testify/assert"
)

var (
	setUpPostgresTest func(t *testing.T) (dbURL string, tearDown func() error)
)

func TestMain(m *testing.M) {
	dbUrl, cleanup, err := StartTestPostgres()
	if err != nil {
		if err := cleanup(); err != nil {
			log.Fatal(err.Error())
		}
		log.Fatal(err.Error())
	}
	setUpPostgresTest = func(t *testing.T) (string, func() error) {
		return dbUrl, SetUpTestPostgresDB(t, dbUrl, bindata.Resource(
			test.AssetNames(),
			func(name string) ([]byte, error) { return test.Asset(name) },
		))
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := cleanup(); err != nil {
		log.Fatal(err.Error())
	}

	os.Exit(code)
}

func TestPostgresInsert1(t *testing.T) {
	dbURL, tearDown := setUpPostgresTest(t)
	defer tearDown()

	db, err := sql.Open("postgres", dbURL)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	result, err := db.Exec(
		"INSERT INTO test.test (id, field_1, field_2) VALUES ($1, $2, $3)",
		"id1",
		"row-1",
		1,
	)
	assert.Nil(t, err)
	nInserted, err := result.RowsAffected()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), nInserted)

	row := db.QueryRow("SELECT COUNT(*) FROM test.test")
	var count int
	err = row.Scan(&count)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
}

func TestPostgresInsert2(t *testing.T) {
	dbURL, tearDown := setUpPostgresTest(t)
	defer tearDown()

	db, err := sql.Open("postgres", dbURL)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	result, err := db.Exec(
		"INSERT INTO test.test (id, field_1, field_2) VALUES ($1, $2, $3), ($4, $5, $6)",
		"id1",
		"row-1",
		1,
		"id2",
		"row-2",
		2,
	)
	assert.Nil(t, err)
	nInserted, err := result.RowsAffected()
	assert.Nil(t, err)
	assert.Equal(t, int64(2), nInserted)

	row := db.QueryRow("SELECT COUNT(*) FROM test.test")
	var count int
	err = row.Scan(&count)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)
}
