package expenses

import (
	"context"
	"fmt"
	"log/slog"
	"ms_expense/internal/core/http_client/groups"
	"shared/apierror"
	"shared/auth/contexts"
	"shared/cache"
	"shared/filters"
	"shared/validator"
	"time"
	"uuid"
)

type Service interface {
	Create(
		ctx context.Context,
		model *Expense,
		participants []uuid.UUID,
	) error
	FindByID(ctx context.Context, id uuid.UUID) (*Expense, error)
	FindAll(
		ctx context.Context,
		groupID uuid.UUID,
		f filters.Filters,
		startDate, endDate *time.Time,
	) ([]*Expense, filters.Metadata, error)
	Update(
		ctx context.Context,
		id uuid.UUID,
		participants []uuid.UUID,
		model *Expense,
	) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type writeExecutor interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}

type producer interface {
	PublishExpenseCreated(
		ctx context.Context,
		event ExpenseCreated,
	) error

	PublishExpenseUpdated(
		ctx context.Context,
		event ExpenseUpdated,
	) error

	PublishExpenseDeleted(
		ctx context.Context,
		event ExpenseDeleted,
	) error
}

type expenseService struct {
	repo        repository
	we          writeExecutor
	groupClient groupClient
	producer    producer
	kb          cache.KeyBuilder
	cache       cache.Cache
	logger      *slog.Logger
}

type groupClient interface {
	GetGroup(ctx context.Context, id uuid.UUID) (groups.GroupDetailDTO, error)
}

func newService(
	repo repository,
	we writeExecutor,
	groupClient groupClient,
	producer producer,
	kb cache.KeyBuilder,
	c cache.Cache,
	logger *slog.Logger,
) *expenseService {
	return &expenseService{
		repo:        repo,
		we:          we,
		groupClient: groupClient,
		producer:    producer,
		kb:          cache.NewKeyBuilder("ms_expense"),
		cache:       c,
		logger:      logger,
	}
}

func (s *expenseService) Create(
	ctx context.Context,
	model *Expense,
	participants []uuid.UUID,
) error {
	v := validator.New()
	if model.Validate(v); !v.Valid() {
		return apierror.NewValidationError(v.Errors)
	}

	groupDetail, err := s.groupClient.GetGroup(ctx, model.GroupID)
	if err != nil {
		return err
	}

	shares, err := s.mountShares(model, groupDetail, participants)
	if err != nil {
		return err
	}

	var sum int64
	for _, sh := range shares {
		sum += sh.Amount
	}
	if sum != model.Amount {
		return apierror.NewHTTPError(
			"shares do not sum to total amount", 422, nil)
	}

	if err := s.we.Execute(ctx, func(ctx context.Context) error {
		if err := s.repo.Insert(ctx, model); err != nil {
			return err
		}

		return s.repo.InsertShares(ctx, model.ID, shares)
	}); err != nil {
		return err
	}

	if err := s.producer.PublishExpenseCreated(ctx, ExpenseCreated{
		EventID:    uuid.New(),
		ExpenseID:  model.ID,
		GroupID:    model.GroupID,
		PaidBy:     model.PaidBy,
		Amount:     model.Amount,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		s.logger.Error("failed to publish expense.created",
			"expense_id", model.ID,
			"group_id", model.GroupID,
			"error", err,
		)
	}

	return nil
}

func (s *expenseService) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*Expense, error) {
	key := s.kb.BuildItemKey(id.String())
	model, err := cache.FetchOrCache(ctx, s.cache, key, func() (*Expense, error) {
		return s.repo.FindByID(ctx, id)
	})
	if err != nil {
		return nil, err
	}

	if err := s.assertMembership(ctx, model.GroupID); err != nil {
		return nil, err
	}

	return model, nil
}

