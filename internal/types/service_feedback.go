package types

import (
	"time"

	"github.com/google/uuid"
)

// region repo types

type ServiceFeedback struct {
	ID        uuid.UUID `db:"id"`
	ServiceID uuid.UUID `db:"service_id"`
	OrderID   uuid.UUID `db:"order_id"`
	Rating    int16     `db:"rating"`
	Comment   string    `db:"comment"`
	CreatedAt time.Time `db:"created_at"`
}

type ServiceFeedbackWithUser struct {
	ServiceFeedback
	UserName string `db:"user_name"`
}

// endregion repo types

// region service types

type ServiceFeedbackGetAllByServiceIDReq struct {
	ID uuid.UUID `param:"id"`
}

func (r ServiceFeedbackGetAllByServiceIDReq) Validate() error {
	if r.ID == uuid.Nil {
		return ErrIDRouteParamRequired
	}

	return nil
}

type ServiceFeedbackGetAllByServiceIDRes struct {
	ID        uuid.UUID `json:"id"`
	UserName  string    `json:"user_name"`
	Rating    int16     `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

// endregion service types
