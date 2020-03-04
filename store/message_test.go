package store

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmitchel/sidebar"
)

func TestMessageFromModel(t *testing.T) {
	m := &sidebar.WebSocketMessage{
		ID:      1,
		Event:   1,
		Content: "test-content",
	}

	n := messageFromModel(m)

	assert.Equal(t, m.ID, n.ID)
	assert.Equal(t, m.Event, n.Event)
	assert.Equal(t, m.Content, n.Content)
}

func TestMessageToModel(t *testing.T) {
	n := &webSocketMessage{
		ID:      1,
		Event:   1,
		Content: "test-content",
	}

	m := n.ToModel()

	assert.Equal(t, n.ID, m.ID)
	assert.Equal(t, n.Event, m.Event)
	assert.Equal(t, n.Content, m.Content)
}

func TestCreateMessage(t *testing.T) {
	tests := []*sidebar.WebSocketMessage{
		&sidebar.WebSocketMessage{
			ID:       1,
			Event:    1,
			Content:  "email-one",
			ToUser:   1,
			FromUser: 1,
			Channel:  1,
		},
		&sidebar.WebSocketMessage{
			ID:       2,
			Event:    2,
			Content:  "email-two",
			ToUser:   1,
			FromUser: 1,
			Channel:  1,
		},
	}

	db, err := NewWithMigration("testing")
	require.NoError(t, err)

	_, err = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert("users").
		Columns("id", "display_name", "email", "password").
		Values(1, "message-test-user", "test", "test").
		RunWith(db).Exec()
	require.NoError(t, err)

	_, err = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert("channels").
		Columns("id", "display_name").
		Values(1, "message-test-channel").
		RunWith(db).Exec()
	require.NoError(t, err)

	for _, test := range tests {
		u, err := db.CreateMessage(test)
		require.NoError(t, err)
		assert.Equal(t, test.ID, u.ID)
		assert.Equal(t, test.Event, u.Event)
		assert.Equal(t, test.Content, u.Content)

		rows := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
			Select("id", "event", "content").
			From("messages").Where(sq.Eq{"id": test.ID}).
			RunWith(db).QueryRow()
		assert.NoError(t, err)

		var dbtest sidebar.WebSocketMessage
		err = rows.Scan(&dbtest.ID, &dbtest.Event, &dbtest.Content)
		assert.NoError(t, err)
		assert.NotEmpty(t, dbtest.ID)
		assert.Equal(t, dbtest.Event, test.Event)
		assert.Equal(t, dbtest.Content, test.Content)
	}
}
