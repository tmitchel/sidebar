package services

import (
	"github.com/pkg/errors"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
	"golang.org/x/crypto/bcrypt"
)

type updater struct {
	DB store.Database
}

// NewUpdater takes the database dependency and uses it to implement
// the sidebar.Updater interface. This interface is used to update
// objects in the database.
func NewUpdater(db store.Database) (sidebar.Updater, error) {
	return &updater{
		DB: db,
	}, nil
}

// UpdateUserInfo updates things like the display_image.
func (u *updater) UpdateUserInfo(user *sidebar.User) error {
	return u.DB.UpdateUserInformation(user)
}

// UpdateUserPassword gets the user, checks they've provided the correct
// old password, then hashes and stores the new password.
func (u *updater) UpdateUserPassword(id string, newPassword, oldPassword []byte) error {
	authUser, err := u.DB.GetUser(id)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword(authUser.Password, []byte(oldPassword)); err != nil {
		return errors.Wrap(err, "Incorrect password")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "Unable to hash")
	}

	return u.DB.UpdateUserPassword(id, hashed)
}

// UpdateChannelInfo updates things like the display_image.
func (u *updater) UpdateChannelInfo(channel *sidebar.Channel) error {
	return u.DB.UpdateChannelInformation(channel)
}
