package repository

import (
	"context"
	"encoding/json"
	"kelarin/internal/types"
	"net/http"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	esTypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/fieldsortnumerictype"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/operator"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/go-errors/errors"
)

type ServiceIndex interface {
	Create(ctx context.Context, req types.ServiceIndex) error
	FindByID(ctx context.Context, ID string) (types.ServiceIndex, int64, int64, error)
	Update(ctx context.Context, req types.ServiceIndex, seqNo int64, primaryTerm int64) error
	Delete(ctx context.Context, req types.ServiceIndex) error
	FindAllByFilter(ctx context.Context, req types.ServiceIndexFilter) ([]types.ServiceIndex, int64, []esTypes.FieldValue, error)
}

type serviceIndexImpl struct {
	esDB *elasticsearch.TypedClient
}

func NewServiceIndex(esDB *elasticsearch.TypedClient) ServiceIndex {
	return &serviceIndexImpl{
		esDB: esDB,
	}
}

func (r *serviceIndexImpl) Create(ctx context.Context, req types.ServiceIndex) error {
	_, err := r.esDB.Index(types.ServiceElasticSearchIndexName).Request(req).Id(req.ID.String()).Do(ctx)
	if err != nil {
		return errors.New(err)
	}

	return nil
}

func (r *serviceIndexImpl) FindByID(ctx context.Context, ID string) (types.ServiceIndex, int64, int64, error) {
	res := types.ServiceIndex{}
	seqNo := int64(0)
	primaryTerm := int64(0)

	svc, err := r.esDB.Get(types.ServiceElasticSearchIndexName, ID).Do(ctx)
	if err != nil {
		return res, seqNo, primaryTerm, errors.New(err)
	}

	if !svc.Found {
		return res, seqNo, primaryTerm, types.ErrNoData
	}

	if err := json.Unmarshal(svc.Source_, &res); err != nil {
		return res, seqNo, primaryTerm, errors.New(err)
	}

	if svc.SeqNo_ != nil {
		seqNo = *svc.SeqNo_
	}
	if svc.PrimaryTerm_ != nil {
		primaryTerm = *svc.PrimaryTerm_
	}

	return res, seqNo, primaryTerm, nil
}

func (r *serviceIndexImpl) Update(ctx context.Context, req types.ServiceIndex, seqNo int64, primaryTerm int64) error {
	seqNoStr := strconv.FormatInt(seqNo, 10)
	primaryTermStr := strconv.FormatInt(primaryTerm, 10)

	updateReq := r.esDB.Update(types.ServiceElasticSearchIndexName, req.ID.String()).
		Doc(req).
		IfSeqNo(seqNoStr).
		IfPrimaryTerm(primaryTermStr)

	esErr := esTypes.NewElasticsearchError()
	_, err := updateReq.Do(ctx)
	if errors.Is(err, esErr) {
		esErr = err.(*esTypes.ElasticsearchError)
		if esErr.Status == http.StatusConflict {
			return nil
		} else {
			return errors.New(err)
		}
	} else if err != nil {
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

func (r *serviceIndexImpl) FindAllByFilter(ctx context.Context, req types.ServiceIndexFilter) ([]types.ServiceIndex, int64, []esTypes.FieldValue, error) {
	res := []types.ServiceIndex{}
	var after []esTypes.FieldValue

	dateTimeFormat := "strict_date_optional_time_nanos"

	searchReq := search.Request{
		Size: &req.Limit,
		Sort: []esTypes.SortCombinations{
			esTypes.SortOptions{
				Score_: &esTypes.ScoreSort{
					Order: &sortorder.Desc,
				},
				SortOptions: map[string]esTypes.FieldSort{
					"created_at": {
						NumericType: &fieldsortnumerictype.Date,
						Format:      &dateTimeFormat,
						Order:       &sortorder.Desc,
					},
				},
			},
		},
	}

	if len(req.After) > 0 {
		searchReq.SearchAfter = req.After
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
		return res, 0, nil, errors.New(err)
	}

	for _, hit := range services.Hits.Hits {
		var service types.ServiceIndex
		if err := json.Unmarshal(hit.Source_, &service); err != nil {
			return res, 0, nil, errors.New(err)
		}

		res = append(res, service)
	}

	if len(services.Hits.Hits) == req.Limit {
		after = services.Hits.Hits[len(services.Hits.Hits)-1].Sort
	}

	return res, services.Hits.Total.Value, after, nil
}
