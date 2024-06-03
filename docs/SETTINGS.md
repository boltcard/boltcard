# Settings

The database connection settings are in the system environment variables.  
Other settings are in the database in a `settings` table. 

Here are the descriptions of values available to use in the `settings` table:

| Name | Value | Description |
| --- | --- | --- |
| LOG_LEVEL | DEBUG | system logs are verbose to enable easier debug |
| | PRODUCTION | system logs are minimal |
| AES_DECRYPT_KEY | | hex encoded 128 bit AES key - see [FAQ](FAQ.md#how-do-i-generate-a-random-key-value-)|
| HOST_DOMAIN | yourdomain.com | the domain for hosting lnurlw & lnurlp services |
| MIN_WITHDRAW_SATS | 1 | minimum satoshis for lnurlw response |
| MAX_WITHDRAW_SATS | 1000000 | maximum satoshis for lnurlw response |
| LN_HOST | your_lnd_node.io | LND node gRPC domain |
| LN_PORT | 9001 | LND node gRPC port |
| LN_TLS_FILE | /home/ubuntu/boltcard/tls.cert | absolute path to your LND TLC certificate |
| LN_MACAROON_FILE | /home/ubuntu/boltcard/boltcard.macaroon | absolute path to your LND macaroon |
| FEE_LIMIT_SAT | 10 | the base fee limit amount for every invoice payment |
| FEE_LIMIT_PERCENT | 0.5 | the percentage fee limit amount added to the base fee limit amount |
| LN_TESTNODE | | lightning node pubkey for allowing only the defined test node |
| FUNCTION_LNURLW | ENABLE | system level switch for LNURLw (bolt card) services |
| FUNCTION_LNURLP | DISABLE | system level switch for LNURLp (lightning address) services |
| FUNCTION_EMAIL | DISABLE | system level switch for email updates on credits & debits |
| DEFAULT_DESCRIPTION | 'bolt card service' | default description of payment |
| AWS_SES_ID | | Amazon Web Services - Simple Email Service - access id |
| AWS_SES_SECRET | | Amazon Web Services - Simple Email Service - access secret |
| AWS_SES_EMAIL_FROM | | Amazon Web Services - Simple Email Service - email from field |
| AWS_REGION | us-east-1 | Amazon Web Services - Account region |
| EMAIL_MAX_TXS | | maximum number of transactions to include in the email body |
| FUNCTION_LNDHUB | DISABLE | system level switch for using LNDHUB in place of LND |
| LNDHUB_URL | | URL for the LNDHUB service |
| FUNCTION_INTERNAL_API | DISABLE | system level switch for activating the internal API |
| SENDGRID_API_KEY      | | User API Key from SendGrid.com             |
| SENDGRID_EMAIL_SENDER | | Single Sender email address verified by SendGrid |
| LN_INVOICE_EXPIRY_SEC | 3600 | LN invoice's expiry time in seconds |
