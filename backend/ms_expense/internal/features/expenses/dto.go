package expenses

import (
	"shared/validator"
	"slices"
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

func (d CreateExpenseDTO) Validate(v *validator.Validator) {
	v.CheckUUIDNotNil(d.GroupID, "group_id")
	v.CheckUUIDNotNil(d.PaidBy, "paid_by")

	v.Check(d.Amount > 0, "amount", "must be greater than zero")

	v.CheckStringNotEmpty(d.Description, "description")
	v.CheckStringMinLen(d.Description, 3, "description")   // era "name", bug
	v.CheckStringMaxLen(d.Description, 100, "description") // era "name", bug

	if d.Category != nil {
		v.CheckStringMaxLen(*d.Category, 50, "category")
	}

	v.Check(
		validator.In(d.SplitMode, "equal", "exact"),
		"split_mode",
		"must be one of: equal, exact",
	)

	switch d.SplitMode {
	case "equal":
		v.CheckSliceNotEmpty(d.Participants, "participants")
		v.Check(len(d.Shares) == 0, "shares", "must not be provided when split_mode is 'equal'")

		ids := make([]string, 0, len(d.Participants))
		for _, p := range d.Participants {
			ids = append(ids, p.String())
		}
		v.Check(validator.Unique(ids), "participants", "must not contain duplicate user IDs")

		v.Check(
			slices.Contains(d.Participants, d.PaidBy),
			"paid_by",
			"must be one of the participants",
		)

	case "exact":
		v.CheckSliceNotEmpty(d.Shares, "shares")
		v.Check(len(d.Participants) == 0, "participants", "must not be provided when split_mode is 'exact'")

		var total int64
		ids := make([]string, 0, len(d.Shares))
		payerInShares := false

		for _, s := range d.Shares {
			v.CheckUUIDNotNil(s.UserID, "shares")
			v.Check(s.Amount >= 0, "shares", "amount must be greater than or equal to zero")
			if s.UserID == d.PaidBy {
				payerInShares = true
			}
			ids = append(ids, s.UserID.String())
			total += s.Amount
		}

		v.Check(validator.Unique(ids), "shares", "must not contain duplicate user IDs")
		v.Check(total == d.Amount, "shares", "sum of shares must equal the total amount")
		v.Check(payerInShares, "paid_by", "must be one of the share users")
	}

	if d.ExpenseDate != "" {
		_, err := time.Parse("2006-01-02", d.ExpenseDate)
		v.Check(err == nil, "expense_date", "must be a valid date in the format YYYY-MM-DD")
	}
}
