package services_test

import (
	"testing"

	"github.com/tmitchel/sidebar/mocks"
	"github.com/tmitchel/sidebar/services"
)

func TestNewCreater(t *testing.T) {
	m := &mocks.MockDatabaseCreater{}
	s, err := services.NewCreater(m)
}
