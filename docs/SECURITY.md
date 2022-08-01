# Security

## secrets
- card AES decrypt key to the environment variable `AES_DECRYPT_KEY`
- card AES cmac keys into the database table `cards`

- `tls.cert` and `SendPaymentV2.macaroon` for the lightning node

- password for the application database user `cardapp`
  - database script in `create_db.sql`
  - application environment variable in `lnurlw.service`
