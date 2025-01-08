package fileSystemUtil

import (
	"kelarin/internal/types"
	"os"

	"github.com/go-errors/errors"
)

func InitTempDir() error {
	tempDir := types.TempFileDir
	_, err := os.Stat(tempDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(tempDir, os.ModePerm)
		if err != nil {
			return errors.New(err)
		}
	} else if err != nil {
		return errors.New(err)
	}

	return nil
}
