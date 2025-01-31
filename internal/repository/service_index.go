package repository

import (
	"context"
	"encoding/json"
	"kelarin/internal/types"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-errors/errors"
)

type ServiceIndex interface {
	Index(ctx context.Context, req types.ServiceIndex) error
	FindByID(ctx context.Context, ID string) (types.ServiceIndex, error)
	Update(ctx context.Context, req types.ServiceIndex) error
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
	_, err := r.esDB.Index(types.ServiceElasticSearchIndexName).Request(req).Id(req.ID.String()).Do(ctx)
	if err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *serviceIndexImpl) FindByID(ctx context.Context, ID string) (types.ServiceIndex, error) {
	res := types.ServiceIndex{}

	svc, err := r.esDB.Get(types.ServiceElasticSearchIndexName, ID).Do(ctx)
	if err != nil {
		return res, errors.New(err)
	}

	if !svc.Found {
		return res, types.ErrNoData
	}

	if err := json.Unmarshal(svc.Source_, &res); err != nil {
		return res, errors.New(err)
	}

	return res, nil
}

func (r *serviceIndexImpl) Update(ctx context.Context, req types.ServiceIndex) error {
	if _, err := r.esDB.Update(types.ServiceElasticSearchIndexName, req.ID.String()).Doc(req).Do(ctx); err != nil {
		return errors.New(err)
	}

	return nil
}
