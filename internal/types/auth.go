package types

import (
	"net/http"

	"github.com/go-errors/errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AuthUserContextKey = "user"
)

var (
	AuthErrTokenExpired       = errors.New(AppErr{Code: http.StatusUnauthorized, Message: "token expired"})
	AuthErrInvalidToken       = errors.New(AppErr{Code: http.StatusUnauthorized, Message: "invalid token"})
	AuthErrInvalidTokenClaims = errors.New(AppErr{Code: http.StatusUnauthorized, Message: "invalid token claims"})
	AuthErrSessionRevoked     = errors.New(AppErr{Code: http.StatusUnauthorized, Message: "session revoked"})
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
	ID                     uuid.UUID `json:"jti"`
	Subject                uuid.UUID `json:"sub"`
	Role                   UserRole  `json:"role"`
	Name                   string    `json:"name"`
	IncompleteRegistration *bool     `json:"incomplete_registration,omitempty"`
}

type AuthCreateSessionReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r AuthCreateSessionReq) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email, validation.Required, is.Email),
		validation.Field(&r.Password, validation.Required),
	)
}

type AuthCreateSessionRes struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthCreateSessionForGoogleLoginRes struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Role         UserRole `json:"role"`
}

type AuthUser struct {
	ID                     uuid.UUID
	SessionID              uuid.UUID
	Role                   UserRole
	Name                   string
	IncompleteRegistration *bool // for service provider role
}

func (r AuthUser) IsZero() bool {
	return r == (AuthUser{})
}

type AuthGenerateToken struct {
	AccessToken  string
	RefreshToken string
}

type AuthCreateSessionForGoogleReq struct {
	IDToken string `json:"id_token"`
}

type AuthValidateGoogleIDToken struct {
	Name  string
	Email string
}

type AuthHeaderReq struct {
	Authorization string `header:"Authorization"`
}

type AuthRenewSessionReq struct {
	RefreshToken string `json:"refresh_token"`
}

func (r AuthRenewSessionReq) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.RefreshToken, validation.Required),
	)
}

type AuthRenewSessionRes AuthCreateSessionRes

type AuthRevokeSessionReq struct {
	AuthUser AuthUser `middleware:"user"`
}

func (r AuthRevokeSessionReq) Validate() error {
	if r.AuthUser.IsZero() {
		return errors.New("AuthUser is required")
	}

	return nil
}

// end of region service types
