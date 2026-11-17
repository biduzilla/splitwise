package transaction

import (
	"context"
	"log/slog"
	"shared/cache"
	"time"
)

type writeExecutor struct {
	tx         Manager
	cache      cache.Cache
	keyBuilder cache.KeyBuilder
}

type WriteExecutor interface {
	Execute(
		ctx context.Context,
		fn func(ctx context.Context) error,
	) error
}

func NewWriterExecutor(
	tx Manager,
	cache cache.Cache,
	keyBuilder cache.KeyBuilder,
) *writeExecutor {
	return &writeExecutor{
		tx:         tx,
		cache:      cache,
		keyBuilder: keyBuilder,
	}
}

func (e *writeExecutor) Execute(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	err := e.tx.RunInTx(ctx, fn)
	if err != nil {
		return err
	}

	detached := context.WithoutCancel(ctx)

	go func() {
		invCtx, cancel := context.WithTimeout(detached, 5*time.Second)
		defer cancel()

		if err := e.cache.DeleteByPrefix(invCtx, e.keyBuilder.GetPrefix()); err != nil {
			slog.Error("cache invalidate failed", "err", err, "prefix", e.keyBuilder.GetPrefix())
		}
	}()

	return nil
}
