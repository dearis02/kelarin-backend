package routes

import (
	"kelarin/internal/handler"

	"github.com/gin-gonic/gin"
)

type User interface {
	Register()
}

type newUserDeps struct {
	g           *gin.Engine
	userHandler handler.User
}

func NewUser(g *gin.Engine, userHandler handler.User) User {
	return &newUserDeps{
		g:           g,
		userHandler: userHandler,
	}
}

func (u *newUserDeps) Register() {
	u.g.POST("/", u.userHandler.GetOne)
}
