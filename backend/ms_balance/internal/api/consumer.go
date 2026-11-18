package api

import (
	"context"
	"log/slog"
	"messaging"
	"ms_balance/internal/core/messaging/consumers"
	"shared/cache"

	"github.com/IBM/sarama"
)

type consumersWrapper struct {
	kafkaConsumer *messaging.Consumer
}

func NewConsumer(
	consumerGroup sarama.ConsumerGroup,
	m *modules,
	cacheClient cache.Cache,
	logger *slog.Logger,
) *consumersWrapper {
	handler := consumers.NewConsumers(m.balance.Service, logger, cacheClient)

	handlersMap := map[string]messaging.EventHandler{
		"rateio.events": handler,
	}

	kafkaSvc := messaging.NewConsumer(
		consumerGroup,
		handlersMap,
		"ms_balance/internal/core/messaging",
		logger,
	)

	return &consumersWrapper{
		kafkaConsumer: kafkaSvc,
	}
}

func (c *consumersWrapper) Start(logger *slog.Logger) (shutdownFunc func()) {
	logger.Info("starting kafka consumers...")

	consumerCtx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		if err := c.kafkaConsumer.Start(consumerCtx); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	shutdownFunc = func() {
		logger.Info("stopping kafka consumers...")
		cancel()
		<-errCh
		logger.Info("kafka consumers stopped")
	}

	go func() {
		if err := <-errCh; err != nil {
			logger.Error("kafka consumer crashed", "error", err)
		}
	}()

	return shutdownFunc
}
