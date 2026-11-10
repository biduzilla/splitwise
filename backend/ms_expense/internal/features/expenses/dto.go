package expenses

import (
	"fmt"
	"shared/models"
	"time"
	"uuid"
)

const expenseDateLayout = "2006-01-02"

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

func (d ShareDTO) ToModel() Share {
	return Share{
		UserID: d.UserID,
		Amount: d.Amount,
	}
}

func (d CreateExpenseDTO) ToModel() (*Expense, error) {
	expenseDate, err := parseExpenseDate(d.ExpenseDate)
	if err != nil {
		return nil, err
	}

	return &Expense{
		GroupID:     d.GroupID,
		PaidBy:      d.PaidBy,
		Amount:      d.Amount,
		Description: d.Description,
		Category:    d.Category,
		SplitMode:   SplitMode(d.SplitMode),
		ExpenseDate: expenseDate,
		Shares:      sharesToModel(d.Shares),
	}, nil
}

func (d UpdateExpenseDTO) ToModel() (*Expense, error) {
	expenseDate, err := parseExpenseDate(d.ExpenseDate)
	if err != nil {
		return nil, err
	}

	return &Expense{
		PaidBy:      d.PaidBy,
		Amount:      d.Amount,
		Description: d.Description,
		Category:    d.Category,
		SplitMode:   d.SplitMode,
		ExpenseDate: expenseDate,
		Shares:      sharesToModel(d.Shares),
		BaseModel:   models.BaseModel{Version: d.Version},
	}, nil
}

func sharesToModel(dtos []ShareDTO) []Share {
	if len(dtos) == 0 {
		return nil
	}
	shares := make([]Share, len(dtos))
	for i, s := range dtos {
		shares[i] = s.ToModel()
	}
	return shares
}

func parseExpenseDate(s string) (time.Time, error) {
	if s == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	t, err := time.Parse(expenseDateLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid expense_date: %w", err)
	}
	return t, nil
}
