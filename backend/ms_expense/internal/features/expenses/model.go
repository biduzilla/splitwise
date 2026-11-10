package expenses

import (
	"shared/models"
	"shared/validator"
	"time"
	"uuid"
)

type SplitMode string

const (
	SplitModeEqual SplitMode = "equal"
	SplitModeExact SplitMode = "exact"
)

func (s SplitMode) String() string {
	return string(s)
}

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

func (e *Expense) ToDTO() ExpenseDTO {
	sharesDto := make([]ShareDTO, len(e.Shares))
	for i, s := range e.Shares {
		sharesDto[i] = s.ToDTO()
	}

	return ExpenseDTO{
		ID:          e.ID,
		GroupID:     e.GroupID,
		PaidBy:      e.PaidBy,
		Amount:      e.Amount,
		Description: e.Description,
		Category:    e.Category,
		SplitMode:   e.SplitMode,
		ExpenseDate: e.ExpenseDate.String(),
		Version:     e.Version,
		CreatedAt:   e.CreatedAt,
		Shares:      sharesDto,
	}
}

func (d Expense) Validate(v *validator.Validator) {
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
		validator.In(d.SplitMode.String(), "equal", "exact"),
		"split_mode",
		"must be one of: equal, exact",
	)

	switch d.SplitMode {
	case "equal":
		v.Check(len(d.Shares) == 0, "shares", "must not be provided when split_mode is 'equal'")

	case "exact":
		v.CheckSliceNotEmpty(d.Shares, "shares")
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
}

type Share struct {
	UserID uuid.UUID
	Amount int64
}

func (m *Share) ToDTO() ShareDTO {
	return ShareDTO{
		UserID: m.UserID,
		Amount: m.Amount,
	}
}
