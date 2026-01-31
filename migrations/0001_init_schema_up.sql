CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    lastname VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_clients_name_lastname ON clients (name, lastname);

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    balance bigint NOT NULL DEFAULT 0,
    CONSTRAINT balance_non_negative CHECK (balance >= 0),
    client_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (client_id) REFERENCES clients(id)
);
CREATE INDEX idx_clientid ON accounts(client_id);

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

