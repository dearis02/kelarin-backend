package utils

import (
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

func NewMidtransSnapClient(serverKey string, env midtrans.EnvironmentType, notificationURL string) *snap.Client {
	client := snap.Client{}
	client.New(serverKey, env)
	client.Options.PaymentOverrideNotification = &notificationURL

	if env == midtrans.Production {
		midtrans.DefaultLoggerLevel = &midtrans.LoggerImplementation{LogLevel: midtrans.LogError}
	}

	return &client
}
