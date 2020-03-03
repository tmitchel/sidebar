package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
