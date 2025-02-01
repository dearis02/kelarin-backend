package pkg

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func GenerateUniqueFileName(fileName string) string {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}

	timeStamp := time.Now().Unix()
	name := fmt.Sprintf("%d_%s", timeStamp, hex.EncodeToString(bytes))

	return name + filepath.Ext(fileName)
}

func IsDirExist(path string) (bool, error) {
	stat, err := os.Stat(filepath.Dir(path))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return stat.IsDir(), nil
}
