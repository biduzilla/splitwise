package expenses

import (
	"time"
	"uuid"
)

type ShareDTO struct {
	UserID uuid.UUID `json:"user_id"`
	Amount int64     `json:"amount"`
}

type CreateExpenseDTO struct {
	GroupID      uuid.UUID   `json:"group_id"`
	PaidBy       uuid.UUID   `json:"paid_by"`
	Amount       int64       `json:"amount"`
	Description  string      `json:"description"`
	Category     *string     `json:"category,omitempty"`
	SplitMode    string      `json:"split_mode"`
	Participants []uuid.UUID `json:"participants,omitempty"`
	Shares       []ShareDTO  `json:"shares,omitempty"`
	ExpenseDate  string      `json:"expense_date,omitempty"`
}

type UpdateExpenseDTO struct {
	PaidBy       uuid.UUID   `json:"paid_by"`
	Amount       int64       `json:"amount"`
	Description  string      `json:"description"`
	Category     *string     `json:"category,omitempty"`
	SplitMode    SplitMode   `json:"split_mode"`
	Participants []uuid.UUID `json:"participants,omitempty"`
	Shares       []ShareDTO  `json:"shares,omitempty"`
	ExpenseDate  string      `json:"expense_date,omitempty"`
	Version      int         `json:"version"`
}

type ExpenseDTO struct {
	ID          uuid.UUID  `json:"id"`
	GroupID     uuid.UUID  `json:"group_id"`
	PaidBy      uuid.UUID  `json:"paid_by"`
	Amount      int64      `json:"amount"`
	Description string     `json:"description"`
	Category    *string    `json:"category"`
	SplitMode   SplitMode  `json:"split_mode"`
	ExpenseDate string     `json:"expense_date"`
	Version     int        `json:"version"`
	CreatedAt   time.Time  `json:"created_at"`
	Shares      []ShareDTO `json:"shares"`
}
