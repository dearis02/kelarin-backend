package service

import (
	"context"
	"kelarin/internal/repository"
	"kelarin/internal/types"
)

type PaymentMethod interface {
	GetAll(ctx context.Context) ([]types.PaymentMethodGetAllRes, error)
}

type paymentMethodImpl struct {
	paymentMethodRepo repository.PaymentMethod
}

func NewPaymentMethod(paymentMethodRepo repository.PaymentMethod) PaymentMethod {
	return &paymentMethodImpl{paymentMethodRepo: paymentMethodRepo}
}

func (s *paymentMethodImpl) GetAll(ctx context.Context) ([]types.PaymentMethodGetAllRes, error) {
	res := []types.PaymentMethodGetAllRes{}

	paymentMethods, err := s.paymentMethodRepo.FindAll(ctx)
	if err != nil {
		return res, err
	}

	for _, p := range paymentMethods {
		res = append(res, types.PaymentMethodGetAllRes{
			ID:           p.ID,
			Name:         p.Name,
			Type:         p.Type,
			AdminFee:     p.AdminFee,
			AdminFeeUnit: p.AdminFeeUnit,
			LogoURL:      p.Logo,
			Enabled:      p.Enabled,
		})
	}

	return res, nil
}
