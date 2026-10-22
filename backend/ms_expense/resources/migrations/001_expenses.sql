-- 001_expenses.sql
CREATE SCHEMA IF NOT EXISTS exp;

CREATE TABLE exp.expenses (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id      UUID NOT NULL,
    paid_by       UUID NOT NULL,
    amount        BIGINT NOT NULL CHECK (amount > 0),
    description   TEXT NOT NULL,
    category      TEXT,
    split_mode    TEXT NOT NULL CHECK (split_mode IN ('equal', 'exact')),
    expense_date  DATE NOT NULL DEFAULT CURRENT_DATE,
    version       INTEGER NOT NULL DEFAULT 1,
    created_by    UUID NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ,
    updated_by    UUID,
    deleted       BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_expenses_group ON exp.expenses(group_id, expense_date DESC)
    WHERE deleted = false;
CREATE INDEX idx_expenses_created_by ON exp.expenses(created_by)
    WHERE deleted = false;

CREATE TABLE exp.expense_shares (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id  UUID NOT NULL REFERENCES exp.expenses(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL,
    amount      BIGINT NOT NULL CHECK (amount >= 0),
    UNIQUE (expense_id, user_id)
);

CREATE INDEX idx_shares_expense ON exp.expense_shares(expense_id);
CREATE INDEX idx_shares_user ON exp.expense_shares(user_id);

-- 002_idempotency.sql
CREATE TABLE exp.idempotency_keys (
    key             UUID PRIMARY KEY,
    user_id         UUID NOT NULL,
    endpoint        TEXT NOT NULL,
    request_hash    TEXT NOT NULL,   -- opcional: detecta mesmo key com body diferente
    response_status INT NOT NULL,
    response_body   JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_exp_idempotency_ttl ON exp.idempotency_keys(created_at);