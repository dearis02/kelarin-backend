package utils

import (
	"encoding/hex"

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
