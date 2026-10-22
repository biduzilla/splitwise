CREATE SCHEMA IF NOT EXISTS grp;
SET search_path TO grp;

CREATE TABLE IF NOT EXISTS groups (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    currency   CHAR(3) NOT NULL DEFAULT 'BRL',
    owner_id   UUID NOT NULL,
    version    INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_at TIMESTAMPTZ,
    updated_by UUID,
    deleted    BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_groups_owner ON groups(owner_id) WHERE deleted = false;

CREATE TABLE IF NOT EXISTS memberships (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id   UUID NOT NULL REFERENCES groups(id),
    user_id    UUID NOT NULL,
    role       TEXT NOT NULL DEFAULT 'member',
    version    INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_at TIMESTAMPTZ,
    updated_by UUID,
    deleted    BOOLEAN NOT NULL DEFAULT false,
    UNIQUE (group_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_memberships_user  ON memberships(user_id)  WHERE deleted = false;
CREATE INDEX IF NOT EXISTS idx_memberships_group ON memberships(group_id) WHERE deleted = false;

CREATE TABLE IF NOT EXISTS invitations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id   UUID NOT NULL REFERENCES groups(id),
    inviter_id UUID NOT NULL,
    jti        UUID UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    used_by    UUID,
    version    INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by UUID,
    updated_at TIMESTAMPTZ,
    updated_by UUID,
    deleted    BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_invitations_jti ON invitations(jti);