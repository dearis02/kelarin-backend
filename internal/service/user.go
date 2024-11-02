package service

import (
	"context"
	"kelarin/internal/types"
	"net/http"

	"github.com/go-errors/errors"
)

type User interface {
	FindOne(c context.Context) error
}

type newUserDeps struct {
}

func NewUser() User {
	return &newUserDeps{}
}

func (u *newUserDeps) FindOne(c context.Context) error {
	return errors.New(types.AppErr{Code: http.StatusNotFound})
}
