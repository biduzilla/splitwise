CREATE TABLE IF NOT EXISTS group_balances (
    group_id    UUID NOT NULL,
    user_id     UUID NOT NULL,
    balance     BIGINT NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE IF NOT EXISTS settlements (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id      UUID NOT NULL,
    from_user_id  UUID NOT NULL,
    to_user_id    UUID NOT NULL,
    amount        BIGINT NOT NULL CHECK (amount > 0),
    settled_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    version       INTEGER NOT NULL DEFAULT 1,
    created_by    UUID NOT NULL,
    deleted       BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_settlements_group
    ON bal.settlements(group_id, settled_at DESC)
    WHERE deleted = false;

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key             UUID PRIMARY KEY,
    user_id         UUID NOT NULL,
    endpoint        TEXT NOT NULL,
    response_status INTEGER NOT NULL,
    response_body   JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);