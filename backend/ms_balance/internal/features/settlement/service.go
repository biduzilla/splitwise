package settlement

import (
	"context"
	"shared/auth/contexts"
	"shared/cache"
	"time"
	"uuid"
)

type settlementService struct {
	repo     repo
	we       writeExecutor
	producer producer
	c        cache.Cache
	kb       cache.KeyBuilder
}

type repo interface {
	Insert(
		ctx context.Context,
		m *Settlement,
	) error

	SoftDelete(
		ctx context.Context,
		id uuid.UUID,
	) error

	FindByGroupID(
		ctx context.Context,
		groupID uuid.UUID,
	) ([]*Settlement, error)

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (*Settlement, error)
}

type writeExecutor interface {
	Execute(
		ctx context.Context,
		fn func(ctx context.Context) error,
	) error
}

type producer interface {
	PublishSettlementCreated(ctx context.Context, event SettlementCreated) error
	PublishSettlementDeleted(ctx context.Context, event SettlementDeleted) error
}

func newService(
	repo repo,
	we writeExecutor,
	producer producer,
	c cache.Cache,
	kb cache.KeyBuilder,
) *settlementService {
	return &settlementService{
		repo:     repo,
		we:       we,
		producer: producer,
		c:        c,
		kb:       kb,
	}
}

func (s *settlementService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*Settlement, error) {
	key := s.kb.BuildListKey(
		cache.UUID(contexts.GetUser(ctx).GetID()),
		cache.UUID(id),
	)

	return cache.FetchOrCache(ctx, s.c, key, func() (*Settlement, error) {
		return s.repo.GetByID(ctx, id)
	})
}

func (s *settlementService) FindByGroupID(
	ctx context.Context,
	groupID uuid.UUID,
) ([]*Settlement, error) {
	key := s.kb.BuildItemKey(groupID.String())

	return cache.FetchOrCache(ctx, s.c, key, func() ([]*Settlement, error) {
		return s.repo.FindByGroupID(ctx, groupID)
	})
}

func (s *settlementService) Create(
	ctx context.Context,
	model *Settlement,
) error {
	err := s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Insert(ctx, model)
	})

	if err != nil {
		return err
	}

	_ = s.producer.PublishSettlementCreated(ctx, SettlementCreated{
		EventID:      uuid.New(),
		SettlementID: model.ID,
		GroupID:      model.GroupID,
		OccurredAt:   time.Now().UTC(),
		Token:        contexts.GetToken(ctx),
	})

	return nil
}

func (s *settlementService) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	err := s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.SoftDelete(ctx, id)
	})

	if err != nil {
		return err
	}

	_ = s.producer.PublishSettlementDeleted(ctx, SettlementDeleted{
		EventID:      uuid.New(),
		SettlementID: id,
		OccurredAt:   time.Now().UTC(),
		Token:        contexts.GetToken(ctx),
	})
	return nil
}
