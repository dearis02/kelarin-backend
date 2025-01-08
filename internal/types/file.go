package types

import (
	"mime/multipart"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const TempFileDir = "temp"

// region repo types

type FileTemp struct {
	Name       string
	Expiration time.Duration
}

type TempFile struct {
	Name string `redis:"name"`
}

// end of region repo types

// region service types

type FileUploadImagesReq struct {
	Files []*multipart.FileHeader
}

type FileUploadFilesRes struct {
	FileKeys []string `json:"file_keys"`
}

type FileGetTempRes TempFile

type FilePresignedURLClaims struct {
	jwt.RegisteredClaims
	FileName string `json:"file_name"`
	Dir      string `json:"dir"`
}

type FileServeReq struct {
	Filename string `param:"filename"`
	Token    string `query:"token"`
}

// end of region service types

const (
	RedisStreamTempFiles = "temp_files"
)

func GetUploadedFileKey(ID string) string {
	return "file:" + ID
}
