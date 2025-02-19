package repository

import (
	"context"
	"encoding/json"
	"kelarin/internal/types"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	esTypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/operator"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/go-errors/errors"
)

type ServiceIndex interface {
	Index(ctx context.Context, req types.ServiceIndex) error
	FindByID(ctx context.Context, ID string) (types.ServiceIndex, error)
	Update(ctx context.Context, req types.ServiceIndex) error
	Delete(ctx context.Context, req types.ServiceIndex) error
	FindAllByFilter(ctx context.Context, req types.ServiceIndexFilter) ([]types.ServiceIndex, int64, *float64, error)
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

func (r *serviceIndexImpl) Delete(ctx context.Context, req types.ServiceIndex) error {
	_, err := r.esDB.Delete(types.ServiceElasticSearchIndexName, req.ID.String()).Do(ctx)
	if err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *serviceIndexImpl) FindAllByFilter(ctx context.Context, req types.ServiceIndexFilter) ([]types.ServiceIndex, int64, *float64, error) {
	res := []types.ServiceIndex{}
	var latestTimeStamp *float64

	searchReq := search.Request{
		Size: &req.Limit,
		Sort: []esTypes.SortCombinations{
			esTypes.SortOptions{
				SortOptions: map[string]esTypes.FieldSort{
					"created_at": {
						Order: &sortorder.Desc,
					},
				},
			},
		},
	}

	if req.LatestTimestamp.Valid {
		searchReq.SearchAfter = []esTypes.FieldValue{req.LatestTimestamp}
	}

	mustQuery := []esTypes.Query{}
	filterQuery := []esTypes.Query{}

	if req.Keyword != "" {
		mustQuery = append(mustQuery, esTypes.Query{
			MultiMatch: &esTypes.MultiMatchQuery{
				Query:    req.Keyword,
				Fields:   []string{"name", "description", "rules.name"},
				Operator: &operator.Or,
			},
		})
	}

	if len(req.Categories) > 0 {
		filterQuery = append(filterQuery, esTypes.Query{
			Terms: &esTypes.TermsQuery{
				TermsQuery: map[string]esTypes.TermsQueryField{
					"categories": req.Categories,
				},
			},
		})
	}

	if req.Province != "" {
		mustQuery = append(mustQuery, esTypes.Query{
			Match: map[string]esTypes.MatchQuery{
				"province": {
					Query:    req.Province,
					Operator: &operator.And,
				},
			},
		})
	}
	if req.City != "" {
		mustQuery = append(mustQuery, esTypes.Query{
			Match: map[string]esTypes.MatchQuery{
				"city": {
					Query:    req.City,
					Operator: &operator.And,
				},
			},
		})
	}

	searchReq.Query = &esTypes.Query{
		Bool: &esTypes.BoolQuery{
			Must:   mustQuery,
			Filter: filterQuery,
		},
	}

	services, err := r.esDB.Search().Index(types.ServiceElasticSearchIndexName).Request(&searchReq).Do(ctx)
	if err != nil {
		return res, 0, latestTimeStamp, errors.New(err)
	}

	for _, hit := range services.Hits.Hits {
		var service types.ServiceIndex
		if err := json.Unmarshal(hit.Source_, &service); err != nil {
			return res, 0, latestTimeStamp, errors.New(err)
		}

		res = append(res, service)
	}

	if len(services.Hits.Hits) > 0 {
		if latest, ok := services.Hits.Hits[len(services.Hits.Hits)-1].Sort[0].(float64); !ok {
			return res, 0, latestTimeStamp, errors.New("latest_timestamp is not float64")
		} else {
			latestTimeStamp = &latest
		}
	} else {
		latestTimeStamp = nil
	}

	return res, services.Hits.Total.Value, latestTimeStamp, nil
}
