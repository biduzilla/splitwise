package producers

import (
	"context"
	"messaging"
	"ms_balance/internal/features/settlement"
)

type PublisherSettlement struct {
	producer *messaging.Producer
}

func NewPublisherSettlement(
	producer *messaging.Producer,
) *PublisherSettlement {
	return &PublisherSettlement{
		producer: producer,
	}
}

func (p *PublisherSettlement) PublishSettlementCreated(
	ctx context.Context,
	event settlement.SettlementCreated,
) error {
	return p.producer.Publish(ctx, messaging.RateioEvents, event)
}

func (p *PublisherSettlement) PublishSettlementDeleted(
	ctx context.Context,
	event settlement.SettlementDeleted,
) error {
	return p.producer.Publish(ctx, messaging.RateioEvents, event)
}
