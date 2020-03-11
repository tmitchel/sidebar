// +build integration

package sidebar_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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
		DisplayName: "default-server-user-one",
		Email:       "default-server-email-user-one",
		Password:    []byte("password-one"),
	},
	&sidebar.User{
		ID:          1002,
		DisplayName: "default-server-user-two",
		Email:       "default-server-email-user-two",
		Password:    []byte("password-two"),
	},
}

var defaultChannel = []*sidebar.Channel{
	&sidebar.Channel{
		ID:        1001,
		Name:      "default-server-channel-one",
		IsSidebar: false,
	},
}

var defaultSidebar = []*sidebar.Channel{
	&sidebar.Channel{
		ID:        1002,
		Name:      "default-server-sidebar-one",
		IsSidebar: true,
	},
}

var defaultMessage = []*sidebar.WebSocketMessage{
	&sidebar.WebSocketMessage{
		ID:      1001,
		Event:   1,
		Content: "default-server-message-one",
	},
}

// Need to finish adding tests for all routes and then start adding
// tests which require the server to do some error handling.

func setupNewServer(t *testing.T, seed bool) (sidebar.Server, store.Database, func() error) {
	db, err := store.NewWithMigration("testing")
	require.NoError(t, err)

	if seed {
		for i, test := range defaultUsers {
			stored, err := db.CreateUser(test)
			require.NoError(t, err)
			defaultUsers[i].ID = stored.ID
		}

		for i, test := range defaultChannel {
			stored, err := db.CreateChannel(test)
			require.NoError(t, err)
			defaultChannel[i].ID = stored.ID
		}

		for i, test := range defaultSidebar {
			test.Parent = defaultChannel[0].ID
			stored, err := db.CreateChannel(test)
			require.NoError(t, err)
			defaultSidebar[i].ID = stored.ID
		}

		for i, test := range defaultMessage {
			test.ToUser = defaultUsers[0].ID
			test.FromUser = defaultUsers[1].ID
			test.Channel = defaultChannel[0].ID
			stored, err := db.CreateMessage(test)
			require.NoError(t, err)
			defaultMessage[i].ID = stored.ID
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

	return server, db, func() error {
		db.Close()
		return nil
	}
}

func TestGetUser(t *testing.T) {
	server, db, clean := setupNewServer(t, true)
	defer clean()

	users, err := db.GetUsers()
	require.NoError(t, err)

	for _, u := range users {
		url := "/test-user/" + strconv.Itoa(u.ID)
		req, err := http.NewRequest("GET", url, nil)
		require.NoError(t, err)

		server.Serve().HandleFunc("/test-user/{id}", server.GetUser())
		rr := httptest.NewRecorder()
		server.Serve().ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		var resp sidebar.User
		err = json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, u.ID, resp.ID)
		assert.Equal(t, u.DisplayName, resp.DisplayName)
		assert.Equal(t, u.Email, resp.Email)
	}
}

func TestGetChannel(t *testing.T) {
	server, db, clean := setupNewServer(t, true)
	defer clean()

	channels, err := db.GetChannels()
	require.NoError(t, err)

	for _, u := range channels {
		url := "/test-channel/" + strconv.Itoa(u.ID)
		req, err := http.NewRequest("GET", url, nil)
		require.NoError(t, err)

		server.Serve().HandleFunc("/test-channel/{id}", server.GetChannel())
		rr := httptest.NewRecorder()
		server.Serve().ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		var resp sidebar.Channel
		err = json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, u.ID, resp.ID)
		assert.Equal(t, u.Name, resp.Name)
	}
}

