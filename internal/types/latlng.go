package types

import (
	"fmt"
	"strings"

	"github.com/go-errors/errors"
	"github.com/golang/geo/s2"
)

type LatLng s2.LatLng

func (l *LatLng) Scan(value interface{}) error {
	if value == nil {
		return errors.New("value is nil")
	}

	strVal, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string but got %T", value)
	}

	strVal = strings.TrimPrefix(strVal, "POINT(")
	strVal = strings.TrimSuffix(strVal, ")")
	strVal = strings.TrimSpace(strVal)

	coords := strings.Split(strVal, " ")
	if len(coords) != 2 {
		return fmt.Errorf("invalid geography format: %s", strVal)
	}

	var lng, lat float64
	_, err := fmt.Sscanf(coords[0], "%f", &lng)
	if err != nil {
		return fmt.Errorf("error parsing longitude: %w", err)
	}
	_, err = fmt.Sscanf(coords[1], "%f", &lat)
	if err != nil {
		return fmt.Errorf("error parsing latitude: %w", err)
	}

	// Set the values to the s2.LatLng object
	*l = LatLng(s2.LatLngFromDegrees(lat, lng))
	return nil
}

func (l LatLng) Value() (interface{}, error) {
	return s2.LatLng(l), nil
}
