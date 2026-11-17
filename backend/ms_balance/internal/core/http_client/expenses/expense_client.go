package expenses

import (
	"context"
	"fmt"
	"shared/filters"
	"shared/httpx/httpclient"
	"time"
	"uuid"
)

type SplitMode string

const (
	SplitModeEqual SplitMode = "equal"
	SplitModeExact SplitMode = "exact"
)

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

type ShareDTO struct {
	UserID uuid.UUID `json:"user_id"`
	Amount int64     `json:"amount"`
}

type HTTPClient struct {
	base *httpclient.BaseClient
}

func NewClient(baseURL string, timout time.Duration) *HTTPClient {
	return &HTTPClient{
		base: httpclient.New(baseURL, timout),
	}
}

func (c *HTTPClient) GetExpense(
	ctx context.Context,
	groupID uuid.UUID,
) ([]ExpenseDTO, error) {
	var allExpenses []ExpenseDTO
	page := 1
	pageSize := 100

	for {
		path := fmt.Sprintf(
			"?group_id=%s&page=%d&page_size=%d",
			groupID.String(),
			page,
			pageSize,
		)

		var response filters.Page[ExpenseDTO]
		err := c.base.Get(ctx, path, &response, nil)
		if err != nil {
			return nil, err
		}

		allExpenses = append(allExpenses, response.Content...)

		if page >= response.Metadata.LastPage {
			break
		}
		page++
	}
	return allExpenses, nil
}
