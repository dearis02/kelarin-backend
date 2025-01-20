package dbUtil

import (
	"kelarin/internal/config"
	"net/http"
	"os"
	"time"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/rs/zerolog/log"
)

func NewElasticsearchClient(cfg *config.Config) (*elasticsearch.Client, error) {
	cert, err := os.ReadFile(cfg.Elasticsearch.CertificatePath)
	if err != nil {
		return nil, err
	}

	c := elasticsearch.Config{
		Addresses: cfg.Elasticsearch.Addresses,
		APIKey:    cfg.Elasticsearch.APIKey,
		CACert:    cert,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.Elasticsearch.MaxIdleCons,
			MaxIdleConnsPerHost: cfg.Elasticsearch.MaxIdleConsPerHost,
			IdleConnTimeout:     90 * time.Second,
		},
		ConnectionPoolFunc: func(c []*elastictransport.Connection, s elastictransport.Selector) elastictransport.ConnectionPool {
			con, err := elastictransport.NewConnectionPool(c, s)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to create connection pool")
			}

			return con
		},
	}

	client, err := elasticsearch.NewClient(c)
	if err != nil {
		return nil, err
	}

	return client, nil
}
