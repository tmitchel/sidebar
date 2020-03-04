package store

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
)

func setup(t *testing.T) Database {
	psqlInfo := fmt.Sprint("host=localhost port=5432 user=postgres " +
		"password=example dbname=testing sslmode=disable")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		t.Fatalf("error opening database %v", err)
	}

	// make sure we have a good connection
	err = db.Ping()
	if err != nil {
		t.Fatalf("error pinging database %v", err)
	}

	if err := migrations(db); err != nil {
		t.Fatalf("error running migrations %v", err)
	}

	return &database{db}
}

func TestCreateUser(t *testing.T) {
	db := setup(t)
	defer db.Close()
}
