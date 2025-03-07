package service

import (
	"kelarin/internal/types"
	"net/http"
	"time"

	"github.com/go-errors/errors"
)

type Util interface {
	ParseUserTimeZone(tz string) (*time.Location, error)

	// to get correct timezone offset on parsing time only, we must include the year
	//
	// issue: https://github.com/golang/go/issues/34101#issuecomment-528260666
	NormalizeTimeOnlyTz(timeOnly time.Time) time.Time
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

func (u *utilImpl) NormalizeTimeOnlyTz(timeOnly time.Time) time.Time {
	return time.Date(time.Now().Year(), 0, 0, timeOnly.Hour(), timeOnly.Minute(), timeOnly.Second(), 0, timeOnly.Location())
}
