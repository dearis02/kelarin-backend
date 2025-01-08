package routes

import (
	"kelarin/internal/handler"

	"github.com/gin-gonic/gin"
)

type File interface {
	Register()
}

type fileImpl struct {
	g           *gin.Engine
	fileHandler handler.File
}

func NewFile(g *gin.Engine, fileHandler handler.File) File {
	return &fileImpl{
		g:           g,
		fileHandler: fileHandler,
	}
}

func (u *fileImpl) Register() {
	u.g.POST("/common/v1/files/_images", u.fileHandler.UploadImages)
}
