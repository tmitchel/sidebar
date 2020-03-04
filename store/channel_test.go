package store

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmitchel/sidebar"
)

func TestChannelFromModel(t *testing.T) {
	m := &sidebar.Channel{
		ID:   1,
		Name: "test-channel",
	}

	n := channelFromModel(m)

	assert.Equal(t, m.ID, n.ID)
	assert.Equal(t, m.Name, n.Name)
}

func TestChannelToModel(t *testing.T) {
	n := &channel{
		ID:   1,
		Name: "test-channel",
	}

	m := n.ToModel()

	assert.Equal(t, n.ID, m.ID)
	assert.Equal(t, n.Name, m.Name)
}

func TestCreateChannel(t *testing.T) {
	tests := []*sidebar.Channel{
		&sidebar.Channel{
			ID:        1,
			Name:      "name-one",
			IsSidebar: false,
		},
		&sidebar.Channel{
			ID:        2,
			Name:      "name-two",
			IsSidebar: false,
		},
		&sidebar.Channel{
			ID:        3,
			Name:      "name-three",
			IsSidebar: true,
			Parent:    2,
		},
	}
	testsFail := []*sidebar.Channel{
		&sidebar.Channel{
			ID:        4,
			Name:      "name-two",
			IsSidebar: false,
		},
	}

	db, err := NewWithMigration("testing")
	require.NoError(t, err)

	for _, test := range tests {
		u, err := db.CreateChannel(test)
		require.NoError(t, err)
		assert.Equal(t, test.ID, u.ID)
		assert.Equal(t, test.Name, u.Name)
		assert.Equal(t, test.IsSidebar, u.IsSidebar)

		rows := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
			Select("id", "display_name", "is_sidebar").
			From("channels").Where(sq.Eq{"id": test.ID}).
			RunWith(db).QueryRow()
		assert.NoError(t, err)

		var dbtest sidebar.Channel
		err = rows.Scan(&dbtest.ID, &dbtest.Name, &dbtest.IsSidebar)
		assert.NoError(t, err)
		assert.NotEmpty(t, dbtest.ID)
		assert.Equal(t, dbtest.Name, test.Name)
		assert.Equal(t, dbtest.IsSidebar, test.IsSidebar)
	}

	for _, test := range testsFail {
		u, err := db.CreateChannel(test)
		assert.Error(t, err)
		assert.Nil(t, u)
	}
}
