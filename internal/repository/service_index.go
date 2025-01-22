package repository

import (
	"context"
	"kelarin/internal/types"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-errors/errors"
)

type ServiceIndex interface {
	Index(ctx context.Context, req types.ServiceIndex) error
}

type serviceIndexImpl struct {
	esDB *elasticsearch.TypedClient
}

func NewServiceIndex(esDB *elasticsearch.TypedClient) ServiceIndex {
	return &serviceIndexImpl{
		esDB: esDB,
	}
}

func (r *serviceIndexImpl) Index(ctx context.Context, req types.ServiceIndex) error {
	_, err := r.esDB.Index(types.ServiceElasticSearchIndexName).Request(req).Do(ctx)
	if err != nil {
		return errors.New(err)
	}

	return nil
}
