package middleware

import (
	"kelarin/internal/config"
	"kelarin/internal/types"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

type Auth interface {
	Authenticated(c *gin.Context)
	Consumer(c *gin.Context)
	ServiceProvider(c *gin.Context)
}

type authImpl struct {
	config *config.Config
}

func NewAuth(config *config.Config) Auth {
	return &authImpl{config: config}
}

func (m *authImpl) Authenticated(c *gin.Context) {
	m.parseAuthorizationHeader(c)
	c.Next()
}

func (m *authImpl) Admin(c *gin.Context) {
	m.parseAuthorizationHeader(c)
	m.nextFunc(c, types.UserRoleAdmin)
}

func (m *authImpl) Consumer(c *gin.Context) {
	m.parseAuthorizationHeader(c)
	m.nextFunc(c, types.UserRoleConsumer)
}

func (m *authImpl) ServiceProvider(c *gin.Context) {
	m.parseAuthorizationHeader(c)
	m.nextFunc(c, types.UserRoleServiceProvider)
}

func (m *authImpl) parseAuthorizationHeader(c *gin.Context) {
	claims, err := m.parseJwt(c)
	if err != nil {
		log.Error().Err(err).Send()
		c.Error(errors.New(types.AppErr{Code: http.StatusUnauthorized}))
		c.Abort()
		return
	}

	c.Set(types.AuthUserContextKey, claims)
	c.Next()
}

func (m *authImpl) nextFunc(c *gin.Context, role types.UserRole) {
	userContext, exists := c.Get(types.AuthUserContextKey)
	if !exists {
		log.Error().Msg("missing user context")
		c.Error(errors.New(types.AppErr{Code: http.StatusUnauthorized}))
		c.Abort()
		return
	}

	authUser, ok := userContext.(*types.AuthJwtCustomClaims)
	if !ok {
		log.Error().Msg("invalid user context")
		c.Error(errors.New(types.AppErr{Code: http.StatusUnauthorized}))
		c.Abort()
		return
	}

	if authUser.Role != role {
		log.Error().Msg("invalid user role")
		c.Error(errors.New(types.AppErr{Code: http.StatusForbidden}))
		c.Abort()
		return
	}

	c.Next()
}

func getTokenFromHeader(c *gin.Context) (string, error) {
	req := types.AuthHeaderReq{}

	if err := c.ShouldBindHeader(&req); err != nil {
		return "", err
	}

	authHeader := strings.Split(req.Authorization, " ")
	if authHeader[0] != "Bearer" {
		return "", errors.New("invalid token type, must be Bearer")
	} else if len(authHeader) != 2 {
		return "", errors.New("invalid token format")
	} else if authHeader[1] == "" {
		return "", errors.New("missing token")
	}

	return authHeader[1], nil
}

func (m *authImpl) parseJwt(c *gin.Context) (*types.AuthJwtCustomClaims, error) {
	claims := &types.AuthJwtCustomClaims{}

	accToken, err := getTokenFromHeader(c)
	if err != nil {
		return claims, err
	}

	token, err := jwt.ParseWithClaims(accToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.config.JWT.SecretKey), nil
	})
	if err != nil {
		return claims, errors.New(err)
	}

	if !token.Valid {
		return claims, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*types.AuthJwtCustomClaims)
	if !ok {
		return claims, errors.New("invalid token claims")
	}

	return claims, nil
}
