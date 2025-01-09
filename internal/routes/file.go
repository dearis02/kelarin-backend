package routes

import (
	"kelarin/internal/handler"
	"kelarin/internal/middleware"

	"github.com/gin-gonic/gin"
)

type File interface {
	Register(m middleware.Auth)
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

func (u *fileImpl) Register(m middleware.Auth) {
	u.g.POST("/common/v1/files/_images", m.Authenticated, u.fileHandler.UploadImages)
}
