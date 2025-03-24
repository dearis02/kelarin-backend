package dbUtil

import (
	"kelarin/internal/config"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

func NewElasticsearchClient(cfg *config.ElasticsearchConfig) (*elasticsearch.TypedClient, error) {
	c := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
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
