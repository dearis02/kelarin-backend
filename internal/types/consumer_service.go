package types

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v9"
)

// region service types

type ConsumerServiceGetAllReq struct {
	Province        string
	City            string
	Categories      []string
	Keyword         string
	LatestTimestamp null.Time
	PaginationReq
}

type ConsumerServiceGetAllRes struct {
	ID                    uuid.UUID       `json:"id"`
	Name                  string          `json:"name"`
	ImageURL              string          `json:"image_url"`
	FeeStartAt            decimal.Decimal `json:"fee_start_at"`
	FeeEndAt              decimal.Decimal `json:"fee_end_at"`
	Address               string          `json:"address"`
	Province              string          `json:"province"`
	City                  string          `json:"city"`
	ReceivedRatingCount   int32           `json:"received_rating_count"`
	ReceivedRatingAverage float32         `json:"received_rating_average"`
}

type ConsumerServiceGetAllMetadata struct {
	LatestTimestamp *float64 `json:"latest_timestamp"`
}

// end of region service types
