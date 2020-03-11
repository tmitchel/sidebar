package sidebar_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestGetUsers(t *testing.T) {
	server, clean := setupNewServer(t, true)
	defer clean()

	req, err := http.NewRequest("GET", "/users", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	server.GetUsers().ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp []*sidebar.User
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	for i, u := range defaultUsers {
		assert.Equal(t, u.DisplayName, resp[i].DisplayName)
		assert.Equal(t, u.Email, resp[i].Email)
	}
}

func TestGetChannels(t *testing.T) {
	server, clean := setupNewServer(t, true)
	defer clean()

	req, err := http.NewRequest("GET", "/channels", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	server.GetChannels().ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp []*sidebar.Channel
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	for i, d := range defaultChannel {
		assert.Equal(t, d.Name, resp[i].Name)
	}
}
