package expense

import (
	"time"
	"uuid"
)

type SplitMode string

const (
	SplitEqual SplitMode = "equal"
	SplitExact SplitMode = "exact"
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
	Version     int
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	UpdatedBy   *uuid.UUID
	Deleted     bool
	Shares      []Share
}

type Share struct {
	UserID uuid.UUID
	Amount int64
}
