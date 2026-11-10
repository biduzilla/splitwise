package api

import (
	"log/slog"
	"messaging"
	"ms_expense/internal/core/messaging/producers"

	"github.com/IBM/sarama"
)

type producer struct {
	expenses *producers.PublisherExpense
}

func NewProducer(p sarama.SyncProducer, logger *slog.Logger) *producer {
	producerGeral := messaging.NewProducer(p, logger)
	return &producer{
		expenses: producers.NewPublisherExpense(producerGeral),
	}
}
