package producers

import (
	"context"
	"messaging"
	"ms_expense/internal/features/expenses"
)

type PublisherExpense struct {
	producer *messaging.Producer
}

func (p *PublisherExpense) PublishExpenseCreated(
	ctx context.Context,
	event expenses.ExpenseCreated,
) error {
	return p.producer.Publish(ctx, messaging.RateioEvents, event)
}

func (p *PublisherExpense) PublishExpenseUpdated(
	ctx context.Context,
	event expenses.ExpenseUpdated,
) error {
	return p.producer.Publish(ctx, messaging.RateioEvents, event)
}

func (p *PublisherExpense) PublishExpenseDeleted(
	ctx context.Context,
	event expenses.ExpenseDeleted,
) error {
	return p.producer.Publish(ctx, messaging.RateioEvents, event)
}
