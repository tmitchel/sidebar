package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
)

func TestUserFromModel(t *testing.T) {
	m := &sidebar.User{
		ID:          1,
		DisplayName: "test-user",
		Email:       "test-email",
		Password:    "test-password",
	}

	n := store.UserFromModel(m)

	assert.Equal(t, m.ID, n.ID)
	assert.Equal(t, m.DisplayName, n.DisplayName)
	assert.Equal(t, m.Email, n.Email)
	assert.Equal(t, m.Password, n.Password)
}

func TestUserToModel(t *testing.T) {
	m := &sidebar.User{
		ID:          1,
		DisplayName: "test-user",
		Email:       "test-email",
		Password:    "test-password",
	}

	n := store.UserFromModel(m)
	m = n.ToModel()

	assert.Equal(t, n.ID, m.ID)
	assert.Equal(t, n.DisplayName, m.DisplayName)
	assert.Equal(t, n.Email, m.Email)
	assert.Equal(t, n.Password, m.Password)
}