func (s *expenseService) FindAll(
	ctx context.Context,
	groupID uuid.UUID,
	f filters.Filters,
	startDate, endDate *time.Time,
) ([]*Expense, filters.Metadata, error) {
	if err := s.assertMembership(ctx, groupID); err != nil {
		return nil, filters.Metadata{}, err
	}

	type listPayload struct {
		models   []*Expense
		metadata filters.Metadata
	}

	key := s.kb.BuildListKey(
		cache.UUID(groupID),
		cache.OptTime(startDate),
		cache.OptTime(endDate),
		cache.Int(f.Page),
		cache.Int(f.PageSize),
		cache.Str(f.Sort),
	)

	payload, err := cache.FetchOrCache(ctx, s.cache, key, func() (listPayload, error) {
		models, metadata, err := s.repo.FindByGroupID(ctx, groupID, startDate, endDate, f)
		if err != nil {
			return listPayload{}, err
		}
		return listPayload{
			models:   models,
			metadata: metadata,
		}, nil
	})

	if err != nil {
		return nil, filters.Metadata{}, err
	}

	return payload.models, payload.metadata, nil
}

func (s *expenseService) Update(
	ctx context.Context,
	id uuid.UUID,
	participants []uuid.UUID,
	model *Expense,
) error {
	saved, err := s.FindByID(ctx, id)
	if err != nil {
		return err
	}

	groupDetail, err := s.groupClient.GetGroup(ctx, saved.GroupID)
	if err != nil {
		return err
	}

	me := contexts.GetUser(ctx).GetID()
	if saved.CreatedBy != &me && groupDetail.OwnerID != me {
		return apierror.ErrForbidden
	}

	shares, err := s.mountShares(model, groupDetail, participants)
	if err != nil {
		return err
	}

	if err := s.we.Execute(ctx, func(ctx context.Context) error {
		if err := s.repo.Update(ctx, model); err != nil {
			return err
		}

		return s.repo.InsertShares(ctx, model.ID, shares)
	}); err != nil {
		return err
	}

	if err := s.producer.PublishExpenseUpdated(ctx, ExpenseUpdated{
		EventID:    uuid.New(),
		ExpenseID:  model.ID,
		GroupID:    model.GroupID,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		s.logger.Error("failed to publish expense.created",
			"expense_id", model.ID,
			"group_id", model.GroupID,
			"error", err,
		)
	}
	return nil
}

func (s *expenseService) Delete(ctx context.Context, id uuid.UUID) error {
	e, err := s.FindByID(ctx, id)
	if err != nil {
		return err
	}

	group, err := s.groupClient.GetGroup(ctx, e.GroupID)
	if err != nil {
		return err
	}
	me := contexts.GetUser(ctx).GetID()
	if e.CreatedBy != &me && group.OwnerID != me {
		return apierror.ErrForbidden
	}

	if err := s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.SoftDelete(ctx, id)
	}); err != nil {
		return err
	}

	if err := s.producer.PublishExpenseDeleted(ctx, ExpenseDeleted{
		EventID:    uuid.New(),
		ExpenseID:  e.ID,
		GroupID:    e.GroupID,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		s.logger.Error("failed to publish expense.deleted",
			"expense_id", e.ID,
			"group_id", e.GroupID,
			"error", err,
		)
	}

	return nil
}

func (s *expenseService) mountShares(
	model *Expense,
	groupDetail groups.GroupDetailDTO,
	participants []uuid.UUID,
) ([]Share, error) {
	memberSet := make(map[uuid.UUID]groups.MemberDTO, len(groupDetail.Members))
	for _, m := range groupDetail.Members {
		if _, exists := memberSet[m.ID]; !exists {
			memberSet[m.ID] = m
		}
	}

	if _, ok := memberSet[model.PaidBy]; !ok {
		return nil, apierror.NewHTTPError(
			"paid_by must be a group member", 422, nil)
	}
	shares := make([]Share, len(model.Shares))
	switch model.SplitMode {
	case SplitModeEqual:
		shares = CalculateEqualShares(model.Amount, participants)
	case SplitModeExact:
		for i, s := range model.Shares {
			if _, ok := memberSet[s.UserID]; !ok {
				return nil, apierror.NewHTTPError(
					fmt.Sprintf("user %s is not a group member", s.UserID), 422, nil)
			}
			shares[i] = Share{UserID: s.UserID, Amount: s.Amount}
		}
	}
	return shares, nil
}

func (s *expenseService) assertMembership(ctx context.Context, groupID uuid.UUID) error {
	if _, err := s.groupClient.GetGroup(ctx, groupID); err != nil {
		return err
	}
	return nil
}
