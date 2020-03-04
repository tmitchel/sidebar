package store

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmitchel/sidebar"
)

var defaultUsers = []*sidebar.User{
	&sidebar.User{
		ID:          1001,
		DisplayName: "default-user-one",
		Email:       "default-email-user-one",
		Password:    []byte("password-one"),
	},
	&sidebar.User{
		ID:          1002,
		DisplayName: "default-user-two",
		Email:       "default-email-user-two",
		Password:    []byte("password-two"),
	},
}

var defaultChannel = []*sidebar.Channel{
	&sidebar.Channel{
		ID:        1001,
		Name:      "default-channel-one",
		IsSidebar: false,
	},
	&sidebar.Channel{
		ID:        1002,
		Name:      "default-sidebar-one",
		IsSidebar: true,
		Parent:    1,
	},
}

func userTestSetup(t *testing.T) Database {
	db, err := NewWithMigration("testing")
	require.NoError(t, err)

	for _, test := range defaultUsers {
		_, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
			Insert("users").Columns("id", "display_name", "email", "password").
			Values(test.ID, test.DisplayName, test.Email, test.Password).RunWith(db).Exec()

		require.NoError(t, err)
	}

	for _, test := range defaultChannel {
		_, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
			Insert("channels").Columns("id", "display_name").
			Values(test.ID, test.Name).RunWith(db).Exec()

		require.NoError(t, err)
	}
	return db
}

func TestUsers(t *testing.T) {
	db := userTestSetup(t)
	t.Run("TestCreateUser", testCreateUser(t, db))
	t.Run("TestGetUser", testGetUser(t, db))
}

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

func testCreateUser(t *testing.T, db Database) func(*testing.T) {
	return func(*testing.T) {
		tests := []*sidebar.User{
			&sidebar.User{
				DisplayName: "test-user-one",
				Email:       "test-email-user-one",
				Password:    []byte("password-one"),
			},
			&sidebar.User{
				DisplayName: "test-user-two",
				Email:       "test-email-user-two",
				Password:    []byte("password-two"),
			},
		}
		testsFail := []*sidebar.User{
			&sidebar.User{
				DisplayName: "test-user-two",
				Email:       "email-three",
				Password:    []byte("password-three"),
			},
		}

		for _, test := range tests {
			u, err := db.CreateUser(test)
			require.NoError(t, err)
			assert.NotEmpty(t, u.ID)
			assert.Equal(t, test.DisplayName, u.DisplayName)
			assert.Equal(t, test.Email, u.Email)
			assert.Equal(t, test.Password, u.Password)

			rows := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
				Select("id", "email", "display_name", "password").
				From("users").Where(sq.Eq{"id": u.ID}).
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
}

func testGetUser(t *testing.T, db Database) func(*testing.T) {
	return func(*testing.T) {
		for _, test := range defaultUsers {
			user, err := db.GetUser(test.ID)
			require.NoError(t, err)
			assert.Equal(t, test.DisplayName, user.DisplayName)
			assert.Equal(t, test.Email, user.Email)
			assert.Equal(t, test.Password, user.Password)
		}
	}
}
