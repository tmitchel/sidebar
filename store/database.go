package store

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

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
	Empty() error

	Close()
}

// database wraps the postgres database to provide methods.
type database struct {
	*sql.DB
}

// New connects to the postgres database
// and returns that connection.
func New(psqlInfo string) (Database, error) {
	for {
		db, err := sql.Open("postgres", psqlInfo)
		if err != nil {
			return nil, errors.Wrap(err, "Error opening database")
		}

		// make sure we have a good connection
		err = db.Ping()
		if err != nil {
			time.Sleep(time.Second)
			logrus.Errorf("Error pinging database %v", err)
		} else {
			logrus.Info("Connected to database.")
			return &database{db}, nil
		}
	}
}

// Empty checks if the database has any workspaces
func (d *database) Empty() error {
	rows, err := psql.Select("id").
		From("workspaces").RunWith(d).Query()
	if err != nil {
		return err
	}

	if !rows.Next() {
		return errors.New("No workspaces exist")
	}

	return nil
}

// Close closes the database.
func (d *database) Close() {
	d.DB.Close()
}
