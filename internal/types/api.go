package types

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-errors/errors"
)

type ApiResponse struct {
	StatusCode int            `json:"status_code"`
	Message    string         `json:"message,omitempty"`
	Data       interface{}    `json:"data,omitempty"`
	Errors     []any          `json:"errors,omitempty"`
	Pagination *PaginationRes `json:"pagination,omitempty"`
	Metadata   any            `json:"metadata,omitempty"`
}

func (r ApiResponse) MarshalJSON() ([]byte, error) {
	if r.Message == "" {
		r.Message = http.StatusText(r.StatusCode)
	}

	type Alias ApiResponse
	return json.Marshal(Alias(r))
}

type PaginationReq struct {
	Page string
	Size string
}

type PaginationRes struct {
	Page      int32 `json:"page"`
	Size      int32 `json:"size"`
	TotalItem int64 `json:"total_item"`
	TotalPage int64 `json:"total_page"`
}

func (r *PaginationReq) ValidateAndNormalize() error {
	if r.Page == "" {
		r.Page = "1"
	}
	if r.Size == "" {
		r.Size = "30"
	}

	pageReq, err := strconv.Atoi(r.Page)
	if err != nil {
		return errors.New(AppErr{
			Code:    http.StatusBadRequest,
			Message: "invalid page query",
		})
	}

	sizeReq, err := strconv.Atoi(r.Size)
	if err != nil {
		return errors.New(AppErr{
			Code:    http.StatusBadRequest,
			Message: "invalid size query",
		})
	}

	if pageReq < 1 {
		r.Page = "1"
	}
	if sizeReq < 1 {
		r.Size = "10"
	}

	return nil
}

func (r PaginationReq) GeneratePaginationResponse(totalItem int64) PaginationRes {
	page, _ := strconv.Atoi(r.Page)
	size, _ := strconv.Atoi(r.Size)

	totalPage := totalItem / int64(size)
	if totalItem%int64(size) > 0 {
		totalPage++
	}

	return PaginationRes{
		Page:      int32(page),
		Size:      int32(size),
		TotalItem: totalItem,
		TotalPage: totalPage,
	}
}
