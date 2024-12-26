package routes

import (
	"kelarin/internal/handler"

	"github.com/gin-gonic/gin"
)

type User interface {
	Register()
}

type userImpl struct {
	g           *gin.Engine
	userHandler handler.User
}

func NewUser(g *gin.Engine, userHandler handler.User) User {
	return &userImpl{
		g:           g,
		userHandler: userHandler,
	}
}

func (u *userImpl) Register() {
	u.g.POST("/", u.userHandler.GetOne)
}
