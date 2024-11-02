package pkg

import (
	"encoding/base64"

	"github.com/go-errors/errors"
)

func ValidateBase64(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.New("must be a string")
	}

	_, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return errors.New("must be a valid base64 string")
	}

	return nil
}
