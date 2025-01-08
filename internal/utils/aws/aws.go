package awsUtil

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

func NewConfig() aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	return cfg
}

func NewS3ClientFromConfig(cfg aws.Config) *s3.Client {
	return s3.NewFromConfig(cfg)
}

func NewS3PresignClient(s3Client *s3.Client) *s3.PresignClient {
	return s3.NewPresignClient(s3Client)
}
