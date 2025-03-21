package middleware

import (
	"kelarin/internal/config"
	"kelarin/internal/types"
	"net/http"
	"reflect"
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
	BindWithRequest(c *gin.Context, req any) error
	WS(c *gin.Context)
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
	accToken, err := getTokenFromHeader(c)
	if err != nil {
		log.Error().Stack().Err(err).Send()
		c.Error(errors.New(types.AppErr{Code: http.StatusUnauthorized}))
		c.Abort()
		return
	}

	claims, err := m.parseJwt(accToken)
	if err != nil {
		log.Error().Stack().Err(err).Send()
		c.Error(errors.New(types.AppErr{Code: http.StatusUnauthorized}))
		c.Abort()
		return
	}

	authUser := types.AuthUser{
		ID:                     claims.Subject,
		SessionID:              claims.ID,
		Role:                   claims.Role,
		Name:                   claims.Name,
		IncompleteRegistration: claims.IncompleteRegistration,
	}

	c.Set(types.AuthUserContextKey, authUser)
}

func (m *authImpl) nextFunc(c *gin.Context, role types.UserRole) {
	userContext, exists := c.Get(types.AuthUserContextKey)
	if !exists {
		log.Error().Stack().Err(errors.New("missing user context")).Send()
		c.Error(errors.New(types.AppErr{Code: http.StatusUnauthorized}))
		c.Abort()
		return
	}

	authUser, ok := userContext.(types.AuthUser)
	if !ok {
		log.Error().Stack().Err(errors.New("invalid user context")).Send()
		c.Error(errors.New(types.AppErr{Code: http.StatusUnauthorized}))
		c.Abort()
		return
	}

	if authUser.Role != role {
		log.Error().Stack().Err(errors.New("invalid user role")).Send()
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

func (m *authImpl) parseJwt(accToken string) (*types.AuthJwtCustomClaims, error) {
	claims := &types.AuthJwtCustomClaims{}

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

// BindWithRequest binds the request body to the req struct and also binds the AuthUser to the struct field with the middleware tag
func (authImpl) BindWithRequest(c *gin.Context, req any) error {
	reqValue := reflect.ValueOf(req)
	if reqValue.Kind() != reflect.Ptr || reqValue.Elem().Kind() != reflect.Struct {
		return errors.New("req must be a pointer to a struct")
	}

	if err := c.ShouldBind(req); err != nil {
		return errors.New(types.AppErr{Code: http.StatusBadRequest, Message: err.Error()})
	}

	if err := c.ShouldBindHeader(req); err != nil {
		return err
	}

	if err := c.ShouldBindQuery(req); err != nil {
		return err
	}

	reqType := reqValue.Elem().Type()
	for i := range reqType.NumField() {
		field := reqType.Field(i)
		tag := field.Tag.Get("middleware")
		if tag == "user" {
			authUser, exists := c.Get("user")
			if !exists {
				return errors.New("auth user not found in context, missing auth middleware on registered routes")
			}

			fieldValue := reqValue.Elem().Field(i)
			if fieldValue.CanSet() && reflect.TypeOf(authUser).AssignableTo(fieldValue.Type()) {
				fieldValue.Set(reflect.ValueOf(authUser))
			} else {
				return errors.New("auth user type mismatch")
			}
		}
	}

	return nil
}

func (m *authImpl) WS(c *gin.Context) {
	token, exist := c.GetQuery("token")
	if !exist {
		c.JSON(http.StatusUnauthorized, types.ApiResponse{StatusCode: http.StatusUnauthorized})
		c.Abort()
		return
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, types.ApiResponse{StatusCode: http.StatusUnauthorized})
		c.Abort()
		return
	}

	claims, err := m.parseJwt(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.ApiResponse{StatusCode: http.StatusUnauthorized})
		c.Abort()
		return
	}

	authUser := types.AuthUser{
		ID:                     claims.Subject,
		SessionID:              claims.ID,
		Role:                   claims.Role,
		Name:                   claims.Name,
		IncompleteRegistration: claims.IncompleteRegistration,
	}

	c.Set(types.AuthUserContextKey, authUser)
}
