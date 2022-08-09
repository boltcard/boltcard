\c card_db

INSERT INTO cards (
		lock_key,	/* this is key 0 on the card */
		aes_cmac,	/* this is key 2 on the card */
		uid,		/* this can be discovered from the service log */
		last_counter_value,	/* can start at zero and will be updated on first use (before issue) */
		lnurlw_request_timeout_sec, /* 60 seconds by default */
		enable_flag,		/* useful for quickly switching card hosting on/off */
		tx_limit_sats,		/* set at a reasonable value for small test payments in 2022 */
		day_limit_sats,		/* set at a reasonable value for small test payments in 2022 */
		card_description,	/* to store a human readable card description (optional) */
                one_time_code_used	/* used to indicate if the one_time_code for card creation is live */
)
	VALUES (
		'00000000000000000000000000000000',
		'00000000000000000000000000000000',
		'00000000000000',
		0,
		60,
		'Y',
		1000,
		10000,
		'bolt card',
		'Y'
	);
