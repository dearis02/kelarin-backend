package service

import (
	"context"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	"net/http"

	"github.com/go-errors/errors"
)

type User interface {
	FindOne(c context.Context) error
}

type userImpl struct {
	userRepo repository.User
}

func NewUser(userRepo repository.User) User {
	return &userImpl{
		userRepo: userRepo,
	}
}

func (u *userImpl) FindOne(c context.Context) error {
	return errors.New(types.AppErr{Code: http.StatusNotFound})
}