func TestGetMessage(t *testing.T) {
	server, db, clean := setupNewServer(t, true)
	defer clean()

	messages, err := db.GetMessages()
	require.NoError(t, err)

	for _, u := range messages {
		url := "/test-messages/" + strconv.Itoa(u.ID)
		req, err := http.NewRequest("GET", url, nil)
		require.NoError(t, err)

		server.Serve().HandleFunc("/test-messages/{id}", server.GetMessage())
		rr := httptest.NewRecorder()
		server.Serve().ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		var resp sidebar.WebSocketMessage
		err = json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, u.ID, resp.ID)
		assert.Equal(t, u.Event, resp.Event)
		assert.Equal(t, u.Content, resp.Content)
		assert.Equal(t, u.ToUser, resp.ToUser)
		assert.Equal(t, u.FromUser, resp.FromUser)
		assert.Equal(t, u.Channel, resp.Channel)
	}
}

func TestGetUsers(t *testing.T) {
	server, _, clean := setupNewServer(t, true)
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
		assert.Equal(t, u.ID, resp[i].ID)
		assert.Equal(t, u.DisplayName, resp[i].DisplayName)
		assert.Equal(t, u.Email, resp[i].Email)
	}
}

func TestGetChannels(t *testing.T) {
	server, _, clean := setupNewServer(t, true)
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
		if resp[i].IsSidebar {
			continue
		}
		assert.Equal(t, d.ID, resp[i].ID)
		assert.Equal(t, d.Name, resp[i].Name)
	}
	for i, d := range defaultSidebar {
		if !resp[i].IsSidebar {
			continue
		}
		assert.Equal(t, d.ID, resp[i].ID)
		assert.Equal(t, d.Name, resp[i].Name)
	}
}

func TestGetSidebars(t *testing.T) {
	server, _, clean := setupNewServer(t, true)
	defer clean()

	req, err := http.NewRequest("GET", "/sidebars", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	server.GetSidebars().ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp []*sidebar.Channel
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	for i, d := range defaultChannel {
		if d.IsSidebar {
			assert.Equal(t, d.ID, resp[i].ID)
			assert.Equal(t, d.Name, resp[i].Name)
			assert.Equal(t, d.Parent, resp[i].Parent)
		}
	}
}

func TestGetMessages(t *testing.T) {
	server, _, clean := setupNewServer(t, true)
	defer clean()

	req, err := http.NewRequest("GET", "/messages", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	server.GetMessages().ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp []*sidebar.WebSocketMessage
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	for i, d := range defaultMessage {
		assert.Equal(t, d.ID, resp[i].ID)
		assert.Equal(t, d.Content, resp[i].Content)
		assert.Equal(t, d.Event, resp[i].Event)
		assert.Equal(t, d.Channel, resp[i].Channel)
		assert.Equal(t, d.ToUser, resp[i].ToUser)
		assert.Equal(t, d.FromUser, resp[i].FromUser)
	}
}

func TestCreateUser(t *testing.T) {
	server, db, clean := setupNewServer(t, false)
	defer clean()

	for _, user := range defaultUsers {
		rr := httptest.NewRecorder()

		u, _ := json.Marshal(user)
		req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(u))
		require.NoError(t, err)

		server.CreateUser().ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		var resp *sidebar.User
		err = json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		assert.Equal(t, user.DisplayName, resp.DisplayName)
		assert.Equal(t, user.Email, resp.Email)

		dbUser, err := db.GetUser(resp.ID)
		require.NoError(t, err)

		assert.Equal(t, dbUser.ID, resp.ID)
		assert.Equal(t, dbUser.DisplayName, resp.DisplayName)
		assert.Equal(t, dbUser.Email, resp.Email)
	}
}

func TestCreateChannel(t *testing.T) {
	server, db, clean := setupNewServer(t, false)
	defer clean()

	for _, channel := range defaultChannel {
		rr := httptest.NewRecorder()

		c, _ := json.Marshal(channel)
		req, err := http.NewRequest("POST", "/channel", bytes.NewBuffer(c))
		require.NoError(t, err)

		server.CreateChannel().ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		var resp *sidebar.Channel
		err = json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		assert.Equal(t, channel.Name, resp.Name)

		dbChannel, err := db.GetChannel(resp.ID)
		require.NoError(t, err)

		assert.Equal(t, dbChannel.ID, resp.ID)
		assert.Equal(t, dbChannel.Name, resp.Name)
	}
}
