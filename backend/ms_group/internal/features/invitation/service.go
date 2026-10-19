package invitation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"messaging"
	"ms_group/internal/features/membership"
	"shared/apierror"
	"shared/auth/contexts"
	"shared/cache"
	"time"
	"uuid"
)

const maxGroupMembers = 10

type InvitationService struct {
	repo              repository
	cache             cache.Cache
	we                WriteExecutor
	kb                cache.KeyBuilder
	producer          producer
	logger            *slog.Logger
	membershipService membershipService
	tokenService      tokenService
	inviteURLBase     string
}

type producer interface {
	Publish(ctx context.Context, topic messaging.Topic, event messaging.Event) error
}

type WriteExecutor interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}

type tokenService interface {
	Generate(
		groupID, inviterID, jti uuid.UUID,
	) (string, time.Time, error)
	Validate(raw string) (*Claims, error)
}

type membershipService interface {
	Insert(ctx context.Context, m *membership.Membership) error
	FindByGroupIDAndUserID(ctx context.Context, groupID, userID uuid.UUID) (*membership.Membership, error)
	CountActiveMembers(ctx context.Context, groupID uuid.UUID) (int, error)
}

func NewService(
	repo repository,
	cache cache.Cache,
	we WriteExecutor,
	kb cache.KeyBuilder,
	producer producer,
	logger *slog.Logger,
	membershipService membershipService,
	tokenService tokenService,
	inviteURLBase string,
) *InvitationService {
	return &InvitationService{
		repo:              repo,
		cache:             cache,
		we:                we,
		kb:                kb,
		producer:          producer,
		membershipService: membershipService,
		tokenService:      tokenService,
		inviteURLBase:     inviteURLBase,
		logger:            logger,
	}
}

type service interface {
	Create(ctx context.Context, groupID uuid.UUID) (*InvitationDTO, error)
	Accept(
		ctx context.Context,
		raw string,
	) (*membership.Membership, error)
}

func (s *InvitationService) Create(
	ctx context.Context,
	groupID uuid.UUID,
) (*InvitationDTO, error) {
	userID := contexts.GetUser(ctx).GetID()

	_, err := s.membershipService.
		FindByGroupIDAndUserID(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}

	count, err := s.membershipService.CountActiveMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if count >= maxGroupMembers {
		return nil, apierror.NewValidationError(map[string]string{
			"group": fmt.Sprintf("group already has %d members (max)", maxGroupMembers),
		})
	}

	jti := uuid.New()
	raw, exp, err := s.tokenService.Generate(
		groupID,
		userID,
		jti,
	)
	if err != nil {
		return nil, err
	}

	inv := &Invitation{
		GroupID:   groupID,
		InviterID: userID,
		JTI:       jti,
		ExpiresAt: exp,
	}

	err = s.we.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Insert(ctx, inv)
	})
	if err != nil {
		return nil, err
	}

	return &InvitationDTO{
		Token:     raw,
		URL:       fmt.Sprintf("%s/join?token=%s", s.inviteURLBase, raw),
		ExpiresAt: exp,
	}, nil
}

func (s *InvitationService) Accept(
	ctx context.Context,
	raw string,
) (*membership.Membership, error) {
	claims, err := s.tokenService.Validate(raw)
	if err != nil {
		return nil, apierror.NewHTTPError("invalid or expired invitation", 400, err)
	}

	jti, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, apierror.NewHTTPError("invalid invitation id", 400, err)
	}

	inv, err := s.repo.FindByJTI(ctx, jti)
	if err != nil {
		return nil, err
	}
	if inv.UsedAt != nil {
		return nil, apierror.NewHTTPError("invitation already used", 409, nil)
	}
	if inv.ExpiresAt.Before(time.Now()) {
		return nil, apierror.NewHTTPError("invitation expired", 400, nil)
	}

	userID := contexts.GetUser(ctx).GetID()

	existing, err := s.membershipService.FindByGroupIDAndUserID(ctx, inv.GroupID, userID)
	if err == nil && existing != nil {
		return existing, nil
	}
	if err != nil && !errors.Is(err, apierror.ErrRecordNotFound) {
		return nil, err
	}

	count, err := s.membershipService.CountActiveMembers(ctx, inv.GroupID)
	if err != nil {
		return nil, err
	}
	if count >= maxGroupMembers {
		return nil, apierror.NewValidationError(map[string]string{
			"group": "group is full",
		})
	}
	m := &membership.Membership{
		GroupID: inv.GroupID,
		UserID:  userID,
		Role:    membership.RoleMember,
	}

	err = s.we.Execute(ctx, func(ctx context.Context) error {
		if err := s.membershipService.Insert(ctx, m); err != nil {
			return err
		}
		return s.repo.MarkUsed(ctx, inv.ID, userID)
	})

	event := GroupMemberAdded{
		EventID:      uuid.New(),
		EventType:    messaging.EventTypeGroupMemberAdded.String(),
		GroupID:      m.GroupID,
		MembershipID: m.ID,
		UserID:       m.UserID,
		Role:         string(m.Role),
		InviterID:    claims.InviterID,
		OccurredAt:   time.Now().UTC(),
	}

	if err := s.producer.Publish(ctx, messaging.RateioEvents, event); err != nil {
		s.logger.Error("failed to publish group.member_added",
			"group_id", m.GroupID,
			"user_id", m.UserID,
			"error", err,
		)
	}

	return m, err
}
