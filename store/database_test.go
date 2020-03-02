package store_test

import (
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/tmitchel/sidebar/store"
)

func TestNew(t *testing.T) {
	db, err := store.New()
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()
}
