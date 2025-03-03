package firebaseUtil

import (
	"context"
	"kelarin/internal/config"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
)

func NewApp(c *config.Config) *firebase.App {
	if _, err := os.Stat(c.FirebaseCredentialFile); os.IsNotExist(err) {
		log.Fatal().Msgf("Firebase credential file does not exist: %s", c.FirebaseCredentialFile)
	}

	credentialFile := option.WithCredentialsFile(c.FirebaseCredentialFile)
	app, err := firebase.NewApp(context.Background(), nil, credentialFile)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	return app
}

func NewMessagingClient(app *firebase.App) *messaging.Client {
	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	return client
}
