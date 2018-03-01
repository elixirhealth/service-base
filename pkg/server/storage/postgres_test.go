package storage

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	_ "github.com/mattes/migrate/database/postgres"
	"github.com/mattes/migrate/source/go-bindata"
	"github.com/stretchr/testify/assert"
	"github.com/elxirhealth/service-base/pkg/server/storage/test"
)

var (
	setUpPostgresTest func(t *testing.T) (dbURL string, tearDown func())
)

func TestMain(m *testing.M) {
	dbURL, cleanup, err := StartTestPostgres()
	if err != nil {
		if err2 := cleanup(); err2 != nil {
			log.Fatal("test postgres cleanup error: " + err2.Error())
		}
		log.Fatal("test postgres start error: " + err.Error())
	}
	as := bindata.Resource(test.AssetNames(), test.Asset)
	setUpPostgresTest = func(t *testing.T) (string, func()) {
		tearDown := SetUpTestPostgres(t, dbURL, as)
		return dbURL, tearDown
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := cleanup(); err != nil {
		log.Fatal(err.Error())
	}

	os.Exit(code)
}

func TestPostgresStartNoOp(t *testing.T) {
	// check that running this again when DB is already running is fine
	_, _, err := StartTestPostgres()
	assert.Nil(t, err)
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
