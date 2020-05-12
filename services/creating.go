package services

import (
	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
	"golang.org/x/crypto/bcrypt"
)

type creater struct {
	DB store.Database
}

// NewCreater wraps a database connection with a *creater that
// implements the sidebar.Creater interface.
func NewCreater(db store.Database) (sidebar.Creater, error) {
	return &creater{
		DB: db,
	}, nil
}

// CreateWorkspace creates a new workspace and saves it. A new ID and join token
// are generated randomly.
func (c *creater) CreateWorkspace(w *sidebar.Workspace) (*sidebar.Workspace, error) {
	if w.DisplayName == "" {
		return nil, errors.New("Invalid fields when trying to create workspace")
	}

	w.ID = uuid.New().String()
	w.Token = uuid.New().String()
	return c.DB.CreateWorkspace(w)
}

// CreateUser takes the new user's information and the token they were sent
// to create an account. The user is assigned an id, their password is hashed
// and they are given a default profile image if they didn't provide one. The
// user is stored in the database.
func (c *creater) CreateUser(u *sidebar.User) (*sidebar.User, error) {
	if u.DisplayName == "" || u.Email == "" || len(u.Password) == 0 {
		return nil, errors.New("Invalid fields when trying to create user")
	}

	u.ID = uuid.New().String()

	hashed, err := bcrypt.GenerateFromPassword(u.Password, bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "Error hashing password")
	}

	u.Password = hashed
	if u.ProfileImg == "" {
		u.ProfileImg = "https://randomuser.me/api/portraits/women/81.jpg"
	}

	ws, err := c.DB.GetDefaultWorkspace()
	if err != nil {
		return nil, err
	}

	user, err := c.DB.CreateUser(u)
	if err != nil {
		return nil, err
	}

	if err := c.DB.AddUserToWorkspace(u.ID, ws.ID); err != nil {
		return nil, err
	}

	return user, nil
}

// CreateChannel takes the information sent for creating a new channel,
// gives it an id and a default image if one isn't provided. The channel is
// saved.
func (c *creater) CreateChannel(ch *sidebar.Channel, wid string) (*sidebar.Channel, error) {
	if ch.Name == "" {
		return nil, errors.New("Invalid fields when trying to create channel")
	}

	ch.ID = uuid.New().String()
	if ch.Image == "" {
		ch.Image = "https://randomuser.me/api/portraits/women/81.jpg"
	}

	// check if workspace exists
	if err := c.DB.GetWorkspaceExists(wid); err != nil {
		return nil, err
	}

	channel, err := c.DB.CreateChannel(ch)
	if err != nil {
		return nil, err
	}

	// add channel to the workspace
	if err := c.DB.AddChannelToWorkspace(ch.ID, wid); err != nil {
		return nil, err
	}

	return channel, nil
}

// CreateMessage gives the message an id and stores in the database.
func (c *creater) CreateMessage(m *sidebar.ChatMessage) (*sidebar.ChatMessage, error) {
	m.ID = uuid.New().String()
	return c.DB.CreateMessage(m)
}
