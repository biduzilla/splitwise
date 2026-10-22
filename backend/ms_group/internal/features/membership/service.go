package membership

import (
	"context"
	"fmt"
	"log/slog"
	"messaging"
	"ms_group/internal/core/httpclient/auth"
	"shared/apierror"
	"shared/auth/contexts"
	"shared/cache"
	"shared/filters"
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
	producer    producer
	groupFinder groupFinder
	logger      *slog.Logger
	authClient  authClient
}

type producer interface {
	Publish(ctx context.Context, topic messaging.Topic, event messaging.Event) error
}

type authClient interface {
	FindAllByIds(
		ctx context.Context,
		ids []uuid.UUID,
		page int,
		pageSize int,
	) ([]auth.UserDTO, filters.Metadata, error)
}

func NewService(
	repo repository,
	cache cache.Cache,
	we WriteExecutor,
	kb cache.KeyBuilder,
	producer producer,
	logger *slog.Logger,
	groupFinder groupFinder,
	authClient authClient,
) *MembershipService {
	return &MembershipService{
		repo:        repo,
		cache:       cache,
		we:          we,
		kb:          kb,
		groupFinder: groupFinder,
		producer:    producer,
		logger:      logger,
		authClient:  authClient,
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
	members, err := cache.FetchOrCache(ctx, s.cache, key, func() ([]*Membership, error) {
		return s.repo.FindByGroupID(ctx, groupID)
	})

	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(members))
	for i, m := range members {
		ids[i] = m.UserID
	}
	users, _, err := s.authClient.FindAllByIds(ctx, ids, 0, 100)
	if err != nil {
		return nil, err
	}

	userMap := make(map[uuid.UUID]auth.UserDTO, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}

	for _, m := range members {
		if u, ok := userMap[m.UserID]; ok {
			m.User = &u
		}
	}

	return members, nil
}

func (s *MembershipService) FindByGroupIDAndUserID(ctx context.Context, groupID, userID uuid.UUID) (*Membership, error) {
	key := s.kb.BuildListKey(
		cache.UUID(groupID),
		cache.UUID(userID),
	)

	m, err := cache.FetchOrCache(ctx, s.cache, key, func() (*Membership, error) {
		return s.repo.FindByGroupIDAndUserID(ctx, groupID, userID)
	})
	if err != nil {
		return nil, err
	}

	user, _, err := s.authClient.FindAllByIds(ctx, []uuid.UUID{m.UserID}, 0, 100)
	if err != nil {
		return nil, err
	}
	m.User = &user[0]

	return m, nil
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

	// var member *Membership
	err = s.we.Execute(ctx, func(ctx context.Context) error {
		// m, err := s.repo.FindByGroupIDAndUserID(ctx, groupID, userID)
		// if err != nil {
		// 	return err
		// }
		// member = m
		return s.repo.Remove(ctx, groupID, userID)
	})

	if err != nil {
		return err
	}

	// event := GroupMemberRemoved{
	// 	EventID:      uuid.New(),
	// 	EventType:    messaging.EventTypeGroupMemberRemoved.String(),
	// 	GroupID:      groupID,
	// 	MembershipID: member.ID,
	// 	UserID:       userID,
	// 	RemovedBy:    actorID,
	// 	OccurredAt:   time.Now().UTC(),
	// }

	// if err := s.producer.Publish(ctx, messaging.RateioEvents, event); err != nil {
	// 	s.logger.Error("failed to publish group.member_removed",
	// 		"group_id", groupID,
	// 		"user_id", userID,
	// 		"error", err,
	// 	)
	// }

	return nil
}

func (s *MembershipService) SetGroupFinder(gf groupFinder) {
	s.groupFinder = gf
}
