package balance

import (
	"context"
	"ms_balance/internal/core/http_client/expenses"
	"ms_balance/internal/core/http_client/groups"
	"ms_balance/internal/features/settlement"
	"shared/cache"
	"shared/transaction"
	"uuid"
)

type repo interface {
	GetByGroupID(
		ctx context.Context,
		groupID uuid.UUID,
	) (map[uuid.UUID]int64, error)

	UpsertBatch(
		ctx context.Context,
		groupID uuid.UUID,
		balances map[uuid.UUID]int64,
	) error
}

type groupClient interface {
	GetGroup(ctx context.Context, id uuid.UUID) (groups.GroupDetailDTO, error)
}

type expenseClient interface {
	GetExpense(
		ctx context.Context,
		groupID uuid.UUID,
	) ([]expenses.ExpenseDTO, error)
}

type settlementService interface {
	FindByGroupID(
		ctx context.Context,
		groupID uuid.UUID,
	) ([]*settlement.Settlement, error)
}

type writeExecutor interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error)
}

type balanceService struct {
	repo              repo
	settlementService settlementService
	groupClient       groupClient
	expenseClient     expenseClient
	kb                cache.KeyBuilder
	c                 cache.Cache
	we                transaction.WriteExecutor
}

func (s *balanceService) GetBalances(
	ctx context.Context,
	groupID uuid.UUID,
) (map[uuid.UUID]int64, error) {
	key := s.kb.BuildItemKey(groupID.String())

	return cache.FetchOrCache(ctx, s.c, key, func() (map[uuid.UUID]int64, error) {
		return s.repo.GetByGroupID(ctx, groupID)
	})
}

func (s *balanceService) GetSuggestedTransfers(
	ctx context.Context,
	groupID uuid.UUID,
) ([]Transfer, error) {
	balances, err := s.GetBalances(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return Minimize(balances), nil
}

func (s *balanceService) Recalculate(
	ctx context.Context,
	groupID uuid.UUID,
) error {
	return s.we.Execute(ctx, func(ctx context.Context) error {
		expenses, err := s.expenseClient.GetExpense(ctx, groupID)
		if err != nil {
			return err
		}

		settlements, err := s.settlementService.FindByGroupID(ctx, groupID)
		if err != nil {
			return err
		}

		balances := make(map[uuid.UUID]int64)

		for _, exp := range expenses {
			balances[exp.PaidBy] += exp.Amount
			for _, share := range exp.Shares {
				balances[share.UserID] -= share.Amount
			}
		}

		for _, set := range settlements {
			balances[set.FromUserID] += set.Amount
			balances[set.ToUserID] -= set.Amount
		}

		if err := s.repo.UpsertBatch(ctx, groupID, balances); err != nil {
			return err
		}
		return nil
	})

}
