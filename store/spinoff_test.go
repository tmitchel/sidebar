package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
)

func TestSpinoffFromModel(t *testing.T) {
	m := &sidebar.Spinoff{
		ID:     1,
		Name:   "test-spinoff",
		Parent: 1,
	}

	n := store.SpinoffFromModel(m)

	assert.Equal(t, m.ID, n.ID)
	assert.Equal(t, m.Name, n.Name)
	assert.Equal(t, m.Parent, n.Parent)
}

func TestSpinoffToModel(t *testing.T) {
	m := &sidebar.Spinoff{
		ID:     1,
		Name:   "test-spinoff",
		Parent: 1,
	}

	n := store.SpinoffFromModel(m)
	m = n.ToModel()

	assert.Equal(t, n.ID, m.ID)
	assert.Equal(t, n.Name, m.Name)
	assert.Equal(t, n.Parent, m.Parent)
}
