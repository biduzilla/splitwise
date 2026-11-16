package settlement

import (
	"shared/models"
	"shared/validator"
	"time"
	"uuid"
)

type Settlement struct {
	ID         uuid.UUID
	GroupID    uuid.UUID
	FromUserID uuid.UUID
	ToUserID   uuid.UUID
	Amount     int64
	SettledAt  time.Time
	models.BaseModel
}

type CreateSettlementDTO struct {
	GroupID    uuid.UUID `json:"group_id"`
	FromUserID uuid.UUID `json:"from_user_id"`
	ToUserID   uuid.UUID `json:"to_user_id"`
	Amount     int64     `json:"amount"`
}

func (d *CreateSettlementDTO) Validate(v *validator.Validator) {
	v.CheckUUIDNotNil(d.GroupID, "group_id")
	v.CheckUUIDNotNil(d.FromUserID, "from_user_id")
	v.CheckUUIDNotNil(d.ToUserID, "to_user_id")
	v.Check(d.FromUserID != d.ToUserID, "to_user_id", "must be different from from_user_id")
	v.Check(d.Amount > 0, "amount", "must be greater than zero")
}

func (d *CreateSettlementDTO) ToModel() *Settlement {
	return &Settlement{
		GroupID:    d.GroupID,
		FromUserID: d.FromUserID,
		ToUserID:   d.ToUserID,
		Amount:     d.Amount,
	}
}
