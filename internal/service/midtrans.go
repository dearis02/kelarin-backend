package service

import (
	"context"

	"github.com/go-errors/errors"
	"github.com/midtrans/midtrans-go/snap"
)

type Midtrans interface {
	CreateTransaction(ctx context.Context, req *snap.Request) (*snap.Response, error)
}

type midtransImpl struct {
	client *snap.Client
}

func NewMidtrans(client *snap.Client) Midtrans {
	return &midtransImpl{client}
}

func (s *midtransImpl) CreateTransaction(ctx context.Context, req *snap.Request) (*snap.Response, error) {
	res, err := s.client.CreateTransaction(req)
	if err != nil {
		return res, errors.New(err)
	}

	return res, nil
}
