package api

import (
	"log/slog"

	"ms_group/internal/core/config"
	"ms_group/internal/features/group"
	"ms_group/internal/features/invitation"
	"ms_group/internal/features/membership"
	"shared/auth/security"
	"shared/cache"
	"shared/transaction"
)

type services struct {
	jwtService *security.JwtService
	group      *group.GroupService
	invitation *invitation.InvitationService
	membership *membership.MembershipService
}

func NewServices(
	r *repositories,
	tx transaction.Manager,
	config config.Config,
	logger *slog.Logger,
) (*services, error) {
	cacheClient, err := cache.NewRedisCache(config.Base.Cache.Addr,
		config.Base.Cache.Password, config.Base.Cache.Db, nil)

	if err != nil {
		return nil, err
	}

	logger.Info("reddis connection pool established")

	jwtService, err := security.NewService(config.Base)
	if err != nil {
		return nil, err
	}

	mbD := newWriteDeps("membership", tx, cacheClient)
	mbService := membership.NewService(
		r.membership,
		mbD.cache,
		mbD.executor,
		mbD.keyBuilder,
		nil,
	)

	grpDeps := newWriteDeps("group", tx, cacheClient)
	groupSvc := group.NewService(
		r.group,
		grpDeps.cache,
		grpDeps.executor,
		grpDeps.keyBuilder,
		mbService,
	)

	mbService.SetGroupFinder(groupSvc)

	invDeps := newWriteDeps("invitation", tx, cacheClient)

	tokenSvc, err := invitation.NewTokenService(
		config.Base.Security.PrivateKeyPath,
		config.Base.Security.PublicKeyPath,
	)
	if err != nil {
		return nil, err
	}

	invitationSvc := invitation.NewService(
		r.invitation,
		invDeps.cache,
		invDeps.executor,
		invDeps.keyBuilder,
		mbService,
		tokenSvc,
		config.InviteURLBase,
	)

	return &services{
		jwtService: jwtService,
		group:      groupSvc,
		invitation: invitationSvc,
		membership: mbService,
	}, nil
}

type writeDeps struct {
	cache      cache.Cache
	keyBuilder cache.KeyBuilder
	executor   *transaction.WriteExecutor
}

func newWriteDeps(prefix string, tx transaction.Manager, cacheClient cache.Cache) writeDeps {
	kb := cache.NewKeyBuilder(prefix)
	we := transaction.NewWriterExecutor(tx, cacheClient, kb)
	return writeDeps{
		cache:      cacheClient,
		keyBuilder: kb,
		executor:   we,
	}
}
