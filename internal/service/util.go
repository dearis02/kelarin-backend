package service

import (
	"kelarin/internal/types"
	"net/http"
	"time"

	"github.com/go-errors/errors"
)

type Util interface {
	ParseUserTimeZone(tz string) (*time.Location, error)
}

type utilImpl struct{}

func NewUtil() Util {
	return &utilImpl{}
}

func (u *utilImpl) ParseUserTimeZone(tz string) (*time.Location, error) {
	if tz == "" {
		return time.Local, nil
	}

	t, err := time.LoadLocation(tz)
	if err != nil {
		return nil, errors.New(types.AppErr{Code: http.StatusBadRequest, Message: "invalid timezone"})
	}

	return t, nil
}
