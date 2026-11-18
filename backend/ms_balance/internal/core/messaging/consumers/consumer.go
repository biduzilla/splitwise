package consumers

import (
	"context"
	"encoding/json/v2"
	"log/slog"
	"shared/auth/contexts"
	"shared/cache"
	"time"
	"uuid"
)

type Consumers struct {
	svc    balanceService
	logger *slog.Logger
	c      cache.Cache
	kb     cache.KeyBuilder
}

type balanceService interface {
	Recalculate(
		ctx context.Context,
		groupID uuid.UUID,
	) error
}

func NewConsumers(
	svc balanceService,
	logger *slog.Logger,
	c cache.Cache,
) *Consumers {
	return &Consumers{
		svc:    svc,
		logger: logger,
		c:      c,
		kb:     cache.NewKeyBuilder("kafka"),
	}
}

type Event struct {
	EventID   uuid.UUID `json:"event_id"`
	GroupID   uuid.UUID `json:"group_id"`
	EventType string    `json:"event_type"`
	Token     string    `json:"token"`
}

func (c *Consumers) Handle(
	ctx context.Context,
	message []byte,
	topic string,
) error {
	var evt Event
	if err := json.Unmarshal(message, &evt); err != nil {
		return err
	}

	seenKey := c.kb.BuildItemKey("kafka:seen:" + evt.EventID.String())
	seen, err := c.c.Exists(ctx, seenKey)
	if err == nil && seen {
		return nil
	}

	ctx = contexts.SetToken(ctx, evt.Token)

	backoff := time.Second
	maxAttempts := 3
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := c.svc.Recalculate(ctx, evt.GroupID); err != nil {
			lastErr = err
			c.logger.Warn("recalculate failed, will retry",
				"group_id", evt.GroupID,
				"attempt", attempt,
				"max_attempts", maxAttempts,
				"error", err,
			)
			if attempt < maxAttempts {
				select {
				case <-time.After(backoff):
					backoff *= 2
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			continue
		}
		ttl := 7 * 24 * time.Hour
		_ = c.c.Set(ctx, seenKey, []byte("1"), &ttl)
		return nil
	}

	c.logger.Error("recalculate failed after max retries",
		"group_id", evt.GroupID,
		"error", lastErr,
	)
	return lastErr
}
