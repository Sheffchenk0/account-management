DROP INDEX IF EXISTS idx_transactions_account_id;
DROP TABLE IF EXISTS transactions;

DROP INDEX IF EXISTS idx_clientid;
DROP INDEX IF EXISTS idx_clients_name_lastname;
DROP TABLE IF EXISTS accounts;

DROP TABLE IF EXISTS clients;

DROP INDEX IF EXISTS idx_outbox_created_at;
DROP TABLE IF EXISTS outbox;

