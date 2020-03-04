package store_test

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"

	_ "github.com/lib/pq" // postgres drivers
)

func TestCreateUser(t *testing.T) {
	tests := []*sidebar.User{
		&sidebar.User{
			ID:          1,
			DisplayName: "test-one",
			Email:       "email-one",
			Password:    []byte("password-one"),
		},
		&sidebar.User{
			ID:          2,
			DisplayName: "test-two",
			Email:       "email-two",
			Password:    []byte("password-two"),
		},
	}
	tests_fail := []*sidebar.User{
		&sidebar.User{
			ID:          3,
			DisplayName: "test-two",
			Email:       "email-three",
			Password:    []byte("password-three"),
		},
	}

	db, err := store.NewWithMigration("testing")
	require.NoError(t, err)

	for _, test := range tests {
		u, err := db.CreateUser(test)
		require.NoError(t, err)
		assert.Equal(t, test.ID, u.ID)
		assert.Equal(t, test.DisplayName, u.DisplayName)
		assert.Equal(t, test.Email, u.Email)
		assert.Equal(t, test.Password, u.Password)

		rows := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
			Select("id", "email", "display_name", "password").
			From("users").Where(sq.Eq{"id": test.ID}).
			RunWith(db).QueryRow()
		assert.NoError(t, err)

		var dbtest sidebar.User
		err = rows.Scan(&dbtest.ID, &dbtest.Email, &dbtest.DisplayName, &dbtest.Password)
		assert.NoError(t, err)
		assert.NotEmpty(t, dbtest.ID)
		assert.Equal(t, dbtest.DisplayName, test.DisplayName)
		assert.Equal(t, dbtest.Email, test.Email)
	}

	for _, test := range tests_fail {
		u, err := db.CreateUser(test)
		assert.Error(t, err)
		assert.Nil(t, u)
	}
}

func TestCreateChannel(t *testing.T) {
	tests := []*sidebar.Channel{
		&sidebar.Channel{
			ID:   1,
			Name: "name-one",
		},
		&sidebar.Channel{
			ID:   2,
			Name: "name-two",
		},
	}
	tests_fail := []*sidebar.Channel{
		&sidebar.Channel{
			ID:   3,
			Name: "name-two",
		},
	}

	db, err := store.NewWithMigration("testing")
	require.NoError(t, err)

	for _, test := range tests {
		u, err := db.CreateChannel(test)
		require.NoError(t, err)
		assert.Equal(t, test.ID, u.ID)
		assert.Equal(t, test.Name, u.Name)

		rows := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
			Select("id", "display_name").
			From("channels").Where(sq.Eq{"id": test.ID}).
			RunWith(db).QueryRow()
		assert.NoError(t, err)

		var dbtest sidebar.Channel
		err = rows.Scan(&dbtest.ID, &dbtest.Name)
		assert.NoError(t, err)
		assert.NotEmpty(t, dbtest.ID)
		assert.Equal(t, dbtest.Name, test.Name)
	}

	for _, test := range tests_fail {
		u, err := db.CreateChannel(test)
		assert.Error(t, err)
		assert.Nil(t, u)
	}
}

func TestCreateSpinoff(t *testing.T) {
	tests := []*sidebar.Spinoff{
		&sidebar.Spinoff{
			ID:     1,
			Name:   "name-one",
			Parent: 1,
		},
		&sidebar.Spinoff{
			ID:     2,
			Name:   "name-two",
			Parent: 1,
		},
	}
	tests_fail := []*sidebar.Spinoff{
		&sidebar.Spinoff{
			ID:     3,
			Name:   "name-two",
			Parent: 1,
		},
	}

	db, err := store.NewWithMigration("testing")
	require.NoError(t, err)

	_, err = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Insert("channels").
		Columns("id", "display_name").
		Values(1, "test").RunWith(db).Exec()
	require.NoError(t, err)

	for _, test := range tests {
		u, err := db.CreateSpinoff(test)
		require.NoError(t, err)
		assert.Equal(t, test.ID, u.ID)
		assert.Equal(t, test.Name, u.Name)
		assert.Equal(t, test.Parent, u.Parent)

		rows := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
			Select("id", "display_name", "parent_id").
			From("spinoffs").Where(sq.Eq{"id": test.ID}).
			RunWith(db).QueryRow()
		assert.NoError(t, err)

		var dbtest sidebar.Spinoff
		err = rows.Scan(&dbtest.ID, &dbtest.Name, &dbtest.Parent)
		assert.NoError(t, err)
		assert.NotEmpty(t, dbtest.ID)
		assert.Equal(t, dbtest.Name, test.Name)
		assert.Equal(t, dbtest.Parent, test.Parent)
	}

	for _, test := range tests_fail {
		u, err := db.CreateSpinoff(test)
		assert.Error(t, err)
		assert.Nil(t, u)
	}
}

func TestCreateMessage(t *testing.T) {
	tests := []*sidebar.WebSocketMessage{
		&sidebar.WebSocketMessage{
			ID:      1,
			Event:   1,
			Content: "email-one",
		},
		&sidebar.WebSocketMessage{
			ID:      2,
			Event:   2,
			Content: "email-two",
		},
	}

	db, err := store.NewWithMigration("testing")
	require.NoError(t, err)

	for _, test := range tests {
		u, err := db.CreateMessage(test)
		require.NoError(t, err)
		assert.Equal(t, test.ID, u.ID)
		assert.Equal(t, test.Event, u.Event)
		assert.Equal(t, test.Content, u.Content)

		rows := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
			Select("id", "event", "content").
			From("messages").Where(sq.Eq{"id": test.ID}).
			RunWith(db).QueryRow()
		assert.NoError(t, err)

		var dbtest sidebar.WebSocketMessage
		err = rows.Scan(&dbtest.ID, &dbtest.Event, &dbtest.Content)
		assert.NoError(t, err)
		assert.NotEmpty(t, dbtest.ID)
		assert.Equal(t, dbtest.Event, test.Event)
		assert.Equal(t, dbtest.Content, test.Content)
	}
}
