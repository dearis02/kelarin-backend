package service

import (
	"context"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	"time"

	"firebase.google.com/go/messaging"
	"github.com/go-errors/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type Notification interface {
	SendPush(ctx context.Context, req types.NotificationSendReq) error
	SaveToken(ctx context.Context, req types.NotificationSaveTokenReq) error
}

type notificationImpl struct {
	db              *sqlx.DB
	messagingClient *messaging.Client
	fcmTokenRepo    repository.FCMToken
}

func NewNotification(db *sqlx.DB, messagingClient *messaging.Client, fcmTokenRepo repository.FCMToken) Notification {
	return &notificationImpl{
		db:              db,
		messagingClient: messagingClient,
		fcmTokenRepo:    fcmTokenRepo,
	}
}

func (s *notificationImpl) SendPush(ctx context.Context, req types.NotificationSendReq) error {
	_, err := s.messagingClient.Send(ctx, &messaging.Message{
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				Title: req.Title,
				Body:  req.Message,
				Icon:  req.IconURL,
				Badge: req.BadgeURL,
				Image: req.ImageURL,
			},
		},
		Token: req.Token,
	})
	if messaging.IsInvalidArgument(err) || messaging.IsRegistrationTokenNotRegistered(err) {
		log.Error().Stack().Err(errors.Errorf("invalid fcm token %s", req.Token))
		return nil
	} else if err != nil {
		return errors.New(err)
	}

	return nil
}

func (s *notificationImpl) SaveToken(ctx context.Context, req types.NotificationSaveTokenReq) error {
	if err := req.Validate(); err != nil {
		return err
	}

	key := types.FCMTokenKey(req.AuthUser.ID)
	token, err := s.fcmTokenRepo.Find(ctx, key)
	if !errors.Is(err, types.ErrNoData) && err != nil {
		return err
	}

	if token == req.Token {
		return nil
	}

	// 6480 hours = 9 months
	if err := s.fcmTokenRepo.Save(ctx, key, req.Token, time.Duration(time.Hour*6480)); err != nil {
		return err
	}

	return nil
}
