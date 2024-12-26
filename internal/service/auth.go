package service

import (
	"context"
	"kelarin/internal/config"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	"net/http"
	"time"

	"github.com/go-errors/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Auth interface {
	CreateSession(ctx context.Context, req types.AuthLoginReq) (types.AuthLoginRes, error)
}

type authImpl struct {
	config      *config.Config
	sessionRepo repository.Session
	userRepo    repository.User
}

func NewAuth(cfg *config.Config, sessionRepo repository.Session, userRepo repository.User) Auth {
	return &authImpl{
		config:      cfg,
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
	}
}

func (s *authImpl) CreateSession(ctx context.Context, req types.AuthLoginReq) (types.AuthLoginRes, error) {
	res := types.AuthLoginRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(types.AppErr{Code: http.StatusUnauthorized, Message: "invalid email or password"})
	} else if err != nil {
		return res, err
	}

	if user.IsSuspended {
		return res, errors.New(types.AppErr{Code: http.StatusUnauthorized, Message: "your account is suspended"})
	} else if user.IsBanned {
		return res, errors.New(types.AppErr{Code: http.StatusUnauthorized, Message: "your account is banned"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte(req.Password)); err != nil {
		return res, errors.New(types.AppErr{
			Code:    http.StatusUnauthorized,
			Message: "invalid email or password",
		})
	}

	sessionId, err := uuid.NewV7()
	if err != nil {
		return res, errors.New(err)
	}

	sessionKey := types.GetSessionKey(sessionId.String())
	err = s.sessionRepo.Set(ctx, sessionKey, user.ID.String(), s.config.JWT.RefreshTokenExpiration)
	if err != nil {
		return res, err
	}

	authUser := types.AuthUser{
		ID:        user.ID,
		SessionID: sessionId,
		Role:      user.Role,
		Name:      user.Name,
	}

	t, err := s.GenerateToken(authUser)
	if err != nil {
		return res, errors.New(err)
	}

	res.AccessToken = t.AccessToken
	res.RefreshToken = t.RefreshToken

	return res, nil
}

func (s *authImpl) GenerateToken(authUser types.AuthUser) (types.AuthGenerateToken, error) {
	res := types.AuthGenerateToken{}

	accToken := jwt.NewWithClaims(jwt.SigningMethodHS256, types.AuthJwtCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.JWT.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.JWT.Expiration)),
		},
		ID:      authUser.SessionID,
		Subject: authUser.ID,
		Role:    authUser.Role,
		Name:    authUser.Name,
	})

	signedAccToken, err := accToken.SignedString([]byte(s.config.JWT.SecretKey))
	if err != nil {
		return res, err
	}

	res.AccessToken = signedAccToken

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, types.AuthJwtCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.JWT.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.JWT.RefreshTokenExpiration)),
		},
		ID:      authUser.SessionID,
		Subject: authUser.ID,
		Role:    authUser.Role,
		Name:    authUser.Name,
	})

	signedRefreshToken, err := refreshToken.SignedString([]byte(s.config.JWT.RefreshTokenSecretKey))
	if err != nil {
		return res, err
	}

	res.RefreshToken = signedRefreshToken

	return res, nil
}
