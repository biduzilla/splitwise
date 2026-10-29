package expenses

import (
	"shared/models"
	"time"
	"uuid"
)

type SplitMode string

const (
	SplitModeEqual SplitMode = "equal"
	SplitModeExact SplitMode = "exact"
)

type Expense struct {
	ID          uuid.UUID
	GroupID     uuid.UUID
	PaidBy      uuid.UUID
	Amount      int64
	Description string
	Category    *string
	SplitMode   SplitMode
	ExpenseDate time.Time
	models.BaseModel
	Shares []Share
}

type Share struct {
	UserID uuid.UUID
	Amount int64
}
