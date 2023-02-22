\c card_db;

CREATE TABLE settings (
	setting_id INT GENERATED ALWAYS AS IDENTITY,
	name VARCHAR(30) UNIQUE NOT NULL DEFAULT '',
	value VARCHAR(128) NOT NULL DEFAULT '',
	PRIMARY KEY(setting_id)
);

CREATE TABLE cards (
	card_id INT GENERATED ALWAYS AS IDENTITY,
	k0_auth_key CHAR(32) NOT NULL,
	k2_cmac_key CHAR(32) NOT NULL,
	k3 CHAR(32) NOT NULL,
	k4 CHAR(32) NOT NULL,
	uid VARCHAR(14) NOT NULL DEFAULT '',
	last_counter_value INTEGER NOT NULL,
	lnurlw_request_timeout_sec INT NOT NULL,
	lnurlw_enable CHAR(1) NOT NULL DEFAULT 'N',
	tx_limit_sats INT NOT NULL,
	day_limit_sats INT NOT NULL,
	lnurlp_enable CHAR(1) NOT NULL DEFAULT 'N',
	card_name VARCHAR(100) UNIQUE NOT NULL DEFAULT '',
	email_address VARCHAR(100) DEFAULT '',
	email_enable CHAR(1) NOT NULL DEFAULT 'N',
	uid_privacy CHAR(1) NOT NULL DEFAULT 'N',
	one_time_code CHAR(32) NOT NULL DEFAULT '',
	one_time_code_expiry TIMESTAMPTZ DEFAULT NOW() + INTERVAL '1 DAY',
	one_time_code_used CHAR(1) NOT NULL DEFAULT 'Y',
	allow_negative_balance CHAR(1) NOT NULL DEFAULT 'N',
	wiped CHAR(1) NOT NULL DEFAULT 'N',
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

CREATE TABLE card_receipts (
	card_receipt_id INT GENERATED ALWAYS AS IDENTITY,
	card_id INT NOT NULL,
	ln_invoice VARCHAR(1024) NOT NULL DEFAULT '',
	r_hash_hex CHAR(64) UNIQUE NOT NULL DEFAULT '',
	amount_msats BIGINT CHECK (amount_msats > 0),
	receipt_status VARCHAR(100) NOT NULL DEFAULT '',
	receipt_status_time TIMESTAMPTZ,
	PRIMARY KEY(card_receipt_id),
	CONSTRAINT fk_card FOREIGN KEY(card_id) REFERENCES cards(card_id)
);


GRANT ALL PRIVILEGES ON TABLE settings TO cardapp;
GRANT ALL PRIVILEGES ON TABLE cards TO cardapp;
GRANT ALL PRIVILEGES ON TABLE card_payments TO cardapp;
GRANT ALL PRIVILEGES ON TABLE card_receipts TO cardapp;

