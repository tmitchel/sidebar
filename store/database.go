package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"

	_ "github.com/lib/pq" // postgres drivers
)

// statement builder using postgres style
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// Database provides methods to query the database.
type Database interface {
	Adder
	Deleter
	Creater
	Getter
	Updater
	Authenticater
	sq.BaseRunner

	Close()
}

// database wraps the postgres database to provide methods.
type database struct {
	*sql.DB
}

// New connects to the postgres database
// and returns that connection.
func New(psqlInfo string) (Database, error) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, errors.Wrap(err, "Error opening database")
	}

	// make sure we have a good connection
	err = db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "Error pinging database")
	}

	return &database{db}, nil
}

// Close closes the database.
func (d *database) Close() {
	d.DB.Close()
}
