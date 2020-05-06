package store

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/tmitchel/sidebar"
)

// Authenticater provies methods for authenticating a user with data
// from the database.
type Authenticater interface {
	UserForAuth(string) (*sidebar.User, error)
}

// UserForAuth takes a user email, queries the database for that user,
// and returns the user's id and password for authenticating.
func (d *database) UserForAuth(email string) (*sidebar.User, error) {
	var authUser sidebar.User
	err := psql.Select("id", "password").
		From("users").Where(sq.Eq{"email": email}).RunWith(d).QueryRow().
		Scan(&authUser.ID, &authUser.Password)
	if err != nil {
		return nil, err
	}

	return &authUser, nil
}
