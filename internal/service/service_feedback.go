package service

import (
	"context"
	"kelarin/internal/repository"
	"kelarin/internal/types"
	"net/http"

	"github.com/go-errors/errors"
)

type ServiceFeedback interface {
	GetAllByServiceID(ctx context.Context, req types.ServiceFeedbackGetAllByServiceIDReq) ([]types.ServiceFeedbackGetAllByServiceIDRes, error)
}

type serviceFeedbackImpl struct {
	serviceFeedbackRepo repository.ServiceFeedback
	serviceRepo         repository.Service
}

func NewServiceFeedback(
	serviceFeedbackRepo repository.ServiceFeedback,
	serviceRepo repository.Service,
) ServiceFeedback {
	return &serviceFeedbackImpl{
		serviceFeedbackRepo: serviceFeedbackRepo,
		serviceRepo:         serviceRepo,
	}
}

func (s *serviceFeedbackImpl) GetAllByServiceID(ctx context.Context, req types.ServiceFeedbackGetAllByServiceIDReq) ([]types.ServiceFeedbackGetAllByServiceIDRes, error) {
	res := []types.ServiceFeedbackGetAllByServiceIDRes{}

	err := req.Validate()
	if err != nil {
		return res, err
	}

	service, err := s.serviceRepo.FindByID(ctx, req.ID)
	if errors.Is(err, types.ErrNoData) {
		return res, errors.New(types.AppErr{
			Code:    http.StatusNotFound,
			Message: "Service not found",
		})
	} else if err != nil {
		return res, err
	}

	feedbacks, err := s.serviceFeedbackRepo.FindByServiceIDWithUser(ctx, service.ID)
	if err != nil {
		return res, err
	}

	for _, feedback := range feedbacks {
		res = append(res, types.ServiceFeedbackGetAllByServiceIDRes{
			ID:        feedback.ID,
			UserName:  feedback.UserName,
			Rating:    feedback.Rating,
			Comment:   feedback.Comment,
			CreatedAt: feedback.CreatedAt,
		})
	}

	return res, nil
}
