package sidebar_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/services"
	"github.com/tmitchel/sidebar/store"
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
}

func setupNewServer(t *testing.T, seed bool) (sidebar.Server, func() error) {
	db, err := store.NewWithMigration("testing")
	require.NoError(t, err)

	if seed {
		for _, test := range defaultUsers {
			_, err := db.CreateUser(test)
			require.NoError(t, err)
		}

		for _, test := range defaultChannel {
			_, err := db.CreateChannel(test)
			require.NoError(t, err)
		}
	}

	auth, err := services.NewAuthenticater(db)
	require.NoError(t, err)

	create, err := services.NewCreater(db)
	require.NoError(t, err)

	delete, err := services.NewDeleter(db)
	require.NoError(t, err)

	add, err := services.NewAdder(db)
	require.NoError(t, err)

	get, err := services.NewGetter(db)
	require.NoError(t, err)

	server := sidebar.NewServer(auth, create, delete, add, get)
	assert.NotEmpty(t, server)

	return server, func() error {
		db.Close()
		return nil
	}
}

// func TestGetUsers(t *testing.T) {
// 	server, clean := setupNewServer(t, true)
// 	defer clean()

// 	req, err := http.NewRequest("GET", "/users", nil)
// 	require.NoError(t, err)

// 	rr := httptest.NewRecorder()
// 	server.GetUsers().ServeHTTP(rr, req)
// 	assert.Equal(t, http.StatusOK, rr.Code)

// 	expect := sidebar.Response{
// 		Message: "Successfully grabbed all users",
// 		Payload: defaultUsers,
// 	}

// 	expectJSON, _ := json.Marshal(expect)
// 	assert.JSONEq(t, string(expectJSON), rr.Body.String())
// }
