package store

import (
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	db, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()
}
