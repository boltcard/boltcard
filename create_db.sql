SELECT 'CREATE DATABASE card_db' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'card_db');

\c card_db;

CREATE TABLE cards (
	card_id INT GENERATED ALWAYS AS IDENTITY,
	k0_auth_key CHAR(32) NOT NULL,
	k2_cmac_key CHAR(32) NOT NULL,
	k3 CHAR(32) NOT NULL,
	k4 CHAR(32) NOT NULL,
	uid CHAR(14) NOT NULL,
	last_counter_value INTEGER NOT NULL,
	lnurlw_request_timeout_sec INT NOT NULL,
	enable_flag CHAR(1) NOT NULL DEFAULT 'N',
	tx_limit_sats INT NOT NULL,
	day_limit_sats INT NOT NULL,
	card_name VARCHAR(100) NOT NULL DEFAULT '',
	one_time_code CHAR(32) NOT NULL DEFAULT '',
	one_time_code_expiry TIMESTAMPTZ DEFAULT NOW() + INTERVAL '1 DAY',
	one_time_code_used CHAR(1) NOT NULL DEFAULT 'Y',
	PRIMARY KEY(card_id)
);

CREATE TABLE card_payments (
	card_payment_id INT GENERATED ALWAYS AS IDENTITY,
	card_id INT NOT NULL,
	lnurlw_k1 CHAR(32) UNIQUE NOT NULL,
	lnurlw_request_time TIMESTAMPTZ NOT NULL,
	ln_invoice VARCHAR(1024) NOT NULL DEFAULT '',
	amount_msats BIGINT CHECK (amount_msats > 0),
	paid_flag CHAR(1) NOT NULL,
	payment_time TIMESTAMPTZ,
	payment_status VARCHAR(100) NOT NULL DEFAULT '',
	failure_reason VARCHAR(100) NOT NULL DEFAULT '',
	payment_status_time TIMESTAMPTZ,
	PRIMARY KEY(card_payment_id),
	CONSTRAINT fk_card FOREIGN KEY(card_id) REFERENCES cards(card_id)
);

GRANT ALL PRIVILEGES ON TABLE cards TO cardapp;
GRANT ALL PRIVILEGES ON TABLE card_payments TO cardapp;
