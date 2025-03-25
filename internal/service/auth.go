package service

import (
	"context"
	"fmt"
	"kelarin/internal/config"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	dbUtil "kelarin/internal/utils/dbutil"
	"net/http"
	"time"

	"github.com/go-errors/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

type Auth interface {
	LocalCreateSession(ctx context.Context, req types.AuthCreateSessionReq) (types.AuthCreateSessionRes, error)
	ConsumerCreateSession(ctx context.Context, req types.AuthCreateSessionForGoogleReq) (types.AuthCreateSessionForGoogleLoginRes, error)
	ProviderCreateSession(ctx context.Context, req types.AuthCreateSessionForGoogleReq) (types.AuthCreateSessionForGoogleLoginRes, error)
	RenewSession(ctx context.Context, req types.AuthRenewSessionReq) (types.AuthRenewSessionRes, error)
}

type authImpl struct {
	config                  *config.Config
	beginMainDBTx           dbUtil.SqlxTx
	sessionRepo             repository.Session
	userRepo                repository.User
	pendingRegistrationRepo repository.PendingRegistration
}

func NewAuth(cfg *config.Config, beginMainDBTx dbUtil.SqlxTx, sessionRepo repository.Session, userRepo repository.User, pendingRegistrationRepo repository.PendingRegistration) Auth {
	return &authImpl{
		config:                  cfg,
		beginMainDBTx:           beginMainDBTx,
		sessionRepo:             sessionRepo,
		userRepo:                userRepo,
		pendingRegistrationRepo: pendingRegistrationRepo,
	}
}

func (s *authImpl) LocalCreateSession(ctx context.Context, req types.AuthCreateSessionReq) (types.AuthCreateSessionRes, error) {
	res := types.AuthCreateSessionRes{}

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

func (s *authImpl) ConsumerCreateSession(ctx context.Context, req types.AuthCreateSessionForGoogleReq) (types.AuthCreateSessionForGoogleLoginRes, error) {
	res := types.AuthCreateSessionForGoogleLoginRes{}

	payload, err := s.ValidateGoogleIDToken(ctx, req.IDToken)
	if err != nil {
		return res, err
	}

	user, err := s.userRepo.FindByEmail(ctx, payload.Email)
	if errors.Is(err, types.ErrNoData) {
		id, err := uuid.NewV7()
		if err != nil {
			return res, errors.New(err)
		}

		user = types.User{
			ID:           id,
			Role:         types.UserRoleConsumer,
			Name:         payload.Name,
			Email:        payload.Email,
			AuthProvider: types.AuthProviderGoogle,
		}

		err = s.userRepo.Create(ctx, user)
		if err != nil {
			return res, err
		}
	} else if err != nil {
		return res, err
	}

	if user.Role != types.UserRoleConsumer {
		return res, errors.New(types.AppErr{Code: http.StatusUnauthorized, Message: "this account has been registered as a service provider, use another account"})
	}

	if user.IsSuspended {
		return res, errors.New(types.AppErr{Code: http.StatusUnauthorized, Message: "your account is suspended"})
	} else if user.IsBanned {
		return res, errors.New(types.AppErr{Code: http.StatusUnauthorized, Message: "your account is banned"})
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

	res = types.AuthCreateSessionForGoogleLoginRes{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		Role:         user.Role,
	}

	return res, nil
}

func (s *authImpl) ProviderCreateSession(ctx context.Context, req types.AuthCreateSessionForGoogleReq) (types.AuthCreateSessionForGoogleLoginRes, error) {
	res := types.AuthCreateSessionForGoogleLoginRes{}

	payload, err := s.ValidateGoogleIDToken(ctx, req.IDToken)
	if err != nil {
		return res, err
	}

	user, err := s.userRepo.FindByEmail(ctx, payload.Email)
	if errors.Is(err, types.ErrNoData) {
		id, err := uuid.NewV7()
		if err != nil {
			return res, errors.New(err)
		}

		user = types.User{
			ID:           id,
			Role:         types.UserRoleServiceProvider,
			Name:         payload.Name,
			Email:        payload.Email,
			AuthProvider: types.AuthProviderGoogle,
		}

		tx, err := s.beginMainDBTx(ctx, nil)
		if err != nil {
			return res, err
		}

		err = s.userRepo.CreateTx(ctx, tx, user)
		if err != nil {
			return res, err
		}

		key := types.GetPendingRegistrationKey(user.ID.String())
		if err = s.pendingRegistrationRepo.Set(ctx, key, user.ID); err != nil {
			return res, err
		}

		if err = tx.Commit(); err != nil {
			return res, errors.New(err)
		}
	} else if err != nil {
		return res, err
	}

	if user.Role != types.UserRoleServiceProvider {
		return res, errors.New(types.AppErr{Code: http.StatusUnauthorized, Message: "this account not registered as a provider, use another account"})
	}

	if user.IsSuspended {
		return res, errors.New(types.AppErr{Code: http.StatusUnauthorized, Message: "your account is suspended"})
	} else if user.IsBanned {
		return res, errors.New(types.AppErr{Code: http.StatusUnauthorized, Message: "your account is banned"})
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

	incompleteRegistration, err := s.pendingRegistrationRepo.IsExists(ctx, types.GetPendingRegistrationKey(user.ID.String()))
	if err != nil {
		return res, err
	}

	authUser := types.AuthUser{
		ID:                     user.ID,
		SessionID:              sessionId,
		Role:                   user.Role,
		Name:                   user.Name,
		IncompleteRegistration: &incompleteRegistration,
	}

	t, err := s.GenerateToken(authUser)
	if err != nil {
		return res, errors.New(err)
	}

	res = types.AuthCreateSessionForGoogleLoginRes{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		Role:         user.Role,
	}

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
		ID:                     authUser.SessionID,
		Subject:                authUser.ID,
		Role:                   authUser.Role,
		Name:                   authUser.Name,
		IncompleteRegistration: authUser.IncompleteRegistration,
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

func (s *authImpl) ValidateGoogleIDToken(ctx context.Context, idToken string) (types.AuthValidateGoogleIDToken, error) {
	res := types.AuthValidateGoogleIDToken{}

	payload, err := idtoken.Validate(ctx, idToken, s.config.Oauth.Google.ClientId)
	if err != nil {
		return res, errors.New(types.AppErr{Code: http.StatusUnauthorized, Message: "invalid google id token"})
	}

	res.Name = payload.Claims["name"].(string)
	res.Email = payload.Claims["email"].(string)

	return res, nil
}

func (s *authImpl) RenewSession(ctx context.Context, req types.AuthRenewSessionReq) (types.AuthRenewSessionRes, error) {
	res := types.AuthRenewSessionRes{}

	if err := req.Validate(); err != nil {
		return res, err
	}

	claims := &types.AuthJwtCustomClaims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWT.RefreshTokenSecretKey), nil
	})
	if errors.Is(err, jwt.ErrTokenExpired) {
		return res, types.AuthErrTokenExpired
	} else if err != nil {
		return res, types.AuthErrInvalidToken
	}

	if !token.Valid {
		return res, types.AuthErrInvalidToken
	}

	claims, ok := token.Claims.(*types.AuthJwtCustomClaims)
	if !ok {
		return res, types.AuthErrInvalidTokenClaims
	}

	sessionKey := types.GetSessionKey(claims.ID.String())
	userId, err := s.sessionRepo.Find(ctx, sessionKey)
	if errors.Is(err, types.ErrNoData) {
		return res, types.AuthErrSessionRevoked
	} else if err != nil {
		return res, err
	}

	uuidUserID, err := uuid.Parse(userId)
	if err != nil {
		return res, errors.New(err)
	}

	user, err := s.userRepo.FindByID(ctx, uuidUserID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(types.AppErr{Code: http.StatusUnauthorized, Message: "user not found", Err: fmt.Errorf("user not found : id %s", userId)})
	} else if err != nil {
		return res, err
	}

	newSessionID, err := uuid.NewV7()
	if err != nil {
		return res, errors.New(err)
	}

	newSessionKey := types.GetSessionKey(newSessionID.String())
	if err = s.sessionRepo.RenewAndDelete(ctx, sessionKey, newSessionKey, user.ID.String(), s.config.JWT.RefreshTokenExpiration); err != nil {
		return res, err
	}

	incompleteRegistration, err := s.pendingRegistrationRepo.IsExists(ctx, types.GetPendingRegistrationKey(userId))
	if err != nil {
		return res, err
	}

	authUser := types.AuthUser{
		ID:        claims.Subject,
		SessionID: newSessionID,
		Role:      claims.Role,
		Name:      claims.Name,
	}

	if incompleteRegistration {
		authUser.IncompleteRegistration = &incompleteRegistration
	}

	t, err := s.GenerateToken(authUser)
	if err != nil {
		return res, errors.New(err)
	}

	res.AccessToken = t.AccessToken
	res.RefreshToken = t.RefreshToken

	return res, nil
}
