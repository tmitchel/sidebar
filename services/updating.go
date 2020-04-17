package services

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/store"
	"golang.org/x/crypto/bcrypt"
)

type updater struct {
	DB store.Database
}

func NewUpdater(db store.Database) (sidebar.Updater, error) {
	return &updater{
		DB: db,
	}, nil
}

func (u *updater) UpdateUserInfo(user *sidebar.User) error {
	return u.DB.UpdateUserInformation(user)
}

func (u *updater) UpdateUserPassword(id string, newPassword, oldPassword []byte) error {
	authUser, err := u.DB.GetUser(id)
	if err != nil {
		logrus.Error("Unable to find user with id: %s", id)
		return err
	}

	if err := bcrypt.CompareHashAndPassword(authUser.Password, []byte(oldPassword)); err != nil {
		logrus.Error(err)
		return errors.Wrap(err, "Incorrect password")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logrus.Error(err)
		return errors.Wrap(err, "Unable to hash")
	}

	return u.DB.UpdateUserPassword(id, hashed)
}
