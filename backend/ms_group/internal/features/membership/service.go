package membership

import (
	"context"
	"fmt"
	"shared/apierror"
	"shared/auth/contexts"
	"shared/cache"
	"uuid"
)

type groupFinder interface {
	FindOwnerID(ctx context.Context, groupID uuid.UUID) (uuid.UUID, error)
}

type WriteExecutor interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}
type MembershipService struct {
	repo        repository
	cache       cache.Cache
	we          WriteExecutor
	kb          cache.KeyBuilder
	groupFinder groupFinder
}

func NewService(
	repo repository,
	cache cache.Cache,
	we WriteExecutor,
	kb cache.KeyBuilder,
	groupFinder groupFinder,
) *MembershipService {
	return &MembershipService{
		repo:        repo,
		cache:       cache,
		we:          we,
		kb:          kb,
		groupFinder: groupFinder,
	}
}

type service interface {
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*Membership, error)
	FindByGroupIDAndUserID(ctx context.Context, groupID, userID uuid.UUID) (*Membership, error)
	Insert(ctx context.Context, m *Membership) error
	Remove(ctx context.Context, groupID, userID uuid.UUID) error
	CountActiveMembers(ctx context.Context, groupID uuid.UUID) (int, error)
}

func (s *MembershipService) FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*Membership, error) {
	key := s.kb.BuildItemKey(groupID.String())
	return cache.FetchOrCache(ctx, s.cache, key, func() ([]*Membership, error) {
		return s.repo.FindByGroupID(ctx, groupID)
	})
}

func (s *MembershipService) FindByGroupIDAndUserID(ctx context.Context, groupID, userID uuid.UUID) (*Membership, error) {
	key := s.kb.BuildListKey(
		cache.UUID(groupID),
		cache.UUID(userID),
	)

	return cache.FetchOrCache(ctx, s.cache, key, func() (*Membership, error) {
		return s.repo.FindByGroupIDAndUserID(ctx, groupID, userID)
	})
}

func (s *MembershipService) CountActiveMembers(ctx context.Context, groupID uuid.UUID) (int, error) {
	return s.repo.CountActiveMembers(ctx, groupID)
}

func (s *MembershipService) Insert(ctx context.Context, m *Membership) error {
	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Insert(ctx, m)
	})
}

func (s *MembershipService) Remove(ctx context.Context, groupID, userID uuid.UUID) error {
	actorID := contexts.GetUser(ctx).GetID()

	ownerID, err := s.groupFinder.FindOwnerID(ctx, groupID)
	if err != nil {
		return err
	}
	if ownerID != actorID {
		return apierror.NewBadRequestError(fmt.Errorf("You are not owner"))
	}
	if userID == actorID {
		return apierror.NewValidationError(map[string]string{
			"user_id": "owner cannot remove themselves",
		})
	}

	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Remove(ctx, groupID, userID)
	})
}

func (s *MembershipService) SetGroupFinder(gf groupFinder) {
	s.groupFinder = gf
}
