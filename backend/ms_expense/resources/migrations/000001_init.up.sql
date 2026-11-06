CREATE TABLE IF NOT EXISTS expenses (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id      UUID NOT NULL,
    paid_by       UUID NOT NULL,
    amount        BIGINT NOT NULL CHECK (amount > 0), 
    description   TEXT NOT NULL,
    category      TEXT,
    split_mode    TEXT NOT NULL DEFAULT 'equal'
                  CHECK (split_mode IN ('equal', 'exact')),
    expense_date  DATE NOT NULL DEFAULT CURRENT_DATE,
    version       INTEGER NOT NULL DEFAULT 1,
    created_by    UUID NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ,
    updated_by    UUID,
    deleted       BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_expenses_group
    ON exp.expenses(group_id, expense_date DESC)
    WHERE deleted = false;

CREATE INDEX idx_expenses_paid_by
    ON exp.expenses(paid_by)
    WHERE deleted = false;

CREATE TABLE IF NOT EXISTS expense_shares (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_id  UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL,
    amount      BIGINT NOT NULL CHECK (amount >= 0),
    UNIQUE (expense_id, user_id)
);

CREATE INDEX idx_shares_expense ON exp.expense_shares(expense_id);
CREATE INDEX idx_shares_user    ON exp.expense_shares(user_id);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key             UUID PRIMARY KEY,
    user_id         UUID NOT NULL,
    endpoint        TEXT NOT NULL,
    response_status INTEGER NOT NULL,
    response_body   JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_exp_idempotency_ttl
    ON exp.idempotency_keys(created_at);