CREATE TABLE IF NOT EXISTS grp_groups (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    currency    CHAR(3) NOT NULL DEFAULT 'BRL',
    owner_id    UUID NOT NULL,
    version     INTEGER NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by  UUID,
    updated_at  TIMESTAMPTZ,
    updated_by  UUID,
    deleted     BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_grp_groups_owner
    ON grp_groups(owner_id) WHERE deleted = false;

CREATE TABLE IF NOT EXISTS grp_memberships (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID NOT NULL REFERENCES grp_groups(id),
    user_id     UUID NOT NULL,
    role        TEXT NOT NULL DEFAULT 'member',
    version     INTEGER NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by  UUID,
    updated_at  TIMESTAMPTZ,
    updated_by  UUID,
    deleted     BOOLEAN NOT NULL DEFAULT false,
    UNIQUE (group_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_grp_memberships_user
    ON grp_memberships(user_id) WHERE deleted = false;
CREATE INDEX IF NOT EXISTS idx_grp_memberships_group
    ON grp_memberships(group_id) WHERE deleted = false;

CREATE TABLE IF NOT EXISTS grp_invitations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID NOT NULL REFERENCES grp_groups(id),
    inviter_id  UUID NOT NULL,
    jti         UUID UNIQUE NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    used_by     UUID,
    version     INTEGER NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by  UUID,
    updated_at  TIMESTAMPTZ,
    updated_by  UUID,
    deleted     BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_grp_invitations_jti
    ON grp_invitations(jti);