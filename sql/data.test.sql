-- connect to card_db
\c card_db;

-- clear out table data
DELETE FROM settings;
DELETE FROM card_payments;
DELETE FROM card_receipts;
DELETE FROM cards;

-- set up test data
INSERT INTO settings (name, value) VALUES ('LOG_LEVEL', 'DEBUG');
INSERT INTO settings (name, value) VALUES ('AES_DECRYPT_KEY', '994de7f8156609a0effafbdb049337b1');
INSERT INTO settings (name, value) VALUES ('HOST_DOMAIN', 'localhost:9000');
INSERT INTO settings (name, value) VALUES ('FUNCTION_INTERNAL_API', 'ENABLE');
INSERT INTO settings (name, value) VALUES ('MIN_WITHDRAW_SATS', '1');
INSERT INTO settings (name, value) VALUES ('MAX_WITHDRAW_SATS', '1000');


INSERT INTO cards 
	(k0_auth_key, k2_cmac_key, k3, k4, lnurlw_enable, last_counter_value, lnurlw_request_timeout_sec,
	tx_limit_sats, day_limit_sats, card_name, pin_limit_sats) 
	VALUES 
	('', 'd3dffa1e12d2477e443a6ee9fcfeab18', '', '', 'Y', 0, 10,
	0, 0, 'test_card', 0);
