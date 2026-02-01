-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    balance bigint NOT NULL DEFAULT 0,
    CONSTRAINT balance_non_negative CHECK (balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL,
    amount bigint NOT NULL,
    type VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (account_id) REFERENCES accounts(id),

    CONSTRAINT valid_type CHECK (type IN ('credit', 'debit')),
    CONSTRAINT valid_amount CHECk (amount > 0)
);
CREATE INDEX idx_transactions_account_id ON transactions(account_id);

CREATE TABLE IF NOT EXISTS outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW() 
);

CREATE INDEX idx_outbox_created_at ON outbox(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_transactions_account_id;
DROP TABLE IF EXISTS transactions;

DROP TABLE IF EXISTS accounts;


DROP INDEX IF EXISTS idx_outbox_created_at;
DROP TABLE IF EXISTS outbox;
-- +goose StatementEnd