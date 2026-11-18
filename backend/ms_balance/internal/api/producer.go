package api

import (
	"log/slog"
	"messaging"
	"ms_balance/internal/core/messaging/producers"

	"github.com/IBM/sarama"
)

type producer struct {
	settlement *producers.PublisherSettlement
}

func newProducer(p sarama.SyncProducer, logger *slog.Logger) *producer {
	producerGeral := messaging.NewProducer(p, logger)
	return &producer{
		settlement: producers.NewPublisherSettlement(producerGeral),
	}
}
