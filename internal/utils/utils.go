package utils

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/go-errors/errors"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

func ParseLatLngFromHexStr(hexStr string) (float64, float64, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, 0, err
	}

	var ewkbPoint ewkb.Point
	if err := ewkbPoint.Scan(bytes); err != nil {
		return 0, 0, err
	}

	return ewkbPoint.Y(), ewkbPoint.X(), nil
}

func IsDateBetween(targetDate string, startDate, endDate time.Time, layout string) (bool, error) {
	tDate, err := time.Parse(layout, targetDate)
	if err != nil {
		return false, errors.Errorf("invalid target date format: %v", err)
	}

	return (tDate.Equal(startDate) || tDate.After(startDate)) && (tDate.Equal(endDate) || tDate.Before(endDate)), nil
}

// target time format HH:mm:ss
func IsTimeBetween(targetTime string, tTimeZone *time.Location, startTime, endTime time.Time) (bool, error) {
	// to get correct timezone offset if parsing time only, we must include the year
	// issue: https://github.com/golang/go/issues/34101#issuecomment-528260666
	tTimeFormat := "2006 15:04:00"
	_tTime, err := time.ParseInLocation(tTimeFormat, fmt.Sprintf("%s %s", "2025", targetTime), tTimeZone)
	if err != nil {
		return false, errors.Errorf("invalid target time format: %v", err)
	}

	localTz, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		return false, errors.New(err)
	}

	_sTime, err := time.ParseInLocation(time.TimeOnly, startTime.Format(time.TimeOnly), localTz)
	if err != nil {
		return false, errors.New(err)
	}

	_eTime, err := time.ParseInLocation(time.TimeOnly, endTime.Format(time.TimeOnly), localTz)
	if err != nil {
		return false, errors.New(err)
	}

	tTime := time.Date(2025, 0, 0, _tTime.Hour(), _tTime.Minute(), _tTime.Second(), 0, tTimeZone)
	sTime := time.Date(2025, 0, 0, _sTime.Hour(), _sTime.Minute(), _sTime.Second(), 0, localTz)
	eTime := time.Date(2025, 0, 0, _eTime.Hour(), _eTime.Minute(), _eTime.Second(), 0, localTz)

	return (tTime.Equal(sTime) || tTime.After(sTime)) && (tTime.Equal(eTime) || tTime.Before(eTime)), nil
}
