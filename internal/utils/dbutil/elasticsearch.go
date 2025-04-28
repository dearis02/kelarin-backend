package dbUtil

import (
	"kelarin/internal/config"
	"net/http"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

func NewElasticsearchClient(cfg config.ElasticsearchConfig) (*elasticsearch.TypedClient, error) {
	cert, err := os.ReadFile(cfg.SSLCertificate)
	if err != nil {
		return nil, err
	}

	c := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		CACert:    cert,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.MaxIdleCons,
			MaxIdleConnsPerHost: cfg.MaxIdleConsPerHost,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	client, err := elasticsearch.NewTypedClient(c)
	if err != nil {
		return nil, err
	}

	return client, nil
}
