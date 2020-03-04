package store

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmitchel/sidebar"
)

func TestUserFromModel(t *testing.T) {
	m := &sidebar.User{
		ID:          1,
		DisplayName: "test-user",
		Email:       "test-email",
		Password:    []byte("test-password"),
	}

	n := userFromModel(m)

	assert.Equal(t, m.ID, n.ID)
	assert.Equal(t, m.DisplayName, n.DisplayName)
	assert.Equal(t, m.Email, n.Email)
	assert.Equal(t, m.Password, n.Password)
}

func TestUserToModel(t *testing.T) {
	n := &user{
		ID:          1,
		DisplayName: "test-user",
		Email:       "test-email",
		Password:    []byte("test-password"),
	}

	m := n.ToModel()

	assert.Equal(t, n.ID, m.ID)
	assert.Equal(t, n.DisplayName, m.DisplayName)
	assert.Equal(t, n.Email, m.Email)
	assert.Equal(t, n.Password, m.Password)
}

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
	testsFail := []*sidebar.User{
		&sidebar.User{
			ID:          3,
			DisplayName: "test-two",
			Email:       "email-three",
			Password:    []byte("password-three"),
		},
	}

	db, err := NewWithMigration("testing")
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

	for _, test := range testsFail {
		u, err := db.CreateUser(test)
		assert.Error(t, err)
		assert.Nil(t, u)
	}
}
