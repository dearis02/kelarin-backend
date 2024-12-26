package types

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// region repo types

type AuthProvider int16

const (
	AuthProviderLocal AuthProvider = iota + 1
	AuthProviderGoogle
)

// end of region repo types

// region service types

type AuthJwtCustomClaims struct {
	jwt.RegisteredClaims
	ID      uuid.UUID `json:"jti"`
	Subject uuid.UUID `json:"sub"`
	Role    UserRole  `json:"role"`
	Name    string    `json:"name"`
}

type AuthLoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r AuthLoginReq) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email, validation.Required, is.Email),
		validation.Field(&r.Password, validation.Required),
	)
}

type AuthLoginRes struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthUser struct {
	ID        uuid.UUID
	SessionID uuid.UUID
	Role      UserRole
	Name      string
}

type AuthGenerateToken struct {
	AccessToken  string
	RefreshToken string
}

// end of region service types
