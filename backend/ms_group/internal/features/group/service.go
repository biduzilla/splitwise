package group

import (
	"context"
	"fmt"
	"ms_group/internal/features/membership"
	"shared/apierror"
	"shared/auth/contexts"
	"shared/cache"
	"shared/filters"
	"shared/transaction"
	"shared/validator"
	"uuid"
)

type GroupService struct {
	repo              repository
	cache             cache.Cache
	we                transaction.WriteExecutor
	kb                cache.KeyBuilder
	membershipService membershipService
}

type membershipService interface {
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*membership.Membership, error)
	FindByGroupIDAndUserID(ctx context.Context, groupID, userID uuid.UUID) (*membership.Membership, error)
	Insert(ctx context.Context, m *membership.Membership) error
	Remove(ctx context.Context, groupID, userID uuid.UUID) error
	CountActiveMembers(ctx context.Context, groupID uuid.UUID) (int, error)
}

func NewService(
	repo repository,
	cache cache.Cache,
	we transaction.WriteExecutor,
	kb cache.KeyBuilder,
	membershipService membershipService,
) *GroupService {
	return &GroupService{
		repo:              repo,
		cache:             cache,
		we:                we,
		kb:                kb,
		membershipService: membershipService,
	}
}

type service interface {
	Create(ctx context.Context, dto CreateGroupDTO) (*Group, error)
	FindByID(ctx context.Context, id uuid.UUID) (*GroupDetail, error)
	ListByUser(ctx context.Context, f filters.Filters) ([]*Group, filters.Metadata, error)
	Delete(ctx context.Context, id uuid.UUID) error
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	IsMember(ctx context.Context, groupID, userID uuid.UUID) (*membership.Membership, error)
}

func (s *GroupService) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*GroupDetail, error) {
	key := s.kb.BuildItemKey(id.String())
	g, err := cache.FetchOrCache(ctx, s.cache, key, func() (*Group, error) {
		return s.repo.FindByID(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	membersPtr, err := s.membershipService.FindByGroupID(ctx, g.ID)
	if err != nil {
		return nil, err
	}

	members := make([]membership.Membership, len(membersPtr))
	for i, m := range membersPtr {
		members[i] = *m
	}

	return &GroupDetail{Group: *g, Members: members}, nil

}

func (s *GroupService) ListByUser(
	ctx context.Context,
	f filters.Filters,
) ([]*Group, filters.Metadata, error) {
	key := s.kb.BuildListKey(
		cache.Int(f.Page),
		cache.Int(f.PageSize),
		cache.Str(f.Sort),
	)

	type listPayload struct {
		Groups   []*Group
		Metadata filters.Metadata
	}

	payload, err := cache.FetchOrCache(ctx, s.cache, key, func() (listPayload, error) {
		groups, meta, err := s.repo.FindAllByUser(ctx, f)
		if err != nil {
			return listPayload{}, err
		}
		return listPayload{Groups: groups, Metadata: meta}, nil
	})
	if err != nil {
		return nil, filters.Metadata{}, err
	}

	return payload.Groups, payload.Metadata, nil
}

func (s *GroupService) Create(
	ctx context.Context,
	dto CreateGroupDTO,
) (*Group, error) {
	v := validator.New()
	dto.Validate(v)
	if !v.Valid() {
		return nil, apierror.NewValidationError(v.Errors)
	}

	userID := contexts.GetUser(ctx).GetID()

	g := &Group{
		Name:    dto.Name,
		OwnerID: userID,
	}

	err := s.we.Execute(ctx, func(ctx context.Context) error {
		if err := s.repo.Insert(ctx, g); err != nil {
			return err
		}

		return s.membershipService.Insert(ctx, &membership.Membership{
			GroupID: g.ID,
			UserID:  userID,
			Role:    membership.RoleOwner,
		})
	})

	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *GroupService) Delete(ctx context.Context, id uuid.UUID) error {
	userID := contexts.GetUser(ctx).GetID()
	g, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if g.OwnerID != userID {
		return apierror.NewBadRequestError(fmt.Errorf("You are not owner that group"))
	}

	return s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.SoftDelete(ctx, id)
	})
}

func (s *GroupService) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	actorID := contexts.GetUser(ctx).GetID()

	g, err := s.repo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}
	if g.OwnerID != actorID {
		return apierror.NewBadRequestError(fmt.Errorf("You are not owner that group"))
	}
	if userID == actorID {
		return apierror.NewValidationError(map[string]string{
			"user_id": "owner cannot remove themselves",
		})
	}

	return s.membershipService.Remove(ctx, groupID, userID)
}

func (s *GroupService) IsMember(ctx context.Context, groupID, userID uuid.UUID) (*membership.Membership, error) {
	return s.membershipService.FindByGroupIDAndUserID(ctx, groupID, userID)
}
