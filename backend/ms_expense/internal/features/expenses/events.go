package expenses

import (
	"time"
	"uuid"
)

type ExpenseCreated struct {
	EventID    uuid.UUID `json:"event_id"`
	ExpenseID  uuid.UUID `json:"expense_id"`
	GroupID    uuid.UUID `json:"group_id"`
	PaidBy     uuid.UUID `json:"paid_by"`
	Amount     int64     `json:"amount"`
	Token      string    `json:"token"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (e ExpenseCreated) Key() string { return e.GroupID.String() }

type ExpenseUpdated struct {
	EventID    uuid.UUID `json:"event_id"`
	ExpenseID  uuid.UUID `json:"expense_id"`
	GroupID    uuid.UUID `json:"group_id"`
	OccurredAt time.Time `json:"occurred_at"`
	Token      string    `json:"token"`
}

func (e ExpenseUpdated) Key() string { return e.GroupID.String() }

type ExpenseDeleted struct {
	EventID    uuid.UUID `json:"event_id"`
	ExpenseID  uuid.UUID `json:"expense_id"`
	GroupID    uuid.UUID `json:"group_id"`
	OccurredAt time.Time `json:"occurred_at"`
	Token      string    `json:"token"`
	EventType  string    `json:"event_type"`
}

func (e ExpenseDeleted) Key() string { return e.GroupID.String() }
