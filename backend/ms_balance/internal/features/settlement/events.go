package settlement

import (
	"time"
	"uuid"
)

type SettlementCreated struct {
	EventID      uuid.UUID `json:"event_id"`
	SettlementID uuid.UUID `json:"settlement_id"`
	GroupID      uuid.UUID `json:"group_id"`
	FromUserID   uuid.UUID `json:"from_user_id"`
	ToUserID     uuid.UUID `json:"to_user_id"`
	Amount       int64     `json:"amount"`
	OccurredAt   time.Time `json:"occurred_at"`
	Token        string    `json:"token"`
}

func (e SettlementCreated) Key() string {
	return e.GroupID.String()
}

type SettlementDeleted struct {
	EventID      uuid.UUID `json:"event_id"`
	SettlementID uuid.UUID `json:"settlement_id"`
	GroupID      uuid.UUID `json:"group_id"`
	OccurredAt   time.Time `json:"occurred_at"`
	Token        string    `json:"token"`
}

func (e SettlementDeleted) Key() string {
	return e.GroupID.String()
}
