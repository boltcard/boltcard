\c card_db

INSERT INTO cards (
		aes_cmac,
		uid,
		last_counter_value,
		lnurlw_request_timeout_sec,
		enable_flag,
		tx_limit_sats,
		day_limit_sats,
		card_description
)
	VALUES (
		'00000000000000000000000000000000',
		'00000000000000',
		0,
		60,
		'Y',
		1000,
		10000,
		'bolt card'
	);
